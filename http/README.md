# HTTP

Standardizing HTTP API development for GO applications.  
Uses [go-playground/validator](https://github.com/go-playground/validator) to validate decoded payloads.  
Uses [Prometheus client_golang](https://github.com/prometheus/client_golang) to expose `/metrics`.  
Uses [OTel net/http instrumentation](https://pkg.go.dev/go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp) for trace propagation and spans.  
Provides shared middleware, handlers, clients, and error helpers so services stay consistent.

Contains HTTP handlers and middleware that can be re-used across APIs
Provides an `auth` packages for consistent authentication
Provides an `errors` package for consistent error handling

## Usage

```bash
go get github.com/jesse0michael/pkg/http
```
