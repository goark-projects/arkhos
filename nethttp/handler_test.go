package nethttp

import (
	"testing"

	"goark.dev/arkarta/servlet/tck"
	tcknethttp "goark.dev/arkarta/servlet/tck/nethttp"
)

func TestHandlerCoreHTTP(t *testing.T) {
	tck.RunCore(t, tcknethttp.NewDriver(Handler))
}
