package profile

import (
	"context"

	"goark.dev/arkarta/servlet"
	"goark.dev/arkarta/servlet/async"
	"goark.dev/arkarta/servlet/multipart"
	"goark.dev/arkarta/servlet/session"
)

// SessionManagerFactory 为每个 WebApp 创建会话管理器。
type SessionManagerFactory func(app *servlet.WebApp) (session.Manager, error)

// MultipartParserFactory 创建 multipart 解析器。
type MultipartParserFactory func() *multipart.Parser

// SecurityPolicy 在请求进入应用处理器前执行认证和授权。
type SecurityPolicy interface {
	Apply(ctx context.Context, req *servlet.Request, res servlet.Response) error
}

// SecurityPolicyFunc 将函数适配为安全策略。
type SecurityPolicyFunc func(ctx context.Context, req *servlet.Request, res servlet.Response) error

func (f SecurityPolicyFunc) Apply(ctx context.Context, req *servlet.Request, res servlet.Response) error {
	return f(ctx, req, res)
}

// Config 描述容器共享的 Servlet Profile 装配策略。
type Config struct {
	SessionManagerFactory SessionManagerFactory
	MultipartParser       MultipartParserFactory
	SecurityPolicy        SecurityPolicy
	AsyncOptions          []async.Option
}

// DefaultConfig 创建生产默认 Profile 配置。
func DefaultConfig() Config {
	return Config{
		SessionManagerFactory: func(*servlet.WebApp) (session.Manager, error) {
			return session.NewMemoryManager(), nil
		},
		MultipartParser: func() *multipart.Parser {
			return multipart.NewParser()
		},
	}
}
