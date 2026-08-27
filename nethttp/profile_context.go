package nethttp

import (
	"context"
	"errors"

	"goark.dev/arkarta/servlet"
	"goark.dev/arkarta/servlet/async"
	servletcontainer "goark.dev/arkarta/servlet/container"
	"goark.dev/arkarta/servlet/multipart"
	"goark.dev/arkarta/servlet/session"
)

const (
	attributeProfileContext = "arkhos.nethttp.profile_context"
)

type profileContext struct {
	sessionAccessor *session.Accessor
	multipartParser *multipart.Parser
	securityPolicy  SecurityPolicy
	asyncOptions    []async.Option
}

type profileApplication struct {
	delegate servletcontainer.Application
	profile  *profileContext
}

func newProfileApplication(delegate servletcontainer.Application, cfg containerOptions) (servletcontainer.Application, error) {
	if delegate == nil {
		return nil, ErrNilApplication
	}
	profile := &profileContext{
		securityPolicy: cfg.securityPolicy,
		asyncOptions:   append([]async.Option(nil), cfg.asyncOptions...),
	}
	if cfg.sessionManagerFactory != nil {
		manager, err := cfg.sessionManagerFactory(delegate.WebApp())
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
	if cfg.multipartParser != nil {
		profile.multipartParser = cfg.multipartParser()
	}
	return &profileApplication{
		delegate: delegate,
		profile:  profile,
	}, nil
}

func (a *profileApplication) WebApp() *servlet.WebApp {
	if a == nil || a.delegate == nil {
		return nil
	}
	return a.delegate.WebApp()
}

func (a *profileApplication) Handler() servlet.Handler {
	if a == nil || a.delegate == nil {
		return servlet.HandlerFunc(func(context.Context, *servlet.Request, servlet.Response) error {
			return ErrNilApplication
		})
	}
	delegate := a.delegate.Handler()
	return servlet.HandlerFunc(func(ctx context.Context, req *servlet.Request, res servlet.Response) error {
		if req != nil {
			req.SetAttribute(attributeProfileContext, a.profile)
		}
		if a.profile != nil && a.profile.securityPolicy != nil {
			if err := a.profile.securityPolicy.Apply(ctx, req, res); err != nil {
				return errors.Join(err, cleanupMultipart(req))
			}
		}
		err := delegate.Serve(ctx, req, res)
		return errors.Join(err, cleanupMultipart(req))
	})
}

func (a *profileApplication) Stop(ctx context.Context) error {
	if a == nil || a.delegate == nil {
		return nil
	}
	return a.delegate.Stop(ctx)
}

func currentProfile(req *servlet.Request) (*profileContext, bool) {
	if req == nil {
		return nil, false
	}
	value, ok := req.Attribute(attributeProfileContext)
	if !ok {
		return nil, false
	}
	profile, ok := value.(*profileContext)
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
