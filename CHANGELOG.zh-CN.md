# 变更日志

[English](CHANGELOG.md) | 简体中文

这里记录 Arkhos 的重要变更。

## [Unreleased]

### 修复

- 修复空闲 HTTP keep-alive 连接导致优雅关闭耗尽全部超时的问题，同时保证正在处理的请求能够完成。
- 修复快速启停时 Hertz 退出结果可能在 Server 等待循环前被消费的竞态。

### Added

- 初始化 Arkhos 仓库，加入 Apache License 2.0、Go module 元数据、双语文档，以及面向 Arkarta `v0.0.1` 实现工作的首批包结构。
- 在 `dev` 分支加入 Arkarta `v0.0.1` 的第一批实现切片：Arkhos `net/http` 容器、Servlet Core 分发、Server 运行封装、Native I/O 兜底发送器，以及已声明 Profile 的 TCK 覆盖。
- 加入 Arkarta Session、Multipart、Async/Stream、Upgrade、Security 和 WebSocket 集成辅助，提供 Arkhos 请求级 Profile 装配和 TCK 覆盖。
