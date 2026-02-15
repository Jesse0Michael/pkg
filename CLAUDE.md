# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Go monorepo of reusable packages using Go workspaces (`go.work`). Each package has its own `go.mod` and is independently versioned via release-please.

Packages: `auth`, `boot`, `cache`, `config`, `grpc`, `http`, `logger`, `test`

## Commands

```bash
make test       # Run tests with coverage across all modules
make lint       # Run golangci-lint across all modules
make tidy       # Run go mod tidy across all modules
make generate   # Run go generate across all modules
make vuln       # Run govulncheck across all modules
```

## Architecture

**Go workspace monorepo** — each subdirectory is an independent module (`github.com/jesse0michael/pkg/<package>`). The root `go.mod` only holds linting tool dependencies.

## Code Style

- Wrap errors with context: `fmt.Errorf("failed to do x: %w", err)`
- Structured logging with `slog`: `slog.InfoContext(ctx, "msg", "key", val)` — use `"err"` attribute for errors
- Comments explain "why", not "what"
- Dependency injection with interfaces for all services

## Testing Conventions

- **Table-driven tests** with `testify/require` for assertions
- Use `t.Context()` for context in tests
- Use `gomock` with a `mockSetup` test case function for mocking
- Test identifiers: `test-{type}` or `test-{type}-{index}` (e.g., `test-user`, `test-user-1`)
- Test fixtures in `testdata/` directories, loaded via `test.LoadFile` / `test.LoadJSONFile`
