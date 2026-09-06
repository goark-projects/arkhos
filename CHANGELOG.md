# Changelog

English | [简体中文](CHANGELOG.zh-CN.md)

All notable changes to Arkhos are documented in this file.

## [Unreleased]

No unreleased changes.

## [0.0.1] - 2026-09-06

### Fixed

- Fixed graceful shutdown with idle HTTP keep-alive connections so active requests can finish without consuming the entire shutdown deadline.
- Fixed a rapid start/stop race that could consume the Hertz exit result before the server wait loop observed it.
- Applied bounded request-header and idle timeouts to the `net/http` server
  while preserving streaming response behavior.

### Added

- Initialized the Arkhos repository with Apache License 2.0, Go module metadata, bilingual documentation, and the first package layout for Arkarta `v0.0.1` implementation work.
- Added the first Arkarta `v0.0.1` implementation slice on `dev`: Arkhos `net/http` container, Servlet Core dispatch, server runtime wrapper, Native I/O fallback sender, and TCK coverage for claimed profiles.
- Added Arkarta Session, Multipart, Async/Stream, Upgrade, Security, and WebSocket integration helpers with request-bound Arkhos profile wiring and TCK coverage.
- Added the Hertz-native container and managed server as the default engine,
  while retaining the `net/http` implementation.
- Added readiness reporting, configurable form limits, immediate close, and
  deterministic graceful shutdown.
- Raised the Go baseline to 1.26 and added cross-platform CI with race tests.
- Aligned all used `golang.org/x` modules with their latest stable releases.

[Unreleased]: https://github.com/goark-projects/arkhos/compare/v0.0.1...HEAD
[0.0.1]: https://github.com/goark-projects/arkhos/releases/tag/v0.0.1
