package nethttp

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
	server, err := NewServer(
		NewContainer(),
		WithAddress("127.0.0.1:0"),
		WithReadTimeout(time.Second),
		WithReadHeaderTimeout(2*time.Second),
		WithWriteTimeout(3*time.Second),
		WithIdleTimeout(4*time.Second),
		WithMaxHeaderBytes(4096),
	)
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}
	httpServer := server.HTTPServer()
	if httpServer.Addr != "127.0.0.1:0" ||
		httpServer.ReadTimeout != time.Second ||
		httpServer.ReadHeaderTimeout != 2*time.Second ||
		httpServer.WriteTimeout != 3*time.Second ||
		httpServer.IdleTimeout != 4*time.Second ||
		httpServer.MaxHeaderBytes != 4096 {
		t.Fatalf("http server options were not applied: %#v", httpServer)
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

	ctx, cancel := context.WithCancel(context.Background())
	server, err := NewServer(container)
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- server.Serve(ctx, listener)
	}()

	body := requestUntilOK(t, "http://"+listener.Addr().String()+"/")
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

func TestServerCloseDoesNotWaitForActiveRequest(t *testing.T) {
	container := NewContainer()
	requestStarted := make(chan struct{})
	releaseRequest := make(chan struct{})
	app, err := servlet.NewWebApp("immediate")
	if err != nil {
		t.Fatalf("NewWebApp failed: %v", err)
	}
	deployment, err := servletcontainer.NewDeployment(app,
		servletcontainer.WithMapping("/", servlet.HandlerFunc(func(_ context.Context, _ *servlet.Request, _ servlet.Response) error {
			close(requestStarted)
			<-releaseRequest
			return nil
		})),
	)
	if err != nil {
		t.Fatalf("NewDeployment failed: %v", err)
	}
	if _, err := container.Deploy(t.Context(), deployment); err != nil {
		t.Fatalf("Deploy failed: %v", err)
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Listen failed: %v", err)
	}
	server, err := NewServer(container)
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}
	serveErrCh := make(chan error, 1)
	go func() { serveErrCh <- server.Serve(context.Background(), listener) }()
	go func() { _, _ = http.Get("http://" + listener.Addr().String() + "/") }()
	select {
	case <-requestStarted:
	case <-time.After(time.Second):
		t.Fatal("request handler did not start")
	}
	started := time.Now()
	if err := server.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}
	if elapsed := time.Since(started); elapsed >= 250*time.Millisecond {
		t.Fatalf("Close elapsed = %s, want < 250ms", elapsed)
	}
	close(releaseRequest)
	select {
	case <-serveErrCh:
	case <-time.After(time.Second):
		t.Fatal("Serve did not return after Close")
	}
}

func requestUntilOK(t *testing.T, target string) string {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for {
		response, err := http.Get(target)
		if err == nil {
			data, readErr := io.ReadAll(response.Body)
			closeErr := response.Body.Close()
			if readErr != nil || closeErr != nil {
				t.Fatalf("read/close response = %v/%v", readErr, closeErr)
			}
			if response.StatusCode == http.StatusOK {
				return string(data)
			}
			t.Fatalf("status = %d, body = %q", response.StatusCode, string(data))
		}
		if time.Now().After(deadline) {
			t.Fatalf("GET %s did not succeed before deadline: %v", target, err)
		}
		time.Sleep(10 * time.Millisecond)
	}
}
