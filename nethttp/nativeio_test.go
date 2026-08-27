package nethttp

import (
	"testing"

	"goark.dev/arkarta/servlet/nativeio"
	"goark.dev/arkarta/servlet/tck"
)

func TestContainerNativeIO(t *testing.T) {
	tck.RunNativeIO(t, func() nativeio.Sender {
		return NewContainer().NativeSender()
	})
}
