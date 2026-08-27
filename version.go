package arkhos

import "goark.dev/arkhos/nethttp"

const (
	// Version 是 Arkhos 当前实现版本。
	Version = "0.0.1"
	// ArkartaVersion 是当前实现对齐的 Arkarta 标准版本。
	ArkartaVersion = "v0.0.1"
)

// New 创建默认 net/http 容器实现。
func New() *nethttp.Container {
	return nethttp.NewContainer()
}
