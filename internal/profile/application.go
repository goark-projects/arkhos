package profile

import (
	"context"
	"errors"

	"goark.dev/arkarta/servlet"
	"goark.dev/arkarta/servlet/async"
	servletcontainer "goark.dev/arkarta/servlet/container"
	"goark.dev/arkarta/servlet/multipart"
	"goark.dev/arkarta/servlet/session"
)

const attributeContext = "arkhos.profile.context"
const attributeAsyncContext = "arkhos.profile.async_context"

type requestContext struct {
	sessionAccessor *session.Accessor
	multipartParser *multipart.Parser
	securityPolicy  SecurityPolicy
	asyncOptions    []async.Option
}

type application struct {
	delegate servletcontainer.Application
	profile  *requestContext
}

// Decorate 为应用绑定容器拥有的可选 Profile 运行时。
func Decorate(delegate servletcontainer.Application, cfg Config) (servletcontainer.Application, error) {
	if delegate == nil {
		return nil, ErrNilApplication
	}
	profile := &requestContext{
		securityPolicy: cfg.SecurityPolicy,
		asyncOptions:   append([]async.Option(nil), cfg.AsyncOptions...),
	}
	if cfg.SessionManagerFactory != nil {
		manager, err := cfg.SessionManagerFactory(delegate.WebApp())
		if err != nil {
			return nil, err
		}
		if manager != nil {
			accessor, err := session.NewAccessorForWebApp(manager, delegate.WebApp())
			if err != nil {
				return nil, err
			}
			profile.sessionAccessor = accessor
		}
	}
	if cfg.MultipartParser != nil {
		profile.multipartParser = cfg.MultipartParser()
	}
	return &application{delegate: delegate, profile: profile}, nil
}

func (a *application) WebApp() *servlet.WebApp {
	if a == nil || a.delegate == nil {
		return nil
	}
	return a.delegate.WebApp()
}

func (a *application) Handler() servlet.Handler {
	if a == nil || a.delegate == nil {
		return servlet.HandlerFunc(func(context.Context, *servlet.Request, servlet.Response) error {
			return ErrNilApplication
		})
	}
	delegate := a.delegate.Handler()
	return servlet.HandlerFunc(func(ctx context.Context, req *servlet.Request, res servlet.Response) error {
		if req != nil {
			req.SetAttribute(attributeContext, a.profile)
		}
		if a.profile != nil && a.profile.securityPolicy != nil {
			if err := a.profile.securityPolicy.Apply(ctx, req, res); err != nil {
				return errors.Join(err, cleanupMultipart(req))
			}
		}
		return errors.Join(delegate.Serve(ctx, req, res), awaitAsync(req), cleanupMultipart(req))
	})
}

func awaitAsync(req *servlet.Request) error {
	if req == nil {
		return nil
	}
	value, ok := req.Attribute(attributeAsyncContext)
	if !ok {
		return nil
	}
	asyncContext, ok := value.(*async.Context)
	if !ok || asyncContext == nil {
		return nil
	}
	return errors.Join(
		asyncContext.Await(context.Background()),
		asyncContext.AwaitQuiescence(context.Background()),
	)
}

func (a *application) Stop(ctx context.Context) error {
	if a == nil || a.delegate == nil {
		return nil
	}
	return a.delegate.Stop(ctx)
}

func current(req *servlet.Request) (*requestContext, bool) {
	if req == nil {
		return nil, false
	}
	value, ok := req.Attribute(attributeContext)
	if !ok {
		return nil, false
	}
	profile, ok := value.(*requestContext)
	return profile, ok && profile != nil
}

func cleanupMultipart(req *servlet.Request) error {
	form, ok := multipart.Current(req)
	if !ok {
		return nil
	}
	req.SetAttribute(multipart.AttributeForm, nil)
	return form.RemoveAll()
}
