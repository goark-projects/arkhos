# 路线图

[English](roadmap.md) | 简体中文

## v0.0.1 实现目标

第一个实现里程碑应证明 Arkarta Servlet Core 可以稳定运行在 `net/http` 之上。

1. `dev` 已完成：容器构建器和部署快照。
2. `dev` 已完成：通过 Arkarta Core 承载 Servlet/Filter 生命周期和分发链。
3. `dev` 已完成：`net/http` 适配器以及核心请求/响应映射。
4. `dev` 已完成：静态资源和错误分发 TCK 覆盖。
5. `dev` 已完成：已实现 Core 和 Native I/O Profile 的 Arkarta TCK 集成。
6. `dev` 已完成：Session 管理器、基于 Cookie 的请求绑定、ID 轮换和 URL rewriting 辅助。
7. `dev` 已完成：Multipart 解析器集成、请求绑定和请求结束清理。
8. `dev` 已完成：Async 生命周期辅助、流式响应包装、HTTP Upgrade、Security 策略钩子和 WebSocket 集成辅助。

## 后续 Profile 切片

- 平台支持时提供 Native I/O 优化映射。
- 通过 Arkarta JSON 集成 JSON Provider。
- 在显式容器边界集成 Validation。
- Goark Boot starter 和内嵌 Arkhos 自动配置。

## 非目标

- WAR 部署。
- JSP。
- Java 类加载器语义。
- Java 注解扫描。
- 依赖重反射的魔法注册。
