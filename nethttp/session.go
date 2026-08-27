package nethttp

import (
	"context"

	"goark.dev/arkarta/servlet"
	"goark.dev/arkarta/servlet/session"
)

// SessionManagerFactory 为每个 WebApp 创建会话管理器。
type SessionManagerFactory func(app *servlet.WebApp) (session.Manager, error)

// NewSessionManager 创建默认内存会话管理器。
func NewSessionManager(options ...session.MemoryManagerOption) *session.MemoryManager {
	return session.NewMemoryManager(options...)
}

// NewMemorySessionManager 创建内存会话管理器。
func NewMemorySessionManager(options ...session.MemoryManagerOption) *session.MemoryManager {
	return session.NewMemoryManager(options...)
}

// GetSession 返回当前请求会话；create 为 true 时会创建并写回会话 Cookie。
func GetSession(ctx context.Context, req *servlet.Request, res servlet.Response, create bool) (session.Session, bool, error) {
	profile, ok := currentProfile(req)
	if !ok || profile.sessionAccessor == nil {
		return nil, false, ErrSessionProfileUnavailable
	}
	return profile.sessionAccessor.Get(ctx, req, res, create)
}

// CurrentSession 返回已经绑定到当前请求的会话。
func CurrentSession(req *servlet.Request) (session.Session, bool) {
	return session.Current(req)
}

// RequestedSessionID 返回客户端提交的原始会话 ID。
func RequestedSessionID(req *servlet.Request) (string, bool) {
	profile, ok := currentProfile(req)
	if !ok || profile.sessionAccessor == nil {
		return "", false
	}
	return profile.sessionAccessor.RequestedID(req)
}

// RequestedSessionIDValid 判断客户端提交的会话 ID 是否有效。
func RequestedSessionIDValid(ctx context.Context, req *servlet.Request) (bool, error) {
	profile, ok := currentProfile(req)
	if !ok || profile.sessionAccessor == nil {
		return false, ErrSessionProfileUnavailable
	}
	return profile.sessionAccessor.RequestedIDValid(ctx, req)
}

// ChangeSessionID 轮换当前请求关联的会话 ID。
func ChangeSessionID(ctx context.Context, req *servlet.Request, res servlet.Response) (string, error) {
	profile, ok := currentProfile(req)
	if !ok || profile.sessionAccessor == nil {
		return "", ErrSessionProfileUnavailable
	}
	return profile.sessionAccessor.ChangeID(ctx, req, res)
}

// EncodeSessionURL 按当前会话跟踪策略重写 URL。
func EncodeSessionURL(req *servlet.Request, rawURL string) (string, error) {
	profile, ok := currentProfile(req)
	if !ok || profile.sessionAccessor == nil {
		return rawURL, ErrSessionProfileUnavailable
	}
	return profile.sessionAccessor.EncodeURL(req, rawURL)
}

// EncodeSessionRedirectURL 按当前会话跟踪策略重写重定向 URL。
func EncodeSessionRedirectURL(req *servlet.Request, rawURL string) (string, error) {
	profile, ok := currentProfile(req)
	if !ok || profile.sessionAccessor == nil {
		return rawURL, ErrSessionProfileUnavailable
	}
	return profile.sessionAccessor.EncodeRedirectURL(req, rawURL)
}
