package hertz

import (
	"time"

	"goark.dev/arkarta/servlet"
)

const defaultServerAddress = ":8080"

// ServerOption 定制 Hertz HTTP Server。
type ServerOption func(*serverOptions)

type serverOptions struct {
	address             string
	readTimeout         time.Duration
	readTimeoutSet      bool
	writeTimeout        time.Duration
	writeTimeoutSet     bool
	idleTimeout         time.Duration
	idleTimeoutSet      bool
	keepAliveTimeout    time.Duration
	keepAliveTimeoutSet bool
	exitWaitTimeout     time.Duration
	maxHeaderBytes      int
	maxRequestBodySize  int
}

// WithAddress 设置监听地址。
func WithAddress(address string) ServerOption {
	return func(options *serverOptions) {
		if address != "" {
			options.address = address
		}
	}
}

// WithReadTimeout 设置请求读取超时。
func WithReadTimeout(timeout time.Duration) ServerOption {
	return func(options *serverOptions) {
		if timeout >= 0 {
			options.readTimeout = timeout
			options.readTimeoutSet = true
		}
	}
}

// WithWriteTimeout 设置响应写出超时。
func WithWriteTimeout(timeout time.Duration) ServerOption {
	return func(options *serverOptions) {
		if timeout >= 0 {
			options.writeTimeout = timeout
			options.writeTimeoutSet = true
		}
	}
}

// WithIdleTimeout 设置 keep-alive 空闲超时。
func WithIdleTimeout(timeout time.Duration) ServerOption {
	return func(options *serverOptions) {
		if timeout >= 0 {
			options.idleTimeout = timeout
			options.idleTimeoutSet = true
		}
	}
}

// WithKeepAliveTimeout 设置 TCP keep-alive 探测间隔。
func WithKeepAliveTimeout(timeout time.Duration) ServerOption {
	return func(options *serverOptions) {
		if timeout >= 0 {
			options.keepAliveTimeout = timeout
			options.keepAliveTimeoutSet = true
		}
	}
}

// WithExitWaitTimeout 设置 Hertz 优雅关闭等待上限。
func WithExitWaitTimeout(timeout time.Duration) ServerOption {
	return func(options *serverOptions) {
		if timeout > 0 {
			options.exitWaitTimeout = timeout
		}
	}
}

// WithMaxHeaderBytes 设置请求头最大字节数。
func WithMaxHeaderBytes(size int) ServerOption {
	return func(options *serverOptions) {
		if size > 0 {
			options.maxHeaderBytes = size
		}
	}
}

// WithMaxRequestBodySize 设置整个请求体最大字节数。
func WithMaxRequestBodySize(size int) ServerOption {
	return func(options *serverOptions) {
		if size > 0 {
			options.maxRequestBodySize = size
		}
	}
}

func buildServerOptions(options []ServerOption) serverOptions {
	config := serverOptions{
		address:            defaultServerAddress,
		maxRequestBodySize: int(servlet.DefaultMaxFormBodySize),
	}
	for _, option := range options {
		if option != nil {
			option(&config)
		}
	}
	return config
}
