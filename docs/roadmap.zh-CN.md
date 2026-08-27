# 路线图

[English](roadmap.md) | 简体中文

## v0.0.1 实现目标

第一个实现里程碑应证明 Arkarta Servlet Core 可以稳定运行在 `net/http` 之上。

1. `dev` 已完成：容器构建器和部署快照。
2. `dev` 已完成：通过 Arkarta Core 承载 Servlet/Filter 生命周期和分发链。
3. `dev` 已完成：`net/http` 适配器以及核心请求/响应映射。
4. `dev` 已完成：静态资源和错误分发 TCK 覆盖。
5. `dev` 已完成：已实现 Core 和 Native I/O Profile 的 Arkarta TCK 集成。

## 后续 Profile 切片

- Session 管理器和 Cookie 策略。
- Multipart 解析和清理策略。
- 异步请求生命周期和流式处理。
- 安全约束和 Principal 传递。
- 平台支持时提供 Native I/O 优化映射。
- WebSocket 握手、子协议协商和帧运行时。
- 通过 Arkarta JSON 集成 JSON Provider。
- 在显式容器边界集成 Validation。

## 非目标

- WAR 部署。
- JSP。
- Java 类加载器语义。
- Java 注解扫描。
- 依赖重反射的魔法注册。
