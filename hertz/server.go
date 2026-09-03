package hertz

import (
	"context"
	"errors"
	"net"
	"sync"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	hertzserver "github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/config"
	"github.com/cloudwego/hertz/pkg/network"
	"github.com/cloudwego/hertz/pkg/network/standard"
	"github.com/cloudwego/hertz/pkg/protocol/consts"

	internaltransport "goark.dev/arkhos/hertz/internal/transport"
)

// Server 将 Arkhos 容器绑定到 Hertz HTTP Server。
type Server struct {
	container *Container
	options   serverOptions
	tracker   *internaltransport.Tracker

	mu     sync.RWMutex
	engine *hertzserver.Hertz
}

// NewServer 创建可监听网络连接的 Arkhos Hertz Server。
func NewServer(container *Container, options ...ServerOption) (*Server, error) {
	if container == nil {
		return nil, ErrNilContainer
	}
	return &Server{
		container: container,
		options:   buildServerOptions(options),
		tracker:   internaltransport.NewTracker(),
	}, nil
}

// Handler 返回容器聚合后的 Hertz Handler。
func (s *Server) Handler() app.HandlerFunc {
	if s == nil || s.container == nil {
		return func(_ context.Context, requestContext *app.RequestContext) {
			requestContext.SetStatusCode(consts.StatusServiceUnavailable)
		}
	}
	return s.container.Handler()
}

// Address 返回配置的监听地址。
func (s *Server) Address() string {
	if s == nil {
		return ""
	}
	return s.options.address
}

// Hertz 返回当前正在运行的底层 Hertz 实例。
func (s *Server) Hertz() *hertzserver.Hertz {
	if s == nil {
		return nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.engine
}

// Running 返回底层 Hertz Engine 是否已进入可服务状态。
func (s *Server) Running() bool {
	engine := s.Hertz()
	return engine != nil && engine.IsRunning()
}

// ListenAndServe 启动默认监听，并在 ctx 取消时优雅关闭。
func (s *Server) ListenAndServe(ctx context.Context) error {
	return s.serve(ctx, nil)
}

// Serve 使用调用方提供的监听器服务请求，并在 ctx 取消时优雅关闭。
func (s *Server) Serve(ctx context.Context, listener net.Listener) error {
	if listener == nil {
		return ErrNilListener
	}
	return s.serve(ctx, listener)
}

// Shutdown 优雅关闭 Hertz Server 并停止全部已部署应用。
func (s *Server) Shutdown(ctx context.Context) error {
	if s == nil || s.container == nil {
		return ErrNilContainer
	}
	ctx = normalizeServerContext(ctx)
	engine := s.Hertz()
	var engineErr error
	if engine != nil && engine.IsRunning() {
		s.tracker.BeginShutdown()
		engineErr = engine.Shutdown(ctx)
	}
	return errors.Join(engineErr, s.container.Shutdown(ctx))
}

// Close 立即关闭 Hertz Server 的监听器和活动连接，并停止全部已部署应用。
func (s *Server) Close() error {
	if s == nil || s.container == nil {
		return ErrNilContainer
	}
	engine := s.Hertz()
	var engineErr error
	if engine != nil {
		engineErr = engine.Close()
		if errors.Is(engineErr, context.DeadlineExceeded) {
			engineErr = nil
		}
	}
	return errors.Join(s.tracker.Close(), engineErr, s.container.Shutdown(context.Background()))
}

func (s *Server) serve(ctx context.Context, listener net.Listener) error {
	if s == nil || s.container == nil {
		return ErrNilContainer
	}
	ctx = normalizeServerContext(ctx)
	if err := s.container.Start(ctx); err != nil {
		return err
	}
	engine := s.newEngine(listener)
	s.mu.Lock()
	s.engine = engine
	s.mu.Unlock()

	errCh := make(chan error, 1)
	go func() {
		errCh <- engine.Run()
	}()
	running, err := waitHertzRunning(engine, errCh)
	if !running {
		return errors.Join(err, s.container.Shutdown(context.Background()))
	}
	select {
	case err := <-errCh:
		return errors.Join(err, s.container.Shutdown(context.Background()))
	case <-ctx.Done():
		shutdownErr := s.Shutdown(context.Background())
		<-errCh
		return errors.Join(ctx.Err(), shutdownErr)
	}
}

func (s *Server) newEngine(listener net.Listener) *hertzserver.Hertz {
	options := s.options.hertzOptions()
	if listener == nil {
		options = append(options, platformTransportOptions(s.tracker)...)
	} else {
		options = append(options,
			hertzserver.WithListener(listener),
			hertzserver.WithTransport(func(options *config.Options) network.Transporter {
				return s.tracker.Wrap(standard.NewTransporter(options))
			}),
		)
	}
	engine := hertzserver.New(options...)
	containerHandler := s.container.Handler()
	handler := func(ctx context.Context, requestContext *app.RequestContext) {
		conn := internaltransport.FromContext(ctx)
		if conn != nil {
			conn.BeginRequest()
			defer conn.FinishRequest()
			if s.tracker.Draining() {
				requestContext.SetConnectionClose()
			}
		}
		containerHandler(ctx, requestContext)
	}
	engine.Any("/*path", handler)
	engine.NoRoute(handler)
	return engine
}

func (o serverOptions) hertzOptions() []config.Option {
	options := []config.Option{
		hertzserver.WithHostPorts(o.address),
		hertzserver.WithDisablePreParseMultipartForm(true),
	}
	if o.readTimeoutSet {
		options = append(options, hertzserver.WithReadTimeout(o.readTimeout))
	}
	if o.writeTimeoutSet {
		options = append(options, hertzserver.WithWriteTimeout(o.writeTimeout))
	}
	if o.idleTimeoutSet {
		options = append(options, hertzserver.WithIdleTimeout(o.idleTimeout))
	}
	if o.keepAliveTimeoutSet {
		options = append(options, hertzserver.WithKeepAliveTimeout(o.keepAliveTimeout))
	}
	if o.exitWaitTimeout > 0 {
		options = append(options, hertzserver.WithExitWaitTime(o.exitWaitTimeout))
	}
	if o.maxHeaderBytes > 0 {
		options = append(options, hertzserver.WithMaxHeaderBytes(o.maxHeaderBytes))
	}
	if o.maxRequestBodySize > 0 {
		options = append(options, hertzserver.WithMaxRequestBodySize(o.maxRequestBodySize))
	}
	return options
}

func waitHertzRunning(engine *hertzserver.Hertz, errCh <-chan error) (bool, error) {
	ticker := time.NewTicker(time.Millisecond)
	defer ticker.Stop()
	for !engine.IsRunning() {
		select {
		case err := <-errCh:
			return false, err
		case <-ticker.C:
		}
	}
	return true, nil
}

func normalizeServerContext(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}
