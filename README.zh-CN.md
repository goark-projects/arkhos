# Arkhos

[English](README.md) | 简体中文

Arkhos 是 Arkarta 企业 Web 标准的 Go 原生 Web 容器实现。第一阶段目标是 Arkarta `v0.0.1`，先完成 Servlet Core 和 `net/http` 执行模型，再逐步扩展可选 Profile。

Arkhos 不是 Java Servlet 兼容层，也不是 WAR/JSP 运行时。它使用显式 Go API、`context.Context`、`net/http`、小包边界、确定性生命周期和 TCK 校验来实现 Arkarta 契约。

## 当前状态

`main` 分支保存项目治理、文档和仓库结构。`dev` 分支已经包含 Arkarta `v0.0.1` 的多批实现切片：基于 `net/http` 的 Servlet Core、Session、Multipart、Async/Stream、Upgrade、Security、WebSocket 集成辅助、Arkhos 具体容器、标准库 HTTP Server 封装，以及 Native I/O 可移植兜底能力。

## 设计目标

- 以生产级标准实现 Arkarta `v0.0.1`。
- 保持 Go 原生设计，不照搬 Java 运行时扫描、重反射部署或类加载器语义。
- 保持清晰的包归属和单一职责文件。
- 每个已实现 Profile 都以 Arkarta TCK 作为兼容性门禁。
- 优先使用 Go 标准库，尤其是 `net/http`、`context` 和 `testing`。

## 初始范围

- Servlet Core 容器生命周期。
- `net/http` 适配器和请求分发。
- 显式 Servlet 与 Filter 注册的部署模型。
- 静态资源处理。
- Native I/O 可移植兜底发送器。
- Session、multipart、async、security、upgrade 和 WebSocket 集成辅助。
- JSON 和 validation Profile 后续分批实现。
- Arkarta `v0.0.1` 的 TCK 集成。

## 已支持 Profile

| Profile | 状态 | 证据 |
| --- | --- | --- |
| Servlet Core | `dev` 已实现 | `servlet/tck.RunCoreHTTP`、`RunHTTPContainer`、`RunLifecycle`、`RunErrorPages`、`RunStaticResources` |
| Native I/O | `dev` 已通过可移植兜底实现 | `servlet/tck.RunNativeIO` |
| Session | `dev` 已实现 | `servlet/tck.RunSessionManager`、`RunSessionRequestBinding`、`RunMemorySessionProfile` |
| Multipart | `dev` 已实现 | `servlet/tck.RunMultipartParser` |
| Async/Stream | `dev` 已实现 | `servlet/tck.RunAsyncLifecycle` |
| Upgrade | `dev` 已实现 | `nethttp.Response` 通过标准库 hijack 支持 Servlet `upgrade.HTTP` |
| Security | `dev` 已实现 | `servlet/tck.RunSecurity`、`BasicSecurityPolicy`、`ConstraintSecurityPolicy` |
| WebSocket | `dev` 已作为 Arkarta 集成辅助实现 | `websocket/tck.RunHandshake`、`RunFrameCodec`、`RunCompression`、`RunEndpointLifecycle` |

## 项目结构

```text
cmd/arkhos/        引入容器 CLI 时使用的命令入口。
internal/config/   内部配置解析与校验。
internal/container 运行时生命周期、部署和分发内部实现。
internal/deploy/   部署装配与校验。
internal/nethttp/  net/http 适配内部实现。
tests/tck/         Arkarta TCK 集成测试。
docs/              架构、路线图、开发规范和决策记录。
```

## 常用命令

```bash
go test ./...
go vet ./...
gofmt -w .
```

## 最小示例

```go
package main

import (
	"context"
	"log"

	"goark.dev/arkarta/servlet"
	servletcontainer "goark.dev/arkarta/servlet/container"
	"goark.dev/arkhos/nethttp"
)

func main() {
	app, err := servlet.NewWebApp("orders")
	if err != nil {
		log.Fatal(err)
	}
	deployment, err := servletcontainer.NewDeployment(app,
		servletcontainer.WithMapping("/", servlet.HandlerFunc(func(_ context.Context, _ *servlet.Request, res servlet.Response) error {
			_, err := res.WriteString("ok")
			return err
		})),
	)
	if err != nil {
		log.Fatal(err)
	}

	container := nethttp.NewContainer()
	if _, err := container.Deploy(context.Background(), deployment); err != nil {
		log.Fatal(err)
	}
	server, err := nethttp.NewServer(container, nethttp.WithAddress(":8080"))
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(server.ListenAndServe(context.Background()))
}
```

## 兼容性策略

只有当对应 Arkarta TCK 在本仓库通过后，Arkhos 才声明支持某个 Arkarta Profile。未完成的部分保持内部状态，直到测试和文档都明确为止。

## 许可证

Arkhos 使用 [Apache License 2.0](LICENSE)。
