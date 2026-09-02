package nethttp

import (
	"errors"

	"goark.dev/arkhos/internal/deploy"
	internalprofile "goark.dev/arkhos/internal/profile"
)

// ErrUnsupportedProfile 表示部署声明了当前 net/http 容器不支持的 Arkarta Profile。
var ErrUnsupportedProfile = deploy.ErrUnsupportedProfile

// ErrNilContainer 表示 Server 缺少容器实例。
var ErrNilContainer = errors.New("arkhos/nethttp: container is nil")

// ErrNilApplication 表示容器缺少已部署应用实例。
var ErrNilApplication = errors.New("arkhos/nethttp: application is nil")

// ErrNilListener 表示 Serve 缺少网络监听器。
var ErrNilListener = errors.New("arkhos/nethttp: listener is nil")

// ErrSessionProfileUnavailable 表示当前请求没有可用 Session Profile。
var ErrSessionProfileUnavailable = internalprofile.ErrSessionUnavailable

// ErrMultipartProfileUnavailable 表示当前请求没有可用 Multipart Profile。
var ErrMultipartProfileUnavailable = internalprofile.ErrMultipartUnavailable
