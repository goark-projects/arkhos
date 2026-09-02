package hertz

import (
	internalprofile "goark.dev/arkhos/internal/profile"

	"goark.dev/arkarta/servlet"
	"goark.dev/arkarta/servlet/async"
	"goark.dev/arkarta/servlet/multipart"
)

// ContainerOption 定制 Arkhos Hertz 容器。
type ContainerOption func(*containerOptions)

type containerOptions struct {
	profiles       internalprofile.Config
	requestOptions []servlet.RequestOption
}

// WithSessionManagerFactory 设置每个 WebApp 的会话管理器工厂。
func WithSessionManagerFactory(factory SessionManagerFactory) ContainerOption {
	return func(options *containerOptions) {
		if factory != nil {
			options.profiles.SessionManagerFactory = factory
		}
	}
}

// WithMultipartParserFactory 设置 multipart 解析器工厂。
func WithMultipartParserFactory(factory MultipartParserFactory) ContainerOption {
	return func(options *containerOptions) {
		if factory != nil {
			options.profiles.MultipartParser = factory
		}
	}
}

// WithMultipartConfig 使用标准 multipart 配置创建解析器。
func WithMultipartConfig(config multipart.Config) ContainerOption {
	return WithMultipartParserFactory(func() *multipart.Parser {
		return multipart.NewParser(multipart.WithConfig(config))
	})
}

// WithSecurityPolicy 设置容器级安全策略。
func WithSecurityPolicy(policy SecurityPolicy) ContainerOption {
	return func(options *containerOptions) {
		options.profiles.SecurityPolicy = policy
	}
}

// WithAsyncOptions 设置容器创建异步上下文时追加的默认选项。
func WithAsyncOptions(options ...async.Option) ContainerOption {
	copied := append([]async.Option(nil), options...)
	return func(config *containerOptions) {
		config.profiles.AsyncOptions = append(config.profiles.AsyncOptions, copied...)
	}
}

// WithMaxFormBodySize 设置 URL 编码表单体解析上限。
func WithMaxFormBodySize(size int64) ContainerOption {
	return func(config *containerOptions) {
		if size > 0 {
			config.requestOptions = append(config.requestOptions, servlet.WithMaxFormBodySize(size))
		}
	}
}

func buildContainerOptions(options []ContainerOption) containerOptions {
	config := containerOptions{profiles: internalprofile.DefaultConfig()}
	for _, option := range options {
		if option != nil {
			option(&config)
		}
	}
	return config
}
