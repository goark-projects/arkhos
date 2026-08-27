# Development

English | [简体中文](development.zh-CN.md)

## Branch Model

- `main` holds repository initialization, release documentation, and stable baselines.
- `dev` holds implementation work for the next Arkhos release.
- Each meaningful implementation slice should be tested and committed independently.

## Local Commands

```bash
go test ./...
go vet ./...
gofmt -w .
git diff --check
```

## Coding Rules

- Use UTF-8 files with LF line endings.
- Write Go code in small packages with single-purpose files.
- Keep public APIs explicit and Go-native.
- Use Simplified Chinese for code comments.
- Add unit tests for Go behavior before declaring a slice complete.
- Do not expose partial profile support as compatible until its Arkarta TCK passes.

## Dependency Policy

Start with Go standard library dependencies. Add external dependencies only when they carry real protocol/runtime value and do not leak into Arkarta contracts.
