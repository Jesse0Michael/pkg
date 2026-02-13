# pkg

A collection of common Go modules that cover the building blocks of Go services. Each module is independent so you can import only what you need.

| Module | Description |
| --- | --- |
| [Auth](./auth) | Utilities for propagating identity through contexts and issuing/validating JWT tokens. |
| [Boot](./boot) | Application bootstrap helpers that coordinate configuration, logging, telemetry, and runner lifecycles. |
| [Cache](./cache) | Caching primitives that pair Redis with local safeguards and cache-control awareness. |
| [Config](./config) | Environment processing helpers and shared configuration structs for common infrastructure. |
| [GRPC](./grpc) | Foundations for consistent gRPC services—middleware, handlers, clients, and error helpers. |
| [HTTP](./http) | Foundations for consistent HTTP services—middleware, handlers, parsers, clients, and error helpers. |
| [Logger](./logger) | Opinionated `slog` setup that standardizes structured logging across services. |
| [Test](./test) | Test helpers for working with time, fixtures, and mock HTTP servers. |
