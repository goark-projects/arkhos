package nethttp

import "goark.dev/arkhos/internal/deploy"

// ErrUnsupportedProfile 表示部署声明了当前 net/http 容器不支持的 Arkarta Profile。
var ErrUnsupportedProfile = deploy.ErrUnsupportedProfile
