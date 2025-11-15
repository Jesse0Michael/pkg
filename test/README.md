# test

Shared test utilities for Go services.  
Uses the standard library's `testing`, `net/http/httptest`, and `time` packages to provide mock HTTP servers, fixture helpers, and deterministic time parsing.

## Usage

```bash
go get github.com/jesse0michael/pkg/test
```

The package includes:
- `test.LoadFile` / `test.LoadJSONFile` – fixture loaders that report failures via `testing.T`.
- `test.ParseTime` / `test.ParseTimeInLocation` – helpers for parsing timestamps without cluttering tests with error handling.
