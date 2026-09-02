package hertz

import "errors"

var (
	// ErrNilApplication 表示容器应用为空。
	ErrNilApplication = errors.New("arkhos/hertz: application is nil")
	// ErrNilContainer 表示容器实例为空。
	ErrNilContainer = errors.New("arkhos/hertz: container is nil")
)
