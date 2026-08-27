# 架构

[English](architecture.md) | 简体中文

Arkhos 是容器实现，不是新的 Web 标准。公共契约应来自 Arkarta。Arkhos 负责运行时装配、生命周期编排、请求分发和运行集成。

## 边界

- `goark.dev/arkarta` 定义公共标准和 TCK。
- `goark.dev/arkhos` 作为具体容器实现这些契约。
- 内部包保持实现细节私有，只有稳定 API 必要时才导出。
- 兼容性按 Profile 声明，并由 TCK 执行结果证明。

## 运行模型

第一版运行模型基于 Go `net/http`。Arkhos 应保留 Go server 已经正确的行为，只在容器边界补充 Arkarta 语义：部署校验、Servlet/Filter 链、分发、生命周期回调、资源解析、Session、异步处理、安全和协议升级。

当前 `dev` 实现由 Arkhos 内部代码负责容器生命周期和应用匹配，请求/响应边界使用 Arkarta 标准 `net/http` 适配器。这样容器身份属于 Arkhos，后续替换底层适配器时不需要改变应用部署代码。

## 包依赖方向

```text
arkhos public API
    -> internal/container
        -> internal/deploy
        -> internal/nethttp
        -> internal/config
    -> goark.dev/arkarta
```

内部包可以依赖 Arkarta。Arkarta 永远不能反向依赖 Arkhos。

## 设计原则

- 显式注册优先，不做运行时扫描。
- 小接口由消费方包定义。
- 启动后部署快照不可变。
- 生命周期和关闭流程必须感知 `context.Context`。
- 用稳定错误值表达行为，不依赖字符串匹配。
- 兼容性声明必须先有测试证明。

## 当前实现证据

- Servlet Core：`RunCoreHTTP`、`RunHTTPContainer`、`RunLifecycle`、`RunErrorPages` 和 `RunStaticResources`。
- Native I/O：通过可移植兜底发送器通过 `RunNativeIO`。
- 真实 Server 路径：基于 `net.Listener` 的 `Server.Serve` 测试，并覆盖 `context` 取消后的优雅关闭。
