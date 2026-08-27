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
