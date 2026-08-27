package nethttp

import (
	"testing"

	"goark.dev/arkarta/servlet/tck"
)

func TestHandlerCoreHTTP(t *testing.T) {
	tck.RunCoreHTTP(t, Handler)
}
