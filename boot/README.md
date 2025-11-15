# boot

Bootstrap Go services. The `boot` module wires together logging, config loading, context cancellation, [OpenTelemetry Go SDK](https://go.opentelemetry.io/otel/v1.37.0) linkage, and background runners so your `main` package stays focused on what matters.  
Uses [`context.WithCancelCause`](https://pkg.go.dev/context#WithCancelCause) to manage lifecycle and signal propagation.  
Uses `config.AppConfig`/`config.OpenTelemetryConfig` to standardize resource metadata and exporter setup.

## Usage

```bash
go get github.com/jesse0michael/pkg/boot
```

```go
package main

import (
	"context"
	"os"

	"github.com/jesse0michael/pkg/boot"
	"github.com/jesse0michael/pkg/config"
	"github.com/jesse0michael/pkg/http/server"
)

type Config struct {
	config.AppConfig
	config.OpenTelemetryConfig
	server.Config
}

type Server struct {
	srv *server.Server
}

func (h *Server) Run(ctx context.Context, cfg Config) error {
	h.srv = server.New(cfg.HTTP, myRouter{})
	return h.srv.ListenAndServe()
}

func (h *Server) Close() error {
	return h.srv.Shutdown(context.Background())
}

func main() {
	app := boot.NewApp[Config]()
	if err := app.Run(&Server{}); err != nil {
		os.Exit(1)
	}
}
```
