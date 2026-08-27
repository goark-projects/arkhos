package nethttp

import (
	"goark.dev/arkarta/servlet"
	"goark.dev/arkarta/servlet/async"
	"goark.dev/arkarta/servlet/multipart"
	"goark.dev/arkarta/servlet/session"
)

// ContainerOption 定制 Arkhos net/http 容器。
type ContainerOption func(*containerOptions)

type containerOptions struct {
	sessionManagerFactory SessionManagerFactory
	multipartParser       MultipartParserFactory
	securityPolicy        SecurityPolicy
	asyncOptions          []async.Option
}

func defaultContainerOptions() containerOptions {
	return containerOptions{
		sessionManagerFactory: func(*servlet.WebApp) (session.Manager, error) {
			return session.NewMemoryManager(), nil
		},
		multipartParser: func() *multipart.Parser {
			return multipart.NewParser()
		},
	}
}

// WithSessionManagerFactory 设置每个 WebApp 的会话管理器工厂。
func WithSessionManagerFactory(factory SessionManagerFactory) ContainerOption {
	return func(options *containerOptions) {
		if factory != nil {
			options.sessionManagerFactory = factory
		}
	}
}

// WithMultipartParserFactory 设置 multipart 解析器工厂。
func WithMultipartParserFactory(factory MultipartParserFactory) ContainerOption {
	return func(options *containerOptions) {
		if factory != nil {
			options.multipartParser = factory
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
		options.securityPolicy = policy
	}
}

// WithAsyncOptions 设置容器创建异步上下文时追加的默认选项。
func WithAsyncOptions(asyncOptions ...async.Option) ContainerOption {
	copied := append([]async.Option(nil), asyncOptions...)
	return func(options *containerOptions) {
		options.asyncOptions = append(options.asyncOptions, copied...)
	}
}

func buildContainerOptions(options []ContainerOption) containerOptions {
	cfg := defaultContainerOptions()
	for _, option := range options {
		if option != nil {
			option(&cfg)
		}
	}
	return cfg
}
