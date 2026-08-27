package container

import (
	"context"
	"errors"
	"testing"

	"goark.dev/arkarta/servlet"
	servletcontainer "goark.dev/arkarta/servlet/container"
)

func TestRuntimeDeployStopsManagedApplicationWhenDecoratorFails(t *testing.T) {
	t.Parallel()

	destroyed := false
	deployment := newRuntimeTestDeployment(t, "decorator-failure", servlet.WithContextListener(servlet.ContextListenerFunc{
		Destroyed: func(context.Context, servlet.ContextEvent) error {
			destroyed = true
			return nil
		},
	}))
	failure := errors.New("decorator failed")
	runtime := NewRuntime(runtimeTestMetadata(), WithApplicationDecorator(func(servletcontainer.Application) (servletcontainer.Application, error) {
		return nil, failure
	}))

	if _, err := runtime.Deploy(context.Background(), deployment); !errors.Is(err, failure) {
		t.Fatalf("Deploy err = %v, want decorator failure", err)
	}
	if !destroyed {
		t.Fatal("decorator failure should stop and destroy the managed application")
	}
}

func TestRuntimeDeployRejectsNilDecoratedApplication(t *testing.T) {
	t.Parallel()

	deployment := newRuntimeTestDeployment(t, "nil-decorator")
	runtime := NewRuntime(runtimeTestMetadata(), WithApplicationDecorator(func(servletcontainer.Application) (servletcontainer.Application, error) {
		return nil, nil
	}))

	if _, err := runtime.Deploy(context.Background(), deployment); !errors.Is(err, ErrNilDecoratedApplication) {
		t.Fatalf("Deploy err = %v, want ErrNilDecoratedApplication", err)
	}
}

func newRuntimeTestDeployment(t *testing.T, name string, options ...servlet.WebAppOption) *servletcontainer.Deployment {
	t.Helper()

	app, err := servlet.NewWebApp(name, options...)
	if err != nil {
		t.Fatalf("NewWebApp failed: %v", err)
	}
	deployment, err := servletcontainer.NewDeployment(app, servletcontainer.WithMapping("/", servlet.HandlerFunc(func(context.Context, *servlet.Request, servlet.Response) error {
		return nil
	})))
	if err != nil {
		t.Fatalf("NewDeployment failed: %v", err)
	}
	return deployment
}

func runtimeTestMetadata() servletcontainer.Metadata {
	return servletcontainer.NewMetadata(
		"arkhos-test",
		"v0",
		[]servletcontainer.Profile{servletcontainer.ProfileCore},
		nil,
	)
}
