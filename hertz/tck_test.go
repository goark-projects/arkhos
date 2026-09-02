package hertz

import (
	"context"
	"testing"

	"github.com/cloudwego/hertz/pkg/app"

	"goark.dev/arkarta/servlet"
	"goark.dev/arkarta/servlet/tck"
)

func TestHandlerCoreTCK(t *testing.T) {
	tck.RunCore(t, tck.DriverFunc(exchangeTCKRequest))
}

func exchangeTCKRequest(ctx context.Context, handler servlet.Handler, request tck.Request) (tck.Response, error) {
	requestContext := app.NewContext(0)
	requestContext.Request.Header.SetMethod(request.Method)
	requestContext.Request.SetRequestURI(request.Target)
	if request.Header != nil {
		request.Header.Visit(func(name, value string) bool {
			requestContext.Request.Header.Add(name, value)
			return true
		})
	}
	requestContext.Request.SetBody(request.Body)

	Handler(handler)(ctx, requestContext)
	return tck.Response{
		Status: requestContext.Response.StatusCode(),
		Header: servlet.CloneHeader(responseHeader{header: &requestContext.Response.Header}),
		Body:   append([]byte(nil), requestContext.Response.Body()...),
	}, nil
}
