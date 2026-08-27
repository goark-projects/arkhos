package tck_test

import (
	"context"
	"net/http"
	"testing"

	"goark.dev/arkhos/nethttp"

	"goark.dev/arkarta/servlet"
	servletcontainer "goark.dev/arkarta/servlet/container"
	"goark.dev/arkarta/servlet/tck"
)

func TestServletLifecycle(t *testing.T) {
	tck.RunLifecycle(t, func(deployment *servletcontainer.Deployment) (servletcontainer.Application, error) {
		return nethttp.NewContainer().Deploy(context.Background(), deployment)
	})
}

func TestServletErrorPages(t *testing.T) {
	tck.RunErrorPages(t, func(handler servlet.Handler, registry *servlet.ErrorPageRegistry) http.Handler {
		return nethttp.HandlerWithOptions(handler, nethttp.WithErrorPages(registry))
	})
}

func TestServletStaticResources(t *testing.T) {
	tck.RunStaticResources(t, nethttp.Handler)
}
