# 规范：Arkhos Hertz 默认容器

## 目标

实现基于 Hertz 的 Arkhos 默认生产容器，使 Arkarta 应用可以像 Jakarta Servlet 应用在 Tomcat、Jetty 之间切换一样，在不同合规容器之间替换而不修改业务代码。`nethttp` 继续作为参考、兼容及 TCK 基准实现。

## 技术栈

- Go 1.25 语言基线；Windows 和 Debian 使用当前项目工具链复验。
- `goark.dev/arkarta`：传输层中立 Servlet 标准和 TCK。
- `github.com/cloudwego/hertz v0.10.6`：默认 HTTP 引擎。
- `github.com/cloudwego/netpoll v0.7.5`：非 Windows 默认网络传输。
- Go 标准库 `net/http`：参考实现和互操作测试工具，不进入 Hertz 生产热路径。

## 命令

Windows：

```powershell
go test ./...
go test -race ./...
go vet ./...
gofmt -l .
git diff --check
```

Debian：

```bash
export PATH=/opt/go/bin:$PATH
go test ./...
go test -race ./...
go vet ./...
test -z "$(gofmt -l .)"
git diff --check
```

## 项目结构

```text
arkarta/servlet/             标准 Request、Response 和 Profile 契约
arkarta/servlet/nethttp/     标准库参考适配器
arkarta/servlet/tck/         与具体传输无关的容器兼容性测试
arkhos/internal/container/   共享部署、匹配和生命周期运行时
arkhos/hertz/                Hertz 公共容器与服务器入口
arkhos/hertz/internal/       Hertz 请求、响应、升级和生命周期适配细节
arkhos/nethttp/              标准库参考、兼容及基准实现
arkhos/tests/tck/            Arkhos 各实现的标准兼容性验证
```

不为了目录形式创建单一用途抽象；只有具有独立职责、可单测边界或需要隐藏第三方类型的逻辑才进入子包。

## 代码规范

```go
// RequestSource 只提供一次请求的不可变传输元数据。
// 容器不得在请求生命周期结束后继续访问底层池化对象。
type RequestSource interface {
	Context() context.Context
	Method() string
	Body() io.ReadCloser
}
```

- 公共接口由使用方定义，保持小而稳定。
- 所有错误使用稳定错误值并通过 `%w` 保留错误链。
- 代码和文档使用 UTF-8、LF；代码注释使用标准简体中文。
- 不使用运行时扫描、反射代理或全局可变注册表。
- 热路径不进行无依据的对象池化；优化必须由分配基准和 CPU Profile 证明。

## 测试策略

- 标准契约先写失败测试，再实现最小可用行为。
- `arkarta/servlet/tck` 对容器使用传输中立测试驱动，覆盖 Core、Session、Multipart、Async/Stream、Security、Upgrade、Native I/O 和 WebSocket。
- `arkhos/nethttp` 与 `arkhos/hertz` 运行同一套 Arkhos TCK。
- Hertz 增加真实监听器集成测试，覆盖取消、优雅关闭、请求正文、Header/Cookie、Flush、错误页和连接升级。
- Async 测试必须证明处理函数返回后不会读取已归还 Hertz 对象池的状态，并通过竞态检测。
- 基准至少报告吞吐、每操作耗时、每操作分配和字节分配；跨引擎结论必须使用相同业务处理器和负载。

## 边界

始终执行：

- 应用和 Goark Web 只依赖 Arkarta 标准。
- 每个实现按已通过 TCK 的结果声明 Profile。
- 每个完成切片测试后独立提交并推送。
- 修改 Arkhos 后检查 Goark、goark-boot、相关 contrib/starter、CLI 模板和 demo。

需要明确决策后执行：

- 增加 Arkarta 标准范围之外的容器扩展。
- 对外声明具体性能倍数或跨协议优势。

禁止执行：

- Arkarta 标准核心导入 Hertz 或暴露具体传输对象。
- Hertz 生产路径通过 `http.Handler` 或合成 `http.Request` 转发。
- Async 生命周期持有可被 Hertz 复用的 RequestContext、Header 或 Body 缓冲区。
- 为通过测试而跳过、削弱或复制一套实现专用 TCK。

## 成功标准

1. Arkarta 公共 Servlet/Profile API 不暴露具体 HTTP 框架类型。
2. `arkhos.New()` 默认创建 Hertz 容器，`arkhos/nethttp` 保持显式可用。
3. Hertz 同步热路径直接桥接 Arkarta，不合成标准库请求。
4. Async/Stream 在池化生命周期边界上内存安全、无竞态、无请求串扰。
5. Hertz 与 `nethttp` 对所有声明 Profile 通过同一套 TCK。
6. Goark、goark-boot、contrib/starter、CLI 模板和 demo 使用默认 Hertz 路径，同时能显式替换为合规容器。
7. Windows 与 Debian 完成单测、竞态、vet、格式、差异检查和真实 Server 冒烟。
8. 基准能够分离容器语义成本与网络引擎成本，不作无数据的性能承诺。

## 未决问题

无。Hertz 默认、`nethttp` 保留和标准层彻底传输中立均已确认。
