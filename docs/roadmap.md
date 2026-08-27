# Roadmap

English | [简体中文](roadmap.zh-CN.md)

## v0.0.1 Implementation Target

The first implementation milestone should prove Arkarta Servlet Core on top of `net/http`.

1. Done on `dev`: container builder and deployment snapshot.
2. Done on `dev`: Servlet/filter lifecycle and dispatch chain through Arkarta Core.
3. Done on `dev`: `net/http` adapter and core request/response mapping.
4. Done on `dev`: static resources and error dispatch TCK coverage.
5. Done on `dev`: Arkarta TCK integration for implemented Core and Native I/O profiles.

## Later Profile Slices

- Session manager and cookie policy.
- Multipart parsing and cleanup policy.
- Async request lifecycle and streaming.
- Security constraints and principal propagation.
- Native I/O optimized platform mappings where the platform supports them.
- WebSocket handshake, subprotocol negotiation, and frame runtime.
- JSON provider integration through Arkarta JSON.
- Validation integration at explicit container boundaries.

## Non-Goals

- WAR deployment.
- JSP.
- Java classloader semantics.
- Java annotation scanning.
- Reflection-heavy magic registration.
