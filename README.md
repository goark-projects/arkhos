# Arkhos

English | [简体中文](README.zh-CN.md)

Arkhos is a Go-native Web container implementation for the Arkarta enterprise Web standard. Its first target is Arkarta `v0.0.1`, starting with Servlet Core and the `net/http` execution model before expanding into optional profiles.

Arkhos is not a Java Servlet compatibility layer and is not a WAR/JSP runtime. It implements Arkarta contracts with explicit Go APIs, `context.Context`, `net/http`, small packages, deterministic lifecycle rules, and TCK-backed behavior.

## Status

The `main` branch contains project governance, documentation, and repository structure. The `dev` branch contains the Arkarta `v0.0.1` implementation slices for Servlet Core, Session, Multipart, Async/Stream, Upgrade, Security, WebSocket integration helpers, a concrete Arkhos `net/http` container, a standard-library HTTP server wrapper, and Native I/O fallback support.

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
- Native I/O portable fallback sender.
- Session, multipart, async, security, upgrade, and WebSocket integration helpers.
- JSON and validation profiles in later slices.
- TCK integration for Arkarta `v0.0.1`.

## Supported Profiles

| Profile | Status | Evidence |
| --- | --- | --- |
| Servlet Core | Implemented on `dev` | `servlet/tck.RunCoreHTTP`, `RunHTTPContainer`, `RunLifecycle`, `RunErrorPages`, `RunStaticResources` |
| Native I/O | Implemented on `dev` through portable fallback | `servlet/tck.RunNativeIO` |
| Session | Implemented on `dev` | `servlet/tck.RunSessionManager`, `RunSessionRequestBinding`, `RunMemorySessionProfile` |
| Multipart | Implemented on `dev` | `servlet/tck.RunMultipartParser` |
| Async/Stream | Implemented on `dev` | `servlet/tck.RunAsyncLifecycle` |
| Upgrade | Implemented on `dev` | `nethttp.Response` implements Servlet `upgrade.HTTP` through standard-library hijack |
| Security | Implemented on `dev` | `servlet/tck.RunSecurity`, `BasicSecurityPolicy`, `ConstraintSecurityPolicy` |
| WebSocket | Implemented on `dev` as Arkarta integration helpers | `websocket/tck.RunHandshake`, `RunFrameCodec`, `RunCompression`, `RunEndpointLifecycle` |

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

## Minimal Example

```go
package main

import (
	"context"
	"log"

	"goark.dev/arkarta/servlet"
	servletcontainer "goark.dev/arkarta/servlet/container"
	"goark.dev/arkhos/nethttp"
)

func main() {
	app, err := servlet.NewWebApp("orders")
	if err != nil {
		log.Fatal(err)
	}
	deployment, err := servletcontainer.NewDeployment(app,
		servletcontainer.WithMapping("/", servlet.HandlerFunc(func(_ context.Context, _ *servlet.Request, res servlet.Response) error {
			_, err := res.WriteString("ok")
			return err
		})),
	)
	if err != nil {
		log.Fatal(err)
	}

	container := nethttp.NewContainer()
	if _, err := container.Deploy(context.Background(), deployment); err != nil {
		log.Fatal(err)
	}
	server, err := nethttp.NewServer(container, nethttp.WithAddress(":8080"))
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(server.ListenAndServe(context.Background()))
}
```

## Compatibility Policy

Arkhos may claim support for an Arkarta profile only after the corresponding Arkarta TCK passes in this repository. Partial implementations remain internal until they have tests and documented behavior.

## License

Arkhos is licensed under the [Apache License 2.0](LICENSE).
