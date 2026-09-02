package hertz

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"

	"goark.dev/arkarta/servlet"
)

// Handler 将 Arkarta Servlet Handler 适配为 Hertz Handler。
func Handler(handler servlet.Handler, options ...servlet.RequestOption) app.HandlerFunc {
	requestOptions := append([]servlet.RequestOption(nil), options...)
	return func(ctx context.Context, requestContext *app.RequestContext) {
		serve(ctx, requestContext, handler, "", requestOptions)
	}
}

func serve(ctx context.Context, requestContext *app.RequestContext, handler servlet.Handler, contextPath string, options []servlet.RequestOption) {
	res := newResponse(requestContext)
	defer res.finish()
	defer recoverPanic(res)
	if handler == nil {
		writeError(res, servlet.NewHTTPError(consts.StatusInternalServerError, "handler is nil", servlet.ErrNilHandler))
		return
	}
	req, err := newRequest(ctx, requestContext, contextPath, options...)
	if err != nil {
		writeError(res, err)
		return
	}
	if err := handler.Serve(ctx, req, res); err != nil {
		writeError(res, err)
	}
}

func recoverPanic(res *response) {
	value := recover()
	if value == nil {
		return
	}
	err := fmt.Errorf("panic recovered: %v\n%s", value, debug.Stack())
	writeError(res, servlet.NewHTTPError(consts.StatusInternalServerError, "Internal Server Error", err))
}

func writeError(res *response, err error) {
	if res.Committed() {
		return
	}
	statusCode := consts.StatusInternalServerError
	message := "Internal Server Error"
	var statusErr servlet.StatusError
	if errors.As(err, &statusErr) {
		statusCode = statusErr.StatusCode()
		message = statusErr.PublicMessage()
	}
	res.Header().Set("Content-Type", "text/plain; charset=utf-8")
	res.SetStatus(statusCode)
	_, _ = res.WriteString(message + "\n")
}
