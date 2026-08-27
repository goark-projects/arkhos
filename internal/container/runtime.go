package container

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"sync"

	"goark.dev/arkhos/internal/deploy"

	servletcontainer "goark.dev/arkarta/servlet/container"
)

// Runtime 管理 Arkhos 容器的部署快照与生命周期状态。
type Runtime struct {
	metadata  servletcontainer.Metadata
	validator deploy.Validator
	decorator ApplicationDecorator

	mu           sync.RWMutex
	applications []servletcontainer.Application
	started      bool
	shutdown     bool
}

// ErrNilDecoratedApplication 表示应用装饰器返回了空应用。
var ErrNilDecoratedApplication = errors.New("arkhos/container: application decorator returned nil")

// ApplicationDecorator 在应用初始化后包装容器运行能力。
type ApplicationDecorator func(servletcontainer.Application) (servletcontainer.Application, error)

// Option 定制容器运行时。
type Option func(*Runtime)

// WithApplicationDecorator 设置应用装饰器。
func WithApplicationDecorator(decorator ApplicationDecorator) Option {
	return func(r *Runtime) {
		r.decorator = decorator
	}
}

// NewRuntime 创建容器运行时。
func NewRuntime(metadata servletcontainer.Metadata, options ...Option) *Runtime {
	runtime := &Runtime{
		metadata:  metadata,
		validator: deploy.NewValidator(metadata),
	}
	for _, option := range options {
		if option != nil {
			option(runtime)
		}
	}
	return runtime
}

// Metadata 返回容器能力元数据。
func (r *Runtime) Metadata() servletcontainer.Metadata {
	if r == nil {
		return servletcontainer.Metadata{}
	}
	return r.metadata
}

// Deploy 初始化并保存一个 Web 应用部署。
func (r *Runtime) Deploy(ctx context.Context, deployment *servletcontainer.Deployment) (servletcontainer.Application, error) {
	if r == nil {
		return nil, http.ErrServerClosed
	}
	ctx = normalizeContext(ctx)
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := r.validator.Validate(deployment); err != nil {
		return nil, err
	}
	managed, err := servletcontainer.NewApplication(ctx, deployment)
	if err != nil {
		return nil, err
	}
	var application servletcontainer.Application = managed
	if r.decorator != nil {
		decorated, err := r.decorator(application)
		if err != nil {
			return nil, errors.Join(err, managed.Stop(ctx))
		}
		if decorated == nil {
			return nil, errors.Join(ErrNilDecoratedApplication, managed.Stop(ctx))
		}
		application = decorated
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	if r.shutdown {
		_ = application.Stop(ctx)
		return nil, http.ErrServerClosed
	}
	r.applications = append(r.applications, application)
	return application, nil
}

// Start 标记运行时可对外处理请求。
func (r *Runtime) Start(ctx context.Context) error {
	if r == nil {
		return http.ErrServerClosed
	}
	ctx = normalizeContext(ctx)
	if err := ctx.Err(); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	if r.shutdown {
		return http.ErrServerClosed
	}
	r.started = true
	return nil
}

// Shutdown 按部署逆序停止所有应用。
func (r *Runtime) Shutdown(ctx context.Context) error {
	if r == nil {
		return nil
	}
	ctx = normalizeContext(ctx)

	r.mu.Lock()
	if r.shutdown {
		r.mu.Unlock()
		return nil
	}
	r.shutdown = true
	r.started = false
	applications := append([]servletcontainer.Application(nil), r.applications...)
	r.mu.Unlock()

	var result error
	for i := len(applications) - 1; i >= 0; i-- {
		result = errors.Join(result, applications[i].Stop(ctx))
	}
	return result
}

// MatchApplication 按最长上下文路径匹配应用。
func (r *Runtime) MatchApplication(path string) (servletcontainer.Application, MatchStatus) {
	if r == nil {
		return nil, MatchUnavailable
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	if !r.started || r.shutdown {
		return nil, MatchUnavailable
	}

	var (
		best     servletcontainer.Application
		bestSize = -1
	)
	for _, application := range r.applications {
		webApp := application.WebApp()
		if webApp == nil {
			continue
		}
		contextPath := webApp.ContextPath()
		if matchContextPath(path, contextPath) && len(contextPath) > bestSize {
			best = application
			bestSize = len(contextPath)
		}
	}
	if best == nil {
		return nil, MatchNotFound
	}
	return best, MatchFound
}

func normalizeContext(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}

func matchContextPath(path, contextPath string) bool {
	if path == "" {
		path = "/"
	}
	if contextPath == "" || contextPath == "/" {
		return true
	}
	return path == contextPath || strings.HasPrefix(path, contextPath+"/")
}
