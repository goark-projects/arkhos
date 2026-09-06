package nethttp

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"goark.dev/arkarta/servlet"
	servletasync "goark.dev/arkarta/servlet/async"
	servletcontainer "goark.dev/arkarta/servlet/container"
	servletsession "goark.dev/arkarta/servlet/session"
	"goark.dev/arkarta/servlet/tck"
	"goark.dev/arkarta/servlet/upgrade"
	websockettck "goark.dev/arkarta/websocket/tck"
)

func TestAsyncProfileStartsContextWithContainerOptions(t *testing.T) {
	events := make(chan string, 3)
	container := NewContainer(WithAsyncOptions(servletasync.WithListener(servletasync.ListenerFunc{
		Start: func(context.Context, servletasync.Event) {
			events <- "start"
		},
		Complete: func(context.Context, servletasync.Event) {
			events <- "complete"
		},
	})))
	handler := servlet.HandlerFunc(
		func(ctx context.Context, req *servlet.Request, res servlet.Response) error {
			asyncCtx, err := StartAsync(ctx, req, res)
			if err != nil {
				return err
			}
			asyncCtx.Go(func(ctx context.Context) error {
				stream, err := NewAsyncStream(res)
				if err != nil {
					return err
				}
				if _, err := stream.Write(ctx, []byte("async")); err != nil {
					return err
				}
				events <- "done"
				return stream.Close(ctx)
			})
			return asyncCtx.Await(context.Background())
		},
	)
	deployProfileApp(t, container, "async", handler, servletcontainer.ProfileAsyncStream)
	if err := container.Start(context.Background()); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	recorder := httptest.NewRecorder()
	container.Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/async", nil))
	if recorder.Code != http.StatusOK || recorder.Body.String() != "async" {
		t.Fatalf("response = %d/%q, want async body", recorder.Code, recorder.Body.String())
	}
	for _, want := range []string{"start", "done", "complete"} {
		if got := <-events; got != want {
			t.Fatalf("async event = %q, want %q", got, want)
		}
	}
}

func TestUpgradeProfileDelegatesConnection(t *testing.T) {
	handler := servlet.HandlerFunc(
		func(ctx context.Context, req *servlet.Request, res servlet.Response) error {
			return UpgradeHTTP(
				ctx,
				req,
				res,
				upgrade.HandlerFunc(func(_ context.Context, conn upgrade.Connection) error {
					_, writeErr := conn.Write([]byte("upgraded\n"))
					closeErr := conn.Close()
					return errors.Join(writeErr, closeErr)
				}),
			)
		},
	)
	server := httptest.NewServer(Handler(handler))
	defer server.Close()

	conn, err := dialHTTPServer(server.URL)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	defer conn.Close()
	request := "GET /upgrade HTTP/1.1\r\n" +
		"Host: example.com\r\n" +
		"Connection: Upgrade\r\n" +
		"Upgrade: arkhos-test\r\n\r\n"
	if _, err := io.WriteString(conn, request); err != nil {
		t.Fatalf("write request failed: %v", err)
	}
	data, err := io.ReadAll(conn)
	if err != nil {
		t.Fatalf("read upgraded data failed: %v", err)
	}
	if !strings.Contains(string(data), "upgraded") {
		t.Fatalf("upgrade data = %q, want upgraded payload", string(data))
	}
}

func TestWebSocketProfileRunsArkartaTCK(t *testing.T) {
	websockettck.RunHandshake(t, NewWebSocketHandshaker)
	websockettck.RunFrameCodec(t)
	websockettck.RunCompression(t, NewPerMessageDeflate)
	websockettck.RunEndpointLifecycle(t, NewWebSocketSession)
}

func TestServletProfileTCKs(t *testing.T) {
	tck.RunSessionManager(t, func() servletsession.Manager {
		return NewSessionManager()
	})
	tck.RunMemorySessionProfile(t, NewMemorySessionManager)
	tck.RunSessionRequestBinding(t, func() servletsession.Manager {
		return NewSessionManager()
	})
	tck.RunMultipartParser(t, NewMultipartParser)
	tck.RunAsyncLifecycle(t)
	tck.RunSecurity(t)
}
