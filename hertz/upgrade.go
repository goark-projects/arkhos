package hertz

import (
	"context"

	"github.com/cloudwego/hertz/pkg/network"

	"goark.dev/arkarta/servlet"
	"goark.dev/arkarta/servlet/upgrade"
)

// UpgradeHTTP 通过 Hertz 原生连接劫持能力移交升级后的连接。
func (r *response) UpgradeHTTP(ctx context.Context, _ *servlet.Request, handler upgrade.Handler) error {
	if handler == nil {
		return upgrade.ErrNilHandler
	}
	if r == nil || r.ctx == nil {
		return servlet.ErrNilResponse
	}
	if r.committed {
		return upgrade.ErrAlreadyCommitted
	}
	r.committed = true
	r.ctx.Response.HijackWriter(discardHTTPResponseWriter{})
	r.ctx.Hijack(func(conn network.Conn) {
		_ = handler.ServeUpgrade(ctx, conn)
	})
	return nil
}

// discardHTTPResponseWriter 阻止 Hertz 在连接移交前重复写出 HTTP 响应。
type discardHTTPResponseWriter struct{}

func (discardHTTPResponseWriter) Write([]byte) (int, error) { return 0, nil }
func (discardHTTPResponseWriter) Flush() error              { return nil }
func (discardHTTPResponseWriter) Finalize() error           { return nil }

var _ upgrade.HTTPUpgrader = (*response)(nil)
var _ network.ExtWriter = discardHTTPResponseWriter{}
