package hertz

import (
	"bytes"
	"context"
	"io"

	"github.com/cloudwego/hertz/pkg/app"

	"goark.dev/arkarta/servlet"
)

func newRequest(ctx context.Context, requestContext *app.RequestContext, contextPath string, options ...servlet.RequestOption) (*servlet.Request, error) {
	request := &requestContext.Request
	uri := request.URI()
	var body io.Reader
	if request.IsBodyStream() {
		body = request.BodyStream()
	} else {
		body = bytes.NewReader(request.Body())
	}
	input := &servlet.RequestInput{
		Context:       ctx,
		Method:        string(request.Header.Method()),
		Protocol:      request.Header.GetProtocol(),
		Scheme:        string(uri.Scheme()),
		Host:          string(uri.Host()),
		RequestURI:    string(uri.RequestURI()),
		Path:          string(uri.Path()),
		QueryString:   string(uri.QueryString()),
		ContextPath:   contextPath,
		Header:        requestHeader{header: &request.Header},
		Body:          io.NopCloser(body),
		ContentLength: int64(request.Header.ContentLength()),
		RemoteAddr:    requestContext.RemoteAddr().String(),
		Trailer:       servlet.NewHeader(),
	}
	if conn := requestContext.GetConn(); conn != nil && conn.LocalAddr() != nil {
		input.LocalAddr = conn.LocalAddr().String()
	}
	return servlet.NewRequestFromInput(input, options...)
}
