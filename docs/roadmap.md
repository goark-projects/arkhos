# Roadmap

English | [简体中文](roadmap.zh-CN.md)

## v0.0.1 Implementation Target

The first implementation milestone should prove Arkarta Servlet Core on top of `net/http`.

1. Done on `dev`: container builder and deployment snapshot.
2. Done on `dev`: Servlet/filter lifecycle and dispatch chain through Arkarta Core.
3. Done on `dev`: `net/http` adapter and core request/response mapping.
4. Done on `dev`: static resources and error dispatch TCK coverage.
5. Done on `dev`: Arkarta TCK integration for implemented Core and Native I/O profiles.
6. Done on `dev`: Session manager, cookie-backed request binding, ID rotation, and URL rewriting helpers.
7. Done on `dev`: Multipart parser integration, request binding, and request-end cleanup.
8. Done on `dev`: Async lifecycle helpers, stream wrappers, HTTP upgrade, Security policy hooks, and WebSocket integration helpers.

## Later Profile Slices

- Native I/O optimized platform mappings where the platform supports them.
- JSON provider integration through Arkarta JSON.
- Validation integration at explicit container boundaries.
- Goark Boot starter and embedded Arkhos auto-configuration.

## Non-Goals

- WAR deployment.
- JSP.
- Java classloader semantics.
- Java annotation scanning.
- Reflection-heavy magic registration.
