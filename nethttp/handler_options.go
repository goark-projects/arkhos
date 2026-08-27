package nethttp

import (
	arkartanethttp "goark.dev/arkarta/servlet/nethttp"

	"goark.dev/arkarta/servlet"
)

// HandlerOption 定制 Servlet 到 net/http 的请求适配行为。
type HandlerOption func(*handlerOptions)

type handlerOptions struct {
	delegate []arkartanethttp.Option
}

// WithErrorPages 设置错误页注册表。
func WithErrorPages(registry *servlet.ErrorPageRegistry) HandlerOption {
	return func(options *handlerOptions) {
		options.delegate = append(options.delegate, arkartanethttp.WithErrorPages(registry))
	}
}

// WithRequestContextPath 设置适配器创建请求时使用的 Web 应用上下文路径。
func WithRequestContextPath(contextPath string) HandlerOption {
	return func(options *handlerOptions) {
		options.delegate = append(options.delegate, arkartanethttp.WithRequestContextPath(contextPath))
	}
}

func buildHandlerOptions(options []HandlerOption) []arkartanethttp.Option {
	cfg := &handlerOptions{}
	for _, option := range options {
		if option != nil {
			option(cfg)
		}
	}
	return cfg.delegate
}
