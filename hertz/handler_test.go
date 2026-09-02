package hertz

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"

	"goark.dev/arkarta/servlet"
)

func TestHandlerBridgesRequestAndResponse(t *testing.T) {
	handler := Handler(servlet.HandlerFunc(func(_ context.Context, req *servlet.Request, res servlet.Response) error {
		cookie, err := req.Cookie("sid")
		if err != nil {
			return err
		}
		body, err := io.ReadAll(req.Body())
		if err != nil {
			return err
		}
		res.SetStatus(consts.StatusCreated)
		res.Header().Set("X-Result", req.Header().Get("X-Trace-ID"))
		_, err = res.WriteString(cookie.Value + ":" + string(body))
		return err
	}))

	ctx := app.NewContext(0)
	ctx.Request.Header.SetMethod("POST")
	ctx.Request.SetRequestURI("/orders")
	ctx.Request.Header.Set("X-Trace-ID", "trace-1")
	ctx.Request.Header.SetCookie("sid", "session-1")
	ctx.Request.SetBodyString("payload")
	handler(t.Context(), ctx)

	if ctx.Response.StatusCode() != consts.StatusCreated {
		t.Fatalf("status = %d, want 201", ctx.Response.StatusCode())
	}
	if got := string(ctx.Response.Header.Peek("X-Result")); got != "trace-1" {
		t.Fatalf("X-Result = %q", got)
	}
	if got := string(ctx.Response.Body()); got != "session-1:payload" {
		t.Fatalf("body = %q", got)
	}
}

func TestHandlerMapsServletError(t *testing.T) {
	handler := Handler(servlet.HandlerFunc(func(context.Context, *servlet.Request, servlet.Response) error {
		return servlet.NewHTTPError(consts.StatusConflict, "conflict", errors.New("internal"))
	}))
	ctx := app.NewContext(0)
	ctx.Request.SetRequestURI("/")
	handler(t.Context(), ctx)

	if ctx.Response.StatusCode() != consts.StatusConflict || string(ctx.Response.Body()) != "conflict\n" {
		t.Fatalf("response = %d/%q", ctx.Response.StatusCode(), ctx.Response.Body())
	}
}

func TestHandlerRecoversPanic(t *testing.T) {
	handler := Handler(servlet.HandlerFunc(func(context.Context, *servlet.Request, servlet.Response) error {
		panic("test panic")
	}))
	ctx := app.NewContext(0)
	ctx.Request.SetRequestURI("/")
	handler(t.Context(), ctx)

	if ctx.Response.StatusCode() != consts.StatusInternalServerError || string(ctx.Response.Body()) != "Internal Server Error\n" {
		t.Fatalf("response = %d/%q", ctx.Response.StatusCode(), ctx.Response.Body())
	}
}

func BenchmarkHandlerDirectBridge(b *testing.B) {
	handler := Handler(servlet.HandlerFunc(func(_ context.Context, req *servlet.Request, res servlet.Response) error {
		res.Header().Set("Content-Type", "text/plain")
		_, err := res.WriteString(req.Path())
		return err
	}))
	b.ReportAllocs()
	for b.Loop() {
		ctx := app.NewContext(0)
		ctx.Request.Header.SetMethod("GET")
		ctx.Request.SetRequestURI("/benchmark")
		handler(context.Background(), ctx)
	}
}
