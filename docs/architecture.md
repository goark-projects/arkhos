# Architecture

English | [简体中文](architecture.zh-CN.md)

Arkhos is a container implementation, not a new Web standard. Public contracts should come from Arkarta. Arkhos owns runtime assembly, lifecycle orchestration, request dispatch, and operational integration.

## Boundaries

- `goark.dev/arkarta` defines the public standard and TCKs.
- `goark.dev/arkhos` implements those contracts as a concrete container.
- Internal packages keep implementation details private until a stable exported API is required.
- Compatibility is profile-based and proven by TCK execution.

## Runtime Model

The first runtime model is based on Go `net/http`. Arkhos should preserve Go server behavior where it is already correct, and add Arkarta semantics only at the container boundary: deployment validation, servlet/filter chains, dispatch, lifecycle callbacks, resource resolution, sessions, async handling, security, and upgrade support.

The current `dev` implementation owns the container lifecycle and application matching in Arkhos internals, while the request/response edge uses Arkarta's standard `net/http` adapter. This keeps the public container identity in Arkhos and allows the low-level adapter to be replaced later without changing application deployment code.

Optional Servlet profiles are wired as Arkhos application decorators. The decorator binds a small request-local profile context before the managed Arkarta application handler runs, so Session, Multipart, Security, Async, and Upgrade helpers remain container-owned without changing Arkarta's core application construction.

## Package Direction

```text
arkhos public API
    -> internal/container
        -> internal/deploy
        -> internal/nethttp
        -> internal/config
    -> goark.dev/arkarta
```

Internal packages may depend on Arkarta. Arkarta must never depend on Arkhos.

## Design Principles

- Explicit registration over runtime scanning.
- Small interfaces owned by the consumer package.
- Immutable deployment snapshots after startup.
- Context-aware lifecycle and shutdown.
- Error values with stable behavior instead of string matching.
- Tests before compatibility claims.

## Current Implementation Evidence

- Servlet Core: `RunCoreHTTP`, `RunHTTPContainer`, `RunLifecycle`, `RunErrorPages`, and `RunStaticResources`.
- Native I/O: `RunNativeIO` through the portable fallback sender.
- Session: `RunSessionManager`, `RunSessionRequestBinding`, and `RunMemorySessionProfile`.
- Multipart: `RunMultipartParser` plus request-end cleanup through the Arkhos profile decorator.
- Async/Stream and Security: `RunAsyncLifecycle`, `RunSecurity`, and focused Arkhos request-bound policy tests.
- Upgrade and WebSocket: standard-library hijack support, Arkarta WebSocket handshake/frame/compression/endpoint TCKs, and Servlet Upgrade integration helpers.
- Real server path: `net.Listener` based `Server.Serve` test with graceful context cancellation.
