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
	if server.Address() != "127.0.0.1:9090" {
		t.Fatalf("address = %q, want 127.0.0.1:9090", server.Address())
	}
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
	if !server.Running() {
		t.Fatal("server should report running after serving a request")
	}
	cancel()
	if err := <-errCh; !errors.Is(err, context.Canceled) {
		t.Fatalf("Serve err = %v, want context.Canceled", err)
	}
	if app.State() != servlet.WebAppStateDestroyed {
		t.Fatalf("web app state = %v, want destroyed", app.State())
	}
}

func TestServerShutdownClosesIdleKeepAliveConnections(t *testing.T) {
	container := NewContainer()
	app, err := servlet.NewWebApp("keep-alive")
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
	if _, err := container.Deploy(t.Context(), deployment); err != nil {
		t.Fatalf("Deploy failed: %v", err)
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Listen failed: %v", err)
	}
	defer listener.Close()
	server, err := NewServer(container)
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}
	errCh := make(chan error, 1)
	go func() {
		errCh <- server.Serve(context.Background(), listener)
	}()

	transport := http.DefaultTransport.(*http.Transport).Clone()
	client := &http.Client{Transport: transport, Timeout: time.Second}
	response, err := client.Get("http://" + listener.Addr().String() + "/")
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}
	if _, err := io.Copy(io.Discard, response.Body); err != nil {
		t.Fatalf("read response failed: %v", err)
	}
	if err := response.Body.Close(); err != nil {
		t.Fatalf("close response failed: %v", err)
	}
	defer transport.CloseIdleConnections()

	started := time.Now()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		t.Fatalf("Shutdown failed: %v", err)
	}
	if elapsed := time.Since(started); elapsed >= 250*time.Millisecond {
		t.Fatalf("Shutdown elapsed = %s, want < 250ms", elapsed)
	}
	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("Serve failed: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Serve did not return after shutdown")
	}
}

func TestServerShutdownWaitsForActiveRequest(t *testing.T) {
	container := NewContainer()
	requestStarted := make(chan struct{})
	releaseRequest := make(chan struct{})
	app, err := servlet.NewWebApp("graceful")
	if err != nil {
		t.Fatalf("NewWebApp failed: %v", err)
	}
	deployment, err := servletcontainer.NewDeployment(app,
		servletcontainer.WithMapping("/", servlet.HandlerFunc(func(_ context.Context, _ *servlet.Request, res servlet.Response) error {
			close(requestStarted)
			<-releaseRequest
			_, err := res.WriteString("completed")
			return err
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
	defer listener.Close()
	server, err := NewServer(container)
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}
	serveErrCh := make(chan error, 1)
	go func() {
		serveErrCh <- server.Serve(context.Background(), listener)
	}()
	responseCh := make(chan string, 1)
	requestErrCh := make(chan error, 1)
	go func() {
		response, requestErr := http.Get("http://" + listener.Addr().String() + "/")
		if requestErr != nil {
			requestErrCh <- requestErr
			return
		}
		defer response.Body.Close()
		body, readErr := io.ReadAll(response.Body)
		if readErr != nil {
			requestErrCh <- readErr
			return
		}
		responseCh <- string(body)
	}()
	select {
	case <-requestStarted:
	case <-time.After(time.Second):
		t.Fatal("request handler did not start")
	}

	shutdownErrCh := make(chan error, 1)
	go func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		shutdownErrCh <- server.Shutdown(shutdownCtx)
	}()
	select {
	case err := <-shutdownErrCh:
		t.Fatalf("Shutdown returned before active request completed: %v", err)
	case <-time.After(25 * time.Millisecond):
	}
	close(releaseRequest)
	select {
	case err := <-requestErrCh:
		t.Fatalf("request failed during shutdown: %v", err)
	case body := <-responseCh:
		if body != "completed" {
			t.Fatalf("body = %q, want completed", body)
		}
	case <-time.After(time.Second):
		t.Fatal("request did not complete during shutdown")
	}
	if err := <-shutdownErrCh; err != nil {
		t.Fatalf("Shutdown failed: %v", err)
	}
	if err := <-serveErrCh; err != nil {
		t.Fatalf("Serve failed: %v", err)
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
