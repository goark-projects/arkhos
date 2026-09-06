package nethttp

import (
	"context"

	internalprofile "goark.dev/arkhos/internal/profile"

	"goark.dev/arkarta/servlet"
	"goark.dev/arkarta/servlet/session"
)

// SessionManagerFactory 为每个 WebApp 创建会话管理器。
type SessionManagerFactory = internalprofile.SessionManagerFactory

// NewSessionManager 创建默认内存会话管理器。
func NewSessionManager(options ...session.MemoryManagerOption) *session.MemoryManager {
	return session.NewMemoryManager(options...)
}

// NewMemorySessionManager 创建内存会话管理器。
func NewMemorySessionManager(options ...session.MemoryManagerOption) *session.MemoryManager {
	return session.NewMemoryManager(options...)
}

// GetSession 返回当前请求会话；create 为 true 时会创建并写回会话 Cookie。
func GetSession(
	ctx context.Context,
	req *servlet.Request,
	res servlet.Response,
	create bool,
) (session.Session, bool, error) {
	return internalprofile.GetSession(ctx, req, res, create)
}

// CurrentSession 返回已经绑定到当前请求的会话。
func CurrentSession(req *servlet.Request) (session.Session, bool) {
	return session.Current(req)
}

// RequestedSessionID 返回客户端提交的原始会话 ID。
func RequestedSessionID(req *servlet.Request) (string, bool) {
	return internalprofile.RequestedSessionID(req)
}

// RequestedSessionIDValid 判断客户端提交的会话 ID 是否有效。
func RequestedSessionIDValid(ctx context.Context, req *servlet.Request) (bool, error) {
	return internalprofile.RequestedSessionIDValid(ctx, req)
}

// ChangeSessionID 轮换当前请求关联的会话 ID。
func ChangeSessionID(
	ctx context.Context,
	req *servlet.Request,
	res servlet.Response,
) (string, error) {
	return internalprofile.ChangeSessionID(ctx, req, res)
}

// EncodeSessionURL 按当前会话跟踪策略重写 URL。
func EncodeSessionURL(req *servlet.Request, rawURL string) (string, error) {
	return internalprofile.EncodeSessionURL(req, rawURL)
}

// EncodeSessionRedirectURL 按当前会话跟踪策略重写重定向 URL。
func EncodeSessionRedirectURL(req *servlet.Request, rawURL string) (string, error) {
	return internalprofile.EncodeSessionRedirectURL(req, rawURL)
}
