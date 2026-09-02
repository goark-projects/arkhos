package hertz

import (
	"context"

	internalprofile "goark.dev/arkhos/internal/profile"

	"goark.dev/arkarta/servlet"
	"goark.dev/arkarta/servlet/async"
)

// StartAsync 创建当前请求的异步上下文。
func StartAsync(ctx context.Context, req *servlet.Request, res servlet.Response, options ...async.Option) (*async.Context, error) {
	return internalprofile.StartAsync(ctx, req, res, options...)
}

// NewAsyncStream 创建流式响应写入器。
func NewAsyncStream(res servlet.Response) (*async.Stream, error) {
	return async.NewStream(res)
}
