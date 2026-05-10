# Config

Configuration loading for Go applications with layered sources.

```bash
go get github.com/jesse0michael/pkg/config
```

## Usage

Define a config struct and call `config.New[T]()` with options to control how values are loaded.

```go
type Config struct {
    config.AppConfig
    Host    string `envconfig:"HOST" default:"localhost"`
    Port    int    `envconfig:"PORT" default:"8080"`
    Debug   bool   `envconfig:"DEBUG"`
    MySQL   config.MysqlConfig
}

func main() {
    cfg, err := config.New[Config](
        config.WithPrefix("MYAPP"),
        config.WithFile("config.json"),
    )
    if err != nil {
        panic(err)
    }
}
```

## Source Hierarchy

Values are layered in order of increasing precedence:

**struct defaults < env vars < config files (in order) < CLI args**

Each layer overwrites values set by previous layers. When multiple config files are provided, they are applied in order so later files override earlier ones.

### Merge Behavior

- **Scalars, slices**: later layer replaces the value entirely
- **Maps**: keys are merged across layers — later layers add or overwrite individual keys without removing existing ones

### Options

| Option | Description |
|---|---|
| `WithPrefix(prefix)` | Namespace env vars with a prefix (e.g. `MYAPP_HOST`) |
| `WithFile(path)` | Load a JSON or YAML config file. Missing files are silently skipped. Can be called multiple times; files are applied in order |

### CLI Args

CLI arguments from `os.Args[1:]` are automatically parsed into the config struct.
Fields are exposed as flags by default; use `arg:"-"` to exclude a field.
Unknown arguments are silently ignored.

```go
type Config struct {
    Host   string `envconfig:"HOST" default:"localhost" help:"server host"`
    Port   int    `envconfig:"PORT" default:"8080" help:"server port"`
    Secret string `envconfig:"SECRET" arg:"-"`
}
```

## Config File Example

JSON:
```json
{
    "host": "production-host",
    "port": 9090
}
```

YAML:
```yaml
host: production-host
port: 9090
```

## Common Config Structs

| Struct | Description |
|---|---|
| `AppConfig` | Environment, app name, version, log level |
| `MysqlConfig` | MySQL connection with `ConnectionString()` and client constructor |
| `PostgresConfig` | Postgres connection with pool settings and `ConnectionString()` |
| `RedisConfig` | Redis connection with TLS and timeout settings |
| `FirestoreConfig` | Firestore API key and client constructor |
| `OpenTelemetryConfig` | OTLP exporter endpoint, TLS, and sample rate |

## Legacy Usage

The `Process` function loads `.env` files from a directory and processes env vars into a config struct:

```go
var cfg Config
if err := config.Process(os.Getenv("ENV_FILES_DIR"), &cfg); err != nil {
    panic(err)
}
```
