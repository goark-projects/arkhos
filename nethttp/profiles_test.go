package nethttp

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"goark.dev/arkarta/servlet"
	servletasync "goark.dev/arkarta/servlet/async"
	servletcontainer "goark.dev/arkarta/servlet/container"
	servletmultipart "goark.dev/arkarta/servlet/multipart"
	"goark.dev/arkarta/servlet/security"
	servletsession "goark.dev/arkarta/servlet/session"
	"goark.dev/arkarta/servlet/tck"
	"goark.dev/arkarta/servlet/upgrade"
	websockettck "goark.dev/arkarta/websocket/tck"
)

func TestContainerMetadataIncludesImplementedProfiles(t *testing.T) {
	metadata := NewContainer().Metadata()
	for _, profile := range []servletcontainer.Profile{
		servletcontainer.ProfileCore,
		servletcontainer.ProfileSession,
		servletcontainer.ProfileMultipart,
		servletcontainer.ProfileAsyncStream,
		servletcontainer.ProfileUpgrade,
		servletcontainer.ProfileNativeIO,
		profileSecurity,
	} {
		if !metadata.Supports(profile) {
			t.Fatalf("metadata should support %q", profile)
		}
	}
}

func TestContainerDeploysAllImplementedProfiles(t *testing.T) {
	app, err := servlet.NewWebApp("profiles")
	if err != nil {
		t.Fatalf("NewWebApp failed: %v", err)
	}
	deployment, err := servletcontainer.NewDeployment(app,
		servletcontainer.WithProfile(servletcontainer.ProfileSession),
		servletcontainer.WithProfile(servletcontainer.ProfileMultipart),
		servletcontainer.WithProfile(servletcontainer.ProfileAsyncStream),
		servletcontainer.WithProfile(servletcontainer.ProfileUpgrade),
		servletcontainer.WithProfile(profileSecurity),
		servletcontainer.WithMapping("/", servlet.HandlerFunc(func(context.Context, *servlet.Request, servlet.Response) error {
			return nil
		})),
	)
	if err != nil {
		t.Fatalf("NewDeployment failed: %v", err)
	}
	if _, err := NewContainer().Deploy(context.Background(), deployment); err != nil {
		t.Fatalf("Deploy failed: %v", err)
	}
}

func TestSessionProfileBindsCookieBackedSession(t *testing.T) {
	container := NewContainer()
	bodyCh := make(chan string, 2)
	handler := servlet.HandlerFunc(func(ctx context.Context, req *servlet.Request, res servlet.Response) error {
		current, ok, err := GetSession(ctx, req, res, true)
		if err != nil {
			return err
		}
		if !ok || current == nil {
			return errors.New("session should be available")
		}
		bodyCh <- current.ID()
		_, err = res.WriteString(current.ID())
		return err
	})
	deployProfileApp(t, container, "sessions", handler, servletcontainer.ProfileSession)
	if err := container.Start(context.Background()); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	first := httptest.NewRecorder()
	container.Handler().ServeHTTP(first, httptest.NewRequest(http.MethodGet, "/", nil))
	firstResult := first.Result()
	defer firstResult.Body.Close()
	cookies := firstResult.Cookies()
	if len(cookies) != 1 || cookies[0].Name != servletsession.DefaultCookieName {
		t.Fatalf("cookies = %#v, want session cookie", cookies)
	}
	firstID := <-bodyCh

	second := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.AddCookie(cookies[0])
	container.Handler().ServeHTTP(second, request)
	secondID := <-bodyCh

	if firstID == "" || secondID != firstID {
		t.Fatalf("session ids = %q/%q, want stable cookie-backed id", firstID, secondID)
	}
}

func TestMultipartProfileParsesRequestWithContainerParser(t *testing.T) {
	container := NewContainer(WithMultipartConfig(servletmultipart.NewConfig(
		servletmultipart.WithLocation(t.TempDir()),
		servletmultipart.WithMaxRequestSize(1<<20),
	)))
	handler := servlet.HandlerFunc(func(_ context.Context, req *servlet.Request, res servlet.Response) error {
		form, err := ParseMultipart(req)
		if err != nil {
			return err
		}
		part, ok := form.Part("file")
		if !ok || part.SubmittedFileName() != "payload.txt" {
			return errors.New("multipart file part should be available")
		}
		_, err = res.WriteString(form.Value("field") + ":" + part.SubmittedFileName())
		return err
	})
	deployProfileApp(t, container, "multipart", handler, servletcontainer.ProfileMultipart)
	if err := container.Start(context.Background()); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	body, contentType := newMultipartBody(t)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/upload", body)
	request.Header.Set("Content-Type", contentType)
	container.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK || recorder.Body.String() != "value:payload.txt" {
		t.Fatalf("response = %d/%q, want multipart data", recorder.Code, recorder.Body.String())
	}
}

func TestMultipartProfileCleansUpWhenSecurityPolicyStopsRequest(t *testing.T) {
	location := t.TempDir()
	container := NewContainer(
		WithMultipartConfig(servletmultipart.NewConfig(servletmultipart.WithLocation(location))),
		WithSecurityPolicy(SecurityPolicyFunc(func(_ context.Context, req *servlet.Request, _ servlet.Response) error {
			if _, err := ParseMultipart(req); err != nil {
				return err
			}
			return servlet.NewHTTPError(http.StatusForbidden, "blocked", nil)
		})),
	)
	handler := servlet.HandlerFunc(func(context.Context, *servlet.Request, servlet.Response) error {
		return errors.New("handler should not run after security denial")
	})
	deployProfileApp(t, container, "multipart-security", handler, servletcontainer.ProfileMultipart, profileSecurity)
	if err := container.Start(context.Background()); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	body, contentType := newMultipartBody(t)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/upload", body)
	request.Header.Set("Content-Type", contentType)
	container.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("response = %d, want 403", recorder.Code)
	}
	entries, err := os.ReadDir(location)
	if err != nil {
		t.Fatalf("ReadDir failed: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("multipart temp files = %d, want 0", len(entries))
	}
}

func TestSecurityProfileAppliesContainerPolicy(t *testing.T) {
	realm := security.NewStaticRealm(security.WithStaticUser("alice", "secret", "orders"))
	container := NewContainer(WithSecurityPolicy(BasicSecurityPolicy(
		security.NewBasicAuthenticator(realm, security.WithBasicRealmName("arkhos")),
		security.NewConstraint(security.WithRoles("orders")),
	)))
	handler := servlet.HandlerFunc(func(_ context.Context, req *servlet.Request, res servlet.Response) error {
		_, err := res.WriteString(security.RemoteUser(req))
		return err
	})
	deployProfileApp(t, container, "security", handler, profileSecurity)
	if err := container.Start(context.Background()); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	denied := httptest.NewRecorder()
	container.Handler().ServeHTTP(denied, httptest.NewRequest(http.MethodGet, "/secure", nil))
	if denied.Code != http.StatusUnauthorized || denied.Header().Get("WWW-Authenticate") != `Basic realm="arkhos"` {
		t.Fatalf("denied response = %d/%q", denied.Code, denied.Header().Get("WWW-Authenticate"))
	}

	wrongPassword := httptest.NewRecorder()
	wrongRequest := httptest.NewRequest(http.MethodGet, "/secure", nil)
	wrongRequest.SetBasicAuth("alice", "bad")
	container.Handler().ServeHTTP(wrongPassword, wrongRequest)
	if wrongPassword.Code != http.StatusUnauthorized {
		t.Fatalf("wrong password response = %d, want 401", wrongPassword.Code)
	}

	allowed := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/secure", nil)
	request.SetBasicAuth("alice", "secret")
	container.Handler().ServeHTTP(allowed, request)
	if allowed.Code != http.StatusOK || allowed.Body.String() != "alice" {
		t.Fatalf("allowed response = %d/%q", allowed.Code, allowed.Body.String())
	}
}

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
	handler := servlet.HandlerFunc(func(ctx context.Context, req *servlet.Request, res servlet.Response) error {
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
	})
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
	handler := servlet.HandlerFunc(func(ctx context.Context, req *servlet.Request, res servlet.Response) error {
		return UpgradeHTTP(ctx, req, res, upgrade.HandlerFunc(func(_ context.Context, conn upgrade.Connection) error {
			_, writeErr := conn.Write([]byte("upgraded\n"))
			closeErr := conn.Close()
			return errors.Join(writeErr, closeErr)
		}))
	})
	server := httptest.NewServer(Handler(handler))
	defer server.Close()

	conn, err := dialHTTPServer(server.URL)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	defer conn.Close()
	if _, err := io.WriteString(conn, "GET /upgrade HTTP/1.1\r\nHost: example.com\r\nConnection: Upgrade\r\nUpgrade: arkhos-test\r\n\r\n"); err != nil {
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

func deployProfileApp(t *testing.T, container *Container, name string, handler servlet.Handler, profiles ...servletcontainer.Profile) {
	t.Helper()
	app, err := servlet.NewWebApp(name)
	if err != nil {
		t.Fatalf("NewWebApp failed: %v", err)
	}
	options := make([]servletcontainer.DeploymentOption, 0, len(profiles)+1)
	for _, profile := range profiles {
		options = append(options, servletcontainer.WithProfile(profile))
	}
	options = append(options, servletcontainer.WithMapping("/", handler))
	deployment, err := servletcontainer.NewDeployment(app, options...)
	if err != nil {
		t.Fatalf("NewDeployment failed: %v", err)
	}
	if _, err := container.Deploy(context.Background(), deployment); err != nil {
		t.Fatalf("Deploy failed: %v", err)
	}
}

func newMultipartBody(t *testing.T) (*bytes.Buffer, string) {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if err := writer.WriteField("field", "value"); err != nil {
		t.Fatalf("WriteField failed: %v", err)
	}
	part, err := writer.CreateFormFile("file", "payload.txt")
	if err != nil {
		t.Fatalf("CreateFormFile failed: %v", err)
	}
	if _, err := part.Write([]byte("payload")); err != nil {
		t.Fatalf("part write failed: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("writer close failed: %v", err)
	}
	return &body, writer.FormDataContentType()
}

func dialHTTPServer(rawURL string) (net.Conn, error) {
	address := strings.TrimPrefix(rawURL, "http://")
	address = strings.TrimPrefix(address, "https://")
	return net.DialTimeout("tcp", address, 3*time.Second)
}
