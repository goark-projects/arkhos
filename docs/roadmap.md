# Roadmap

English | [简体中文](roadmap.zh-CN.md)

## v0.0.1 Implementation Target

The first implementation milestone should prove Arkarta Servlet Core on top of `net/http`.

1. Container builder and deployment snapshot.
2. Servlet/filter lifecycle and dispatch chain.
3. `net/http` adapter and core request/response mapping.
4. Static resources and error dispatch.
5. Arkarta TCK integration for the implemented core profile.

## Later Profile Slices

- Session manager and cookie policy.
- Multipart parsing and cleanup policy.
- Async request lifecycle and streaming.
- Security constraints and principal propagation.
- Native I/O profile mapping where the platform supports it.
- WebSocket handshake, subprotocol negotiation, and frame runtime.
- JSON provider integration through Arkarta JSON.
- Validation integration at explicit container boundaries.

## Non-Goals

- WAR deployment.
- JSP.
- Java classloader semantics.
- Java annotation scanning.
- Reflection-heavy magic registration.
