package hertz

import (
	"errors"

	internalprofile "goark.dev/arkhos/internal/profile"
)

var (
	// ErrNilApplication 表示容器应用为空。
	ErrNilApplication = errors.New("arkhos/hertz: application is nil")
	// ErrNilContainer 表示容器实例为空。
	ErrNilContainer = errors.New("arkhos/hertz: container is nil")
)

var (
	// ErrSessionProfileUnavailable 表示当前请求没有可用 Session Profile。
	ErrSessionProfileUnavailable = internalprofile.ErrSessionUnavailable
	// ErrMultipartProfileUnavailable 表示当前请求没有可用 Multipart Profile。
	ErrMultipartProfileUnavailable = internalprofile.ErrMultipartUnavailable
)
