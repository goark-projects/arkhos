package nethttp

import (
	"context"
	"net/http"

	internalcontainer "goark.dev/arkhos/internal/container"
	internalnethttp "goark.dev/arkhos/internal/nethttp"
	internalprofile "goark.dev/arkhos/internal/profile"

	"goark.dev/arkarta/servlet"
	servletcontainer "goark.dev/arkarta/servlet/container"
	"goark.dev/arkarta/servlet/nativeio"
)

// Container 是基于标准库 net/http 的 Arkhos Servlet 容器。
type Container struct {
	runtime        *internalcontainer.Runtime
	sender         nativeio.Sender
	requestOptions []servlet.RequestOption
}

// profileSecurity 与 Arkarta Servlet Security Profile 标准值保持一致。
const profileSecurity servletcontainer.Profile = "security"

// NewContainer 创建默认 Arkhos net/http 容器。
func NewContainer(options ...ContainerOption) *Container {
	cfg := buildContainerOptions(options)
	return &Container{
		runtime: internalcontainer.NewRuntime(
			defaultMetadata(),
			internalcontainer.WithApplicationDecorator(func(application servletcontainer.Application) (servletcontainer.Application, error) {
				return internalprofile.Decorate(application, cfg.profiles)
			}),
		),
		sender:         nativeio.NewStandardSender(),
		requestOptions: append([]servlet.RequestOption(nil), cfg.requestOptions...),
	}
}

// Metadata 返回容器元数据。
func (c *Container) Metadata() servletcontainer.Metadata {
	if c == nil || c.runtime == nil {
		return servletcontainer.Metadata{}
	}
	return c.runtime.Metadata()
}

// NativeSender 返回当前容器的 Native I/O 文件发送器。
func (c *Container) NativeSender() nativeio.Sender {
	if c == nil || c.sender == nil {
		return nativeio.NewStandardSender()
	}
	return c.sender
}

// Deploy 部署 Web 应用。
func (c *Container) Deploy(ctx context.Context, deployment *servletcontainer.Deployment) (servletcontainer.Application, error) {
	if c == nil || c.runtime == nil {
		return nil, http.ErrServerClosed
	}
	return c.runtime.Deploy(ctx, deployment)
}

// Start 标记容器开始服务请求。
func (c *Container) Start(ctx context.Context) error {
	if c == nil || c.runtime == nil {
		return http.ErrServerClosed
	}
	return c.runtime.Start(ctx)
}

// Shutdown 停止容器和所有已部署应用。
func (c *Container) Shutdown(ctx context.Context) error {
	if c == nil || c.runtime == nil {
		return nil
	}
	return c.runtime.Shutdown(ctx)
}

// Handler 返回容器聚合后的标准库 HTTP 处理器。
func (c *Container) Handler() http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request == nil || request.URL == nil {
			http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		if c == nil || c.runtime == nil {
			http.Error(writer, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
			return
		}
		application, status := c.runtime.MatchApplication(request.URL.Path)
		switch status {
		case internalcontainer.MatchFound:
			internalnethttp.ApplicationHandler(application, c.requestOptions...).ServeHTTP(writer, request)
		case internalcontainer.MatchUnavailable:
			http.Error(writer, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
		default:
			http.NotFound(writer, request)
		}
	})
}

func defaultMetadata() servletcontainer.Metadata {
	return servletcontainer.NewMetadata(
		Name,
		Version,
		[]servletcontainer.Profile{
			servletcontainer.ProfileCore,
			servletcontainer.ProfileSession,
			servletcontainer.ProfileMultipart,
			servletcontainer.ProfileAsyncStream,
			servletcontainer.ProfileUpgrade,
			servletcontainer.ProfileNativeIO,
			profileSecurity,
		},
		map[string]string{
			"arkarta":   "v0.0.1",
			"transport": "net/http",
		},
	)
}
