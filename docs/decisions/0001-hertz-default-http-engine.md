# ADR-0001：Hertz 作为 Arkhos 默认 HTTP 引擎

## 状态

已接受

## 日期

2026-09-02

## 背景

Arkhos 是 Arkarta Servlet 标准的容器实现。现有实现使用 Go 标准库 `net/http`，它仍然适合作为可移植参考实现、兼容入口和 TCK 行为基准，但不能代表 Arkhos 的高性能默认网络路径。

候选方案包括 fasthttp、Hertz、gnet 和 CloudWeGo Netpoll。Arkhos 需要完整 HTTP 语义、清晰生命周期、流式响应、连接升级、可扩展协议栈以及生产级维护能力，而不是只提供事件循环或连接读写能力。

## 决策

1. `goark.dev/arkhos/hertz` 是 Arkhos 默认生产容器实现。
2. Hertz 固定使用 `github.com/cloudwego/hertz v0.10.6`；非 Windows 默认使用其 Netpoll 传输，Windows 使用 Hertz 标准网络传输保证开发和测试可用性。
3. `goark.dev/arkhos/nethttp` 保留为参考实现、兼容实现和 TCK 行为基准，不删除、不降级测试覆盖。
4. 根包 `arkhos.New` 返回默认 Hertz 容器；需要参考实现的调用方显式导入 `arkhos/nethttp`。
5. Hertz 适配器直接桥接 Arkarta 传输中立契约，不在生产热路径构造 `http.Request` 或通过 `http.Handler` 转发。
6. Arkhos 的部署、应用匹配和生命周期继续由共享的 `internal/container` 负责；Hertz 与 `nethttp` 只拥有各自协议边界。
7. Async 请求在越过 Hertz `RequestContext` 生命周期前必须完成所有权转移或必要快照，禁止持有将被对象池复用的状态。
8. 每个 Arkarta Profile 只有在 Hertz 容器通过对应 Arkhos TCK 后才能出现在容器元数据中。

## 备选方案

### fasthttp

HTTP/1.1 性能成熟，但对象模型和协议路线限制不适合作为企业 Servlet 容器的长期基础，因此拒绝。

### 直接使用 gnet 或 Netpoll

二者属于网络 I/O 层。直接采用意味着 Arkhos 自行承担 HTTP 解析、协议状态机、限流、防御性边界和升级行为，维护及安全成本过高，因此拒绝。Netpoll 通过 Hertz 间接使用。

### 彻底删除 `nethttp`

会失去标准库兼容入口和独立行为基准，也不利于定位 Hertz 适配问题，因此拒绝。

## 后果

- Arkhos 增加 Hertz 依赖和独立的 `hertz` 子包。
- starter、demo 和文档默认装配 Hertz，同时保留显式选择 `nethttp` 的能力。
- 性能声明必须来自相同主机、相同协议、相同业务处理器的基准结果。
- Linux 与 Windows 的网络传输实现不同，但 Arkarta 可观察语义必须由同一套 TCK 保持一致。
