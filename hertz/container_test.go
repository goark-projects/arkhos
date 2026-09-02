package hertz

import (
	"context"
	"testing"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"

	"goark.dev/arkarta/servlet"
	servletcontainer "goark.dev/arkarta/servlet/container"
)

func TestContainerDispatchesDirectHertzRequest(t *testing.T) {
	container := NewContainer()
	appDefinition, err := servlet.NewWebApp("orders", servlet.WithContextPath("/orders"))
	if err != nil {
		t.Fatalf("NewWebApp failed: %v", err)
	}
	deployment, err := servletcontainer.NewDeployment(appDefinition, servletcontainer.WithMapping("/", servlet.HandlerFunc(func(_ context.Context, req *servlet.Request, res servlet.Response) error {
		res.Header().Set("X-Transport", "hertz")
		_, writeErr := res.WriteString(req.Method() + " " + req.Path())
		return writeErr
	})))
	if err != nil {
		t.Fatalf("NewDeployment failed: %v", err)
	}
	if _, err := container.Deploy(t.Context(), deployment); err != nil {
		t.Fatalf("Deploy failed: %v", err)
	}
	if err := container.Start(t.Context()); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	ctx := app.NewContext(0)
	ctx.Request.Header.SetMethod("GET")
	ctx.Request.SetRequestURI("/orders/42?full=true")
	container.Handler()(t.Context(), ctx)

	if ctx.Response.StatusCode() != consts.StatusOK {
		t.Fatalf("status = %d, want 200", ctx.Response.StatusCode())
	}
	if got := string(ctx.Response.Header.Peek("X-Transport")); got != "hertz" {
		t.Fatalf("X-Transport = %q", got)
	}
	if got := string(ctx.Response.Body()); got != "GET /42" {
		t.Fatalf("body = %q, want GET /42", got)
	}
}

func TestContainerRejectsRequestsBeforeStart(t *testing.T) {
	ctx := app.NewContext(0)
	ctx.Request.SetRequestURI("/")
	NewContainer().Handler()(t.Context(), ctx)
	if ctx.Response.StatusCode() != consts.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", ctx.Response.StatusCode())
	}
}
