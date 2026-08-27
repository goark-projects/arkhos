package nethttp

import (
	"context"

	"goark.dev/arkarta/servlet"
	"goark.dev/arkarta/servlet/async"
)

// StartAsync 创建当前请求的异步上下文。
func StartAsync(ctx context.Context, req *servlet.Request, res servlet.Response, options ...async.Option) (*async.Context, error) {
	profile, _ := currentProfile(req)
	merged := make([]async.Option, 0, len(options))
	if profile != nil {
		merged = append(merged, profile.asyncOptions...)
	}
	merged = append(merged, options...)
	return async.NewContext(ctx, req, res, merged...)
}

// NewAsyncStream 创建流式响应写入器。
func NewAsyncStream(res servlet.Response) (*async.Stream, error) {
	return async.NewStream(res)
}
