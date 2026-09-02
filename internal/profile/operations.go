package profile

import (
	"context"

	"goark.dev/arkarta/servlet"
	"goark.dev/arkarta/servlet/async"
	"goark.dev/arkarta/servlet/multipart"
	"goark.dev/arkarta/servlet/session"
)

func GetSession(ctx context.Context, req *servlet.Request, res servlet.Response, create bool) (session.Session, bool, error) {
	profile, ok := current(req)
	if !ok || profile.sessionAccessor == nil {
		return nil, false, ErrSessionUnavailable
	}
	return profile.sessionAccessor.Get(ctx, req, res, create)
}

func RequestedSessionID(req *servlet.Request) (string, bool) {
	profile, ok := current(req)
	if !ok || profile.sessionAccessor == nil {
		return "", false
	}
	return profile.sessionAccessor.RequestedID(req)
}

func RequestedSessionIDValid(ctx context.Context, req *servlet.Request) (bool, error) {
	profile, ok := current(req)
	if !ok || profile.sessionAccessor == nil {
		return false, ErrSessionUnavailable
	}
	return profile.sessionAccessor.RequestedIDValid(ctx, req)
}

func ChangeSessionID(ctx context.Context, req *servlet.Request, res servlet.Response) (string, error) {
	profile, ok := current(req)
	if !ok || profile.sessionAccessor == nil {
		return "", ErrSessionUnavailable
	}
	return profile.sessionAccessor.ChangeID(ctx, req, res)
}

func EncodeSessionURL(req *servlet.Request, rawURL string) (string, error) {
	profile, ok := current(req)
	if !ok || profile.sessionAccessor == nil {
		return rawURL, ErrSessionUnavailable
	}
	return profile.sessionAccessor.EncodeURL(req, rawURL)
}

func EncodeSessionRedirectURL(req *servlet.Request, rawURL string) (string, error) {
	profile, ok := current(req)
	if !ok || profile.sessionAccessor == nil {
		return rawURL, ErrSessionUnavailable
	}
	return profile.sessionAccessor.EncodeRedirectURL(req, rawURL)
}

func ParseMultipart(req *servlet.Request) (*multipart.Form, error) {
	profile, ok := current(req)
	if !ok || profile.multipartParser == nil {
		return nil, ErrMultipartUnavailable
	}
	return multipart.ParseRequest(req, profile.multipartParser)
}

func MultipartPart(req *servlet.Request, name string) (multipart.Part, bool, error) {
	profile, ok := current(req)
	if !ok || profile.multipartParser == nil {
		return multipart.Part{}, false, ErrMultipartUnavailable
	}
	return multipart.RequestPart(req, name, profile.multipartParser)
}

func MultipartParts(req *servlet.Request) ([]multipart.Part, error) {
	profile, ok := current(req)
	if !ok || profile.multipartParser == nil {
		return nil, ErrMultipartUnavailable
	}
	return multipart.Parts(req, profile.multipartParser)
}

func StartAsync(ctx context.Context, req *servlet.Request, res servlet.Response, options ...async.Option) (*async.Context, error) {
	profile, _ := current(req)
	merged := make([]async.Option, 0, len(options))
	if profile != nil {
		merged = append(merged, profile.asyncOptions...)
	}
	merged = append(merged, options...)
	asyncContext, err := async.NewContext(ctx, req, res, merged...)
	if err != nil {
		return nil, err
	}
	req.SetAttribute(attributeAsyncContext, asyncContext)
	return asyncContext, nil
}
