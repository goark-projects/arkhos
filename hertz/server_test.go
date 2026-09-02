package hertz

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"goark.dev/arkarta/servlet"
	servletcontainer "goark.dev/arkarta/servlet/container"
)

func TestNewServerRejectsNilContainer(t *testing.T) {
	if _, err := NewServer(nil); !errors.Is(err, ErrNilContainer) {
		t.Fatalf("NewServer err = %v, want ErrNilContainer", err)
	}
}

func TestNewServerAppliesOptions(t *testing.T) {
	server, err := NewServer(NewContainer(),
		WithAddress("127.0.0.1:9090"),
		WithReadTimeout(0),
		WithWriteTimeout(time.Second),
		WithIdleTimeout(2*time.Second),
		WithKeepAliveTimeout(3*time.Second),
		WithExitWaitTimeout(4*time.Second),
		WithMaxHeaderBytes(4096),
		WithMaxRequestBodySize(8192),
	)
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}
	options := server.newEngine(nil).GetOptions()
	if options.Addr != "127.0.0.1:9090" || options.ReadTimeout != 0 ||
		options.WriteTimeout != time.Second || options.IdleTimeout != 2*time.Second ||
		options.KeepAliveTimeout != 3*time.Second || options.ExitWaitTimeout != 4*time.Second ||
		options.MaxHeaderBytes != 4096 || options.MaxRequestBodySize != 8192 ||
		!options.DisablePreParseMultipartForm {
		t.Fatalf("Hertz options were not applied: %#v", options)
	}
}

func TestServerServeStartsContainerAndShutsDown(t *testing.T) {
	container := NewContainer()
	app, err := servlet.NewWebApp("orders")
	if err != nil {
		t.Fatalf("NewWebApp failed: %v", err)
	}
	deployment, err := servletcontainer.NewDeployment(app,
		servletcontainer.WithMapping("/", servlet.HandlerFunc(func(_ context.Context, _ *servlet.Request, res servlet.Response) error {
			_, err := res.WriteString("served")
			return err
		})),
	)
	if err != nil {
		t.Fatalf("NewDeployment failed: %v", err)
	}
	if _, err := container.Deploy(context.Background(), deployment); err != nil {
		t.Fatalf("Deploy failed: %v", err)
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Listen failed: %v", err)
	}
	defer listener.Close()
	server, err := NewServer(container, WithReadTimeout(time.Second), WithMaxHeaderBytes(4096))
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() {
		errCh <- server.Serve(ctx, listener)
	}()
	body := requestHertzUntilOK(t, "http://"+listener.Addr().String()+"/")
	if body != "served" {
		t.Fatalf("body = %q, want served", body)
	}
	cancel()
	if err := <-errCh; !errors.Is(err, context.Canceled) {
		t.Fatalf("Serve err = %v, want context.Canceled", err)
	}
	if app.State() != servlet.WebAppStateDestroyed {
		t.Fatalf("web app state = %v, want destroyed", app.State())
	}
}

func requestHertzUntilOK(t *testing.T, target string) string {
	t.Helper()
	client := &http.Client{Transport: &http.Transport{DisableKeepAlives: true}}
	deadline := time.Now().Add(3 * time.Second)
	for {
		response, err := client.Get(target)
		if err == nil {
			data, readErr := io.ReadAll(response.Body)
			closeErr := response.Body.Close()
			if readErr != nil || closeErr != nil {
				t.Fatalf("read/close response = %v/%v", readErr, closeErr)
			}
			if response.StatusCode == http.StatusOK {
				return string(data)
			}
			t.Fatalf("status = %d, body = %q", response.StatusCode, data)
		}
		if time.Now().After(deadline) {
			t.Fatalf("GET %s did not succeed before deadline: %v", target, err)
		}
		time.Sleep(10 * time.Millisecond)
	}
}
