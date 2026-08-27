# Arkhos

English | [简体中文](README.zh-CN.md)

Arkhos is a Go-native Web container implementation for the Arkarta enterprise Web standard. Its first target is Arkarta `v0.0.1`, starting with Servlet Core and the `net/http` execution model before expanding into optional profiles.

Arkhos is not a Java Servlet compatibility layer and is not a WAR/JSP runtime. It implements Arkarta contracts with explicit Go APIs, `context.Context`, `net/http`, small packages, deterministic lifecycle rules, and TCK-backed behavior.

## Status

Arkhos is being initialized. The `main` branch contains project governance, documentation, and repository structure. Runtime implementation begins on the `dev` branch.

## Design Goals

- Implement Arkarta `v0.0.1` as a production-grade Web container.
- Keep the standard Go-native instead of copying Java runtime scanning, reflection-heavy deployment, or classloader semantics.
- Preserve clear package ownership and single-responsibility files.
- Treat Arkarta TCKs as the compatibility gate for every implemented profile.
- Build on the Go standard library first, especially `net/http`, `context`, and `testing`.

## Initial Scope

- Servlet Core container lifecycle.
- `net/http` adapter and request dispatch.
- Deployment model for explicit servlet and filter registration.
- Static resource handling.
- Session, multipart, async, security, native I/O, WebSocket, JSON, and validation profiles in later slices.
- TCK integration for Arkarta `v0.0.1`.

## Project Layout

```text
cmd/arkhos/        Command entry point, added when the container CLI is introduced.
internal/config/   Internal configuration parsing and validation.
internal/container Runtime lifecycle, deployment, and dispatch internals.
internal/deploy/   Deployment assembly and validation.
internal/nethttp/  net/http adapter internals.
tests/tck/         Arkarta TCK integration tests.
docs/              Architecture, roadmap, development, and decision records.
```

## Commands

```bash
go test ./...
go vet ./...
gofmt -w .
```

## Compatibility Policy

Arkhos may claim support for an Arkarta profile only after the corresponding Arkarta TCK passes in this repository. Partial implementations remain internal until they have tests and documented behavior.

## License

Arkhos is licensed under the [Apache License 2.0](LICENSE).
