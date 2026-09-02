package hertz

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"mime/multipart"
	"strings"
	"testing"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"

	"goark.dev/arkarta/servlet"
	servletasync "goark.dev/arkarta/servlet/async"
	servletcontainer "goark.dev/arkarta/servlet/container"
	servletmultipart "goark.dev/arkarta/servlet/multipart"
	"goark.dev/arkarta/servlet/security"
)

func TestContainerMetadataDeclaresVerifiedProfiles(t *testing.T) {
	metadata := NewContainer().Metadata()
	for _, profile := range []servletcontainer.Profile{
		servletcontainer.ProfileCore,
		servletcontainer.ProfileSession,
		servletcontainer.ProfileMultipart,
		servletcontainer.ProfileAsyncStream,
		servletcontainer.ProfileNativeIO,
		profileSecurity,
	} {
		if !metadata.Supports(profile) {
			t.Fatalf("metadata should support %q", profile)
		}
	}
	if metadata.Supports(servletcontainer.ProfileUpgrade) {
		t.Fatal("upgrade must not be declared before Hertz upgrade tests pass")
	}
}

func TestSessionProfileUsesHertzCookieHeaders(t *testing.T) {
	container := NewContainer()
	deployHertzProfile(t, container, servlet.HandlerFunc(func(ctx context.Context, req *servlet.Request, res servlet.Response) error {
		current, ok, err := GetSession(ctx, req, res, true)
		if err != nil || !ok {
			return errors.Join(err, errors.New("session unavailable"))
		}
		_, err = res.WriteString(current.ID())
		return err
	}), servletcontainer.ProfileSession)

	first := newHertzRequest("GET", "/")
	container.Handler()(t.Context(), first)
	firstID := string(first.Response.Body())
	setCookie := string(first.Response.Header.Peek("Set-Cookie"))
	if firstID == "" || setCookie == "" {
		t.Fatalf("session response = %q/%q", firstID, setCookie)
	}

	second := newHertzRequest("GET", "/")
	second.Request.Header.Set("Cookie", strings.Split(setCookie, ";")[0])
	container.Handler()(t.Context(), second)
	if secondID := string(second.Response.Body()); secondID != firstID {
		t.Fatalf("session ids = %q/%q", firstID, secondID)
	}
}

func TestMultipartProfileUsesHertzBody(t *testing.T) {
	container := NewContainer(WithMultipartConfig(servletmultipart.NewConfig(servletmultipart.WithMaxRequestSize(1 << 20))))
	deployHertzProfile(t, container, servlet.HandlerFunc(func(_ context.Context, req *servlet.Request, res servlet.Response) error {
		form, err := ParseMultipart(req)
		if err != nil {
			return err
		}
		_, err = res.WriteString(form.Value("field"))
		return err
	}), servletcontainer.ProfileMultipart)

	body, contentType := hertzMultipartBody(t)
	ctx := newHertzRequest("POST", "/upload")
	ctx.Request.Header.Set("Content-Type", contentType)
	ctx.Request.SetBody(body)
	container.Handler()(t.Context(), ctx)
	if ctx.Response.StatusCode() != consts.StatusOK || string(ctx.Response.Body()) != "value" {
		t.Fatalf("response = %d/%q", ctx.Response.StatusCode(), ctx.Response.Body())
	}
}

func TestContainerAppliesFormBodyLimit(t *testing.T) {
	container := NewContainer(WithMaxFormBodySize(4))
	deployHertzProfile(t, container, servlet.HandlerFunc(func(_ context.Context, req *servlet.Request, res servlet.Response) error {
		if err := req.ParseParameters(); errors.Is(err, servlet.ErrFormBodyTooLarge) {
			_, writeErr := res.WriteString("limited")
			return writeErr
		} else if err != nil {
			return err
		}
		_, err := res.WriteString("accepted")
		return err
	}))

	ctx := newHertzRequest("POST", "/form")
	ctx.Request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	ctx.Request.SetBodyString("field=value")
	container.Handler()(t.Context(), ctx)
	if ctx.Response.StatusCode() != consts.StatusOK || string(ctx.Response.Body()) != "limited" {
		t.Fatalf("response = %d/%q, want 200/limited", ctx.Response.StatusCode(), ctx.Response.Body())
	}
}

func TestSecurityProfileUsesHertzAuthorizationHeader(t *testing.T) {
	realm := security.NewStaticRealm(security.WithStaticUser("alice", "secret", "orders"))
	container := NewContainer(WithSecurityPolicy(BasicSecurityPolicy(
		security.NewBasicAuthenticator(realm),
		security.NewConstraint(security.WithRoles("orders")),
	)))
	deployHertzProfile(t, container, servlet.HandlerFunc(func(_ context.Context, req *servlet.Request, res servlet.Response) error {
		_, err := res.WriteString(security.RemoteUser(req))
		return err
	}), profileSecurity)

	denied := newHertzRequest("GET", "/secure")
	container.Handler()(t.Context(), denied)
	if denied.Response.StatusCode() != consts.StatusUnauthorized {
		t.Fatalf("denied status = %d", denied.Response.StatusCode())
	}

	allowed := newHertzRequest("GET", "/secure")
	credentials := base64.StdEncoding.EncodeToString([]byte("alice:secret"))
	allowed.Request.Header.Set("Authorization", "Basic "+credentials)
	container.Handler()(t.Context(), allowed)
	if allowed.Response.StatusCode() != consts.StatusOK || string(allowed.Response.Body()) != "alice" {
		t.Fatalf("allowed response = %d/%q", allowed.Response.StatusCode(), allowed.Response.Body())
	}
}

func TestAsyncProfileWaitsWithoutExilingHertzContext(t *testing.T) {
	container := NewContainer()
	deployHertzProfile(t, container, servlet.HandlerFunc(func(ctx context.Context, req *servlet.Request, res servlet.Response) error {
		asyncContext, err := StartAsync(ctx, req, res)
		if err != nil {
			return err
		}
		asyncContext.Go(func(ctx context.Context) error {
			stream, err := NewAsyncStream(res)
			if err != nil {
				return err
			}
			_, err = stream.Write(ctx, []byte("async"))
			return err
		})
		return nil
	}), servletcontainer.ProfileAsyncStream)

	ctx := newHertzRequest("GET", "/async")
	container.Handler()(t.Context(), ctx)
	if string(ctx.Response.Body()) != "async" || ctx.IsExiled() {
		t.Fatalf("async response = %q, exiled = %v", ctx.Response.Body(), ctx.IsExiled())
	}
}

func TestAsyncProfileWaitsForTimedOutWorkerBeforeReturning(t *testing.T) {
	container := NewContainer()
	started := make(chan struct{})
	release := make(chan struct{})
	deployHertzProfile(t, container, servlet.HandlerFunc(func(ctx context.Context, req *servlet.Request, res servlet.Response) error {
		asyncContext, err := StartAsync(ctx, req, res, servletasync.WithTimeout(time.Millisecond))
		if err != nil {
			return err
		}
		return asyncContext.Go(func(context.Context) error {
			close(started)
			<-release
			return nil
		})
	}), servletcontainer.ProfileAsyncStream)

	ctx := newHertzRequest("GET", "/async-timeout")
	done := make(chan struct{})
	go func() {
		container.Handler()(t.Context(), ctx)
		close(done)
	}()
	<-started
	select {
	case <-done:
		t.Fatal("handler returned before the timed-out worker exited")
	case <-time.After(20 * time.Millisecond):
	}
	close(release)
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("handler did not return after the worker exited")
	}
	if ctx.IsExiled() {
		t.Fatal("completed async request must remain reusable")
	}
}

func deployHertzProfile(t *testing.T, container *Container, handler servlet.Handler, profiles ...servletcontainer.Profile) {
	t.Helper()
	webApp, err := servlet.NewWebApp("profiles")
	if err != nil {
		t.Fatalf("NewWebApp failed: %v", err)
	}
	options := make([]servletcontainer.DeploymentOption, 0, len(profiles)+1)
	for _, profile := range profiles {
		options = append(options, servletcontainer.WithProfile(profile))
	}
	options = append(options, servletcontainer.WithMapping("/", handler))
	deployment, err := servletcontainer.NewDeployment(webApp, options...)
	if err != nil {
		t.Fatalf("NewDeployment failed: %v", err)
	}
	if _, err := container.Deploy(t.Context(), deployment); err != nil {
		t.Fatalf("Deploy failed: %v", err)
	}
	if err := container.Start(t.Context()); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
}

func newHertzRequest(method, target string) *app.RequestContext {
	ctx := app.NewContext(0)
	ctx.Request.Header.SetMethod(method)
	ctx.Request.SetRequestURI(target)
	return ctx
}

func hertzMultipartBody(t *testing.T) ([]byte, string) {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if err := writer.WriteField("field", "value"); err != nil {
		t.Fatalf("WriteField failed: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}
	return body.Bytes(), writer.FormDataContentType()
}
