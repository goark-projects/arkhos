package arkhos

import "goark.dev/arkhos/hertz"

const (
	// Version 是 Arkhos 当前实现版本。
	Version = "0.0.1"
	// ArkartaVersion 是当前实现对齐的 Arkarta 标准版本。
	ArkartaVersion = "v0.0.1"
)

// New 创建默认 Hertz 容器实现。
func New(options ...hertz.ContainerOption) *hertz.Container {
	return hertz.NewContainer(options...)
}

// NewServer 创建默认 Hertz HTTP Server。
func NewServer(container *hertz.Container, options ...hertz.ServerOption) (*hertz.Server, error) {
	return hertz.NewServer(container, options...)
}
