package hertz

import (
	"context"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"goark.dev/arkarta/servlet"
	servletcontainer "goark.dev/arkarta/servlet/container"
)

func TestServerCloseDoesNotWaitForActiveRequest(t *testing.T) {
	container := NewContainer()
	requestStarted := make(chan struct{})
	releaseRequest := make(chan struct{})
	app, err := servlet.NewWebApp("immediate")
	if err != nil {
		t.Fatalf("NewWebApp failed: %v", err)
	}
	deployment, err := servletcontainer.NewDeployment(
		app,
		servletcontainer.WithMapping(
			"/",
			servlet.HandlerFunc(
				func(_ context.Context, _ *servlet.Request, _ servlet.Response) error {
					close(requestStarted)
					<-releaseRequest
					return nil
				},
			),
		),
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
