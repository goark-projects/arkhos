package hertz

import (
	"io"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"

	"goark.dev/arkarta/servlet"
)

type response struct {
	ctx       *app.RequestContext
	header    responseHeader
	status    int
	committed bool
}

func newResponse(ctx *app.RequestContext) *response {
	return &response{
		ctx:    ctx,
		header: responseHeader{header: &ctx.Response.Header},
		status: consts.StatusOK,
	}
}

func (r *response) Header() servlet.Header {
	return r.header
}

func (r *response) SetStatus(code int) {
	if r.committed {
		return
	}
	if code < 100 || code > 999 {
		code = consts.StatusInternalServerError
	}
	r.status = code
}

func (r *response) Status() int {
	return r.status
}

func (r *response) Write(data []byte) (int, error) {
	r.commit()
	return r.ctx.Write(data)
}

func (r *response) WriteString(value string) (int, error) {
	r.commit()
	return r.ctx.WriteString(value)
}

func (r *response) Flush() error {
	r.commit()
	return r.ctx.Flush()
}

func (r *response) Committed() bool {
	return r.committed
}

func (r *response) Reset() error {
	if r.committed {
		return servlet.ErrResponseCommitted
	}
	r.ctx.Response.Reset()
	r.status = consts.StatusOK
	return nil
}

func (r *response) BodyWriter() io.Writer {
	return r
}

func (r *response) finish() {
	r.commit()
}

func (r *response) commit() {
	if r.committed {
		return
	}
	r.committed = true
	r.ctx.SetStatusCode(r.status)
}
