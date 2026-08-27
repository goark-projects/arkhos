package nethttp

import (
	"context"

	"goark.dev/arkarta/servlet"
	"goark.dev/arkarta/servlet/upgrade"
)

// UpgradeHTTP 将当前 HTTP 请求升级并移交连接所有权。
func UpgradeHTTP(ctx context.Context, req *servlet.Request, res servlet.Response, handler upgrade.Handler) error {
	return upgrade.HTTP(ctx, req, res, handler)
}
