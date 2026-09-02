package hertz

import (
	"context"

	internalprofile "goark.dev/arkhos/internal/profile"

	"goark.dev/arkarta/servlet"
	"goark.dev/arkarta/servlet/session"
)

// SessionManagerFactory 为每个 WebApp 创建会话管理器。
type SessionManagerFactory = internalprofile.SessionManagerFactory

// NewSessionManager 创建内存会话管理器。
func NewSessionManager(options ...session.MemoryManagerOption) *session.MemoryManager {
	return session.NewMemoryManager(options...)
}

// NewMemorySessionManager 创建内存会话管理器。
func NewMemorySessionManager(options ...session.MemoryManagerOption) *session.MemoryManager {
	return session.NewMemoryManager(options...)
}

// GetSession 获取当前请求的会话，create 为 true 时按需创建。
func GetSession(ctx context.Context, req *servlet.Request, res servlet.Response, create bool) (session.Session, bool, error) {
	return internalprofile.GetSession(ctx, req, res, create)
}

// CurrentSession 返回已绑定到请求的会话。
func CurrentSession(req *servlet.Request) (session.Session, bool) {
	return session.Current(req)
}

// RequestedSessionID 返回客户端提交的会话标识。
func RequestedSessionID(req *servlet.Request) (string, bool) {
	return internalprofile.RequestedSessionID(req)
}

// RequestedSessionIDValid 判断客户端提交的会话标识是否有效。
func RequestedSessionIDValid(ctx context.Context, req *servlet.Request) (bool, error) {
	return internalprofile.RequestedSessionIDValid(ctx, req)
}

// ChangeSessionID 轮换当前会话标识。
func ChangeSessionID(ctx context.Context, req *servlet.Request, res servlet.Response) (string, error) {
	return internalprofile.ChangeSessionID(ctx, req, res)
}

// EncodeSessionURL 在需要时把会话标识编码到普通 URL。
func EncodeSessionURL(req *servlet.Request, rawURL string) (string, error) {
	return internalprofile.EncodeSessionURL(req, rawURL)
}

// EncodeSessionRedirectURL 在需要时把会话标识编码到重定向 URL。
func EncodeSessionRedirectURL(req *servlet.Request, rawURL string) (string, error) {
	return internalprofile.EncodeSessionRedirectURL(req, rawURL)
}
