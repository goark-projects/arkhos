package profile

import "errors"

var (
	// ErrNilApplication 表示待装饰应用为空。
	ErrNilApplication = errors.New("arkhos/profile: application is nil")
	// ErrSessionUnavailable 表示当前请求没有 Session Profile。
	ErrSessionUnavailable = errors.New("arkhos/profile: session profile unavailable")
	// ErrMultipartUnavailable 表示当前请求没有 Multipart Profile。
	ErrMultipartUnavailable = errors.New("arkhos/profile: multipart profile unavailable")
)
