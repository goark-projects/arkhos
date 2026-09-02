package hertz

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"

	internalcontainer "goark.dev/arkhos/internal/container"
	internalprofile "goark.dev/arkhos/internal/profile"

	"goark.dev/arkarta/servlet"
	servletcontainer "goark.dev/arkarta/servlet/container"
	"goark.dev/arkarta/servlet/nativeio"
)

// Container 是基于 Hertz 的 Arkhos Servlet 容器。
type Container struct {
	runtime        *internalcontainer.Runtime
	sender         nativeio.Sender
	requestOptions []servlet.RequestOption
}

// NewContainer 创建 Arkhos Hertz 容器。
func NewContainer(options ...ContainerOption) *Container {
	config := buildContainerOptions(options)
	return &Container{
		runtime: internalcontainer.NewRuntime(
			defaultMetadata(),
			internalcontainer.WithApplicationDecorator(func(application servletcontainer.Application) (servletcontainer.Application, error) {
				return internalprofile.Decorate(application, config.profiles)
			}),
		),
		sender:         nativeio.NewStandardSender(),
		requestOptions: append([]servlet.RequestOption(nil), config.requestOptions...),
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
		return nil, ErrNilContainer
	}
	return c.runtime.Deploy(ctx, deployment)
}

// Start 标记容器开始服务请求。
func (c *Container) Start(ctx context.Context) error {
	if c == nil || c.runtime == nil {
		return ErrNilContainer
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

// Handler 返回容器聚合后的 Hertz Handler。
func (c *Container) Handler() app.HandlerFunc {
	return func(ctx context.Context, requestContext *app.RequestContext) {
		if c == nil || c.runtime == nil {
			requestContext.SetStatusCode(consts.StatusServiceUnavailable)
			return
		}
		path := string(requestContext.Path())
		application, status := c.runtime.MatchApplication(path)
		switch status {
		case internalcontainer.MatchFound:
			serve(ctx, requestContext, application.Handler(), application.WebApp().ContextPath(), c.requestOptions)
		case internalcontainer.MatchUnavailable:
			requestContext.SetStatusCode(consts.StatusServiceUnavailable)
		default:
			requestContext.SetStatusCode(consts.StatusNotFound)
		}
	}
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
			servletcontainer.ProfileNativeIO,
			profileSecurity,
		},
		map[string]string{"arkarta": "v0.0.1", "transport": "hertz"},
	)
}

const profileSecurity servletcontainer.Profile = "security"
