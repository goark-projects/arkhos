package hertz

import (
	internalprofile "goark.dev/arkhos/internal/profile"

	"goark.dev/arkarta/servlet/async"
	"goark.dev/arkarta/servlet/multipart"
)

// ContainerOption 定制 Arkhos Hertz 容器。
type ContainerOption func(*internalprofile.Config)

// WithSessionManagerFactory 设置每个 WebApp 的会话管理器工厂。
func WithSessionManagerFactory(factory SessionManagerFactory) ContainerOption {
	return func(options *internalprofile.Config) {
		if factory != nil {
			options.SessionManagerFactory = factory
		}
	}
}

// WithMultipartParserFactory 设置 multipart 解析器工厂。
func WithMultipartParserFactory(factory MultipartParserFactory) ContainerOption {
	return func(options *internalprofile.Config) {
		if factory != nil {
			options.MultipartParser = factory
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
	return func(options *internalprofile.Config) {
		options.SecurityPolicy = policy
	}
}

// WithAsyncOptions 设置容器创建异步上下文时追加的默认选项。
func WithAsyncOptions(options ...async.Option) ContainerOption {
	copied := append([]async.Option(nil), options...)
	return func(config *internalprofile.Config) {
		config.AsyncOptions = append(config.AsyncOptions, copied...)
	}
}

func buildContainerOptions(options []ContainerOption) internalprofile.Config {
	config := internalprofile.DefaultConfig()
	for _, option := range options {
		if option != nil {
			option(&config)
		}
	}
	return config
}
