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

## Options

`NewApp` accepts options to configure how the underlying `config.New[T]()` loads configuration.

| Option | Description |
|---|---|
| `WithConfigPrefix(prefix)` | Namespace env vars with a prefix (e.g. `MYAPP_HOST`) |
| `WithConfigFile(path)` | Load a JSON or YAML config file. Can be called multiple times; files are applied in order |

Configuration is loaded using the `config` package with the following precedence:

**`Init()` < struct defaults < env vars < config files (in order) < CLI args**

CLI arguments from `os.Args[1:]` are automatically parsed into the config struct. Use `arg:"-"` to exclude fields. Config structs can implement `config.Initializer` to set dynamic initial values (e.g. `AppConfig` auto-populates name and version from build info). See the [config README](../config/README.md) for full details.

```go
app := boot.NewApp[Config](
    boot.WithConfigPrefix("MYAPP"),
    boot.WithConfigFile("config.yaml"),
)
```
