package nethttp

import (
	"context"

	"goark.dev/arkarta/servlet"
	"goark.dev/arkarta/servlet/upgrade"
	"goark.dev/arkarta/websocket"
	servletwebsocket "goark.dev/arkarta/websocket/servlet"
)

// NewWebSocketHandshaker 创建 WebSocket 握手协商器。
func NewWebSocketHandshaker(options ...websocket.HandshakeOption) *websocket.Handshaker {
	return websocket.NewHandshaker(options...)
}

// NewPerMessageDeflate 创建 permessage-deflate 协商器。
func NewPerMessageDeflate(options ...websocket.PerMessageDeflateOption) *websocket.PerMessageDeflate {
	return websocket.NewPerMessageDeflate(options...)
}

// NewWebSocketSession 创建标准 WebSocket 会话。
func NewWebSocketSession(id string, conn websocket.Connection, options ...websocket.SessionOption) (*websocket.StandardSession, error) {
	return websocket.NewSession(id, conn, options...)
}

// WebSocketHandler 将 WebSocket Endpoint 暴露为 Servlet Handler。
func WebSocketHandler(sessionID string, endpoint websocket.Endpoint, options ...websocket.HandshakeOption) servlet.Handler {
	return servlet.HandlerFunc(func(ctx context.Context, req *servlet.Request, res servlet.Response) error {
		_, err := servletwebsocket.Upgrade(ctx, req, res, servletwebsocket.EndpointHandler(sessionID, endpoint), options...)
		return err
	})
}

// WebSocketUpgrade 执行 WebSocket 握手并移交升级连接。
func WebSocketUpgrade(ctx context.Context, req *servlet.Request, res servlet.Response, handler servletwebsocket.Handler, options ...websocket.HandshakeOption) (websocket.Handshake, error) {
	return servletwebsocket.Upgrade(ctx, req, res, handler, options...)
}

// WebSocketEndpointHandler 创建基于升级连接运行 Endpoint 的处理器。
func WebSocketEndpointHandler(sessionID string, endpoint websocket.Endpoint, options ...servletwebsocket.FrameConnectionOption) servletwebsocket.Handler {
	return servletwebsocket.HandlerFunc(func(ctx context.Context, handshake websocket.Handshake, conn upgrade.Connection) error {
		return servletwebsocket.ServeEndpoint(ctx, sessionID, handshake, conn, endpoint, options...)
	})
}
