package nethttp

import (
	"errors"

	"goark.dev/arkhos/internal/deploy"
)

// ErrUnsupportedProfile 表示部署声明了当前 net/http 容器不支持的 Arkarta Profile。
var ErrUnsupportedProfile = deploy.ErrUnsupportedProfile

// ErrNilContainer 表示 Server 缺少容器实例。
var ErrNilContainer = errors.New("arkhos/nethttp: container is nil")

// ErrNilListener 表示 Serve 缺少网络监听器。
var ErrNilListener = errors.New("arkhos/nethttp: listener is nil")
