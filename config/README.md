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

**`Init()` < struct defaults < env vars < config files (in order) < CLI args**

Each layer overwrites values set by previous layers. When multiple config files are provided, they are applied in order so later files override earlier ones.

### Initializer

Config structs can implement the `Initializer` interface to set dynamic initial values before any other source is applied:

```go
type Initializer interface {
    Init()
}
```

`AppConfig` implements this to populate `Name` and `Version` from Go build info (`runtime/debug.ReadBuildInfo`), giving you the module name and VCS revision automatically without env vars.

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
| `AppConfig` | Environment, app name, version — auto-populated from build info via `Init()` |
| `MysqlConfig` | MySQL connection with `ConnectionString()` and client constructor |
| `PostgresConfig` | Postgres connection with pool settings and `ConnectionString()` |
| `RedisConfig` | Redis connection with TLS and timeout settings |
| `AWSConfig` | AWS SDK region, profile and static credentials with `NewAWSConfig` |
| `S3Config` | AWS S3 or S3-compatible bucket (MinIO, Cloudflare R2, Google Cloud Storage, ...) with `NewS3Client` and `NewS3PresignClient` |
| `FirestoreConfig` | Firestore API key and client constructor |
| `OpenTelemetryConfig` | OTLP exporter endpoint, TLS, and sample rate |

### AWSConfig

`NewAWSConfig(ctx, cfg)` returns the `aws.Config` shared by every AWS client. Every field is an optional override; anything left empty is resolved by the SDK's default chain (`AWS_*` env vars, shared config and credentials files, SSO, web identity roles, instance roles).

| Variable | Description |
|---|---|
| `AWS_REGION` | Region |
| `AWS_PROFILE` | Shared config profile |
| `AWS_ACCESS_KEY_ID` / `AWS_SECRET_ACCESS_KEY` | Static credentials, used only when both are set |

### S3Config

`NewS3Client(awsCfg, cfg)` returns an `*s3.Client` and `NewS3PresignClient(awsCfg, cfg)` an `*s3.PresignClient` for AWS S3 or any S3-compatible store. Setting an endpoint switches to path-style addressing and sends checksums only when required, which S3-compatible stores need.

| Variable | Description |
|---|---|
| `S3_BUCKET` | Bucket name |
| `AWS_ENDPOINT_URL_S3` | Endpoint of an S3-compatible store; empty means AWS S3 |
| `S3_REGION` | Region override for the bucket; required with an endpoint |
| `S3_PUBLIC_ENDPOINT` | Endpoint browsers reach the store at, used to sign object URLs; defaults to the endpoint |

| Store | `AWS_ENDPOINT_URL_S3` | `S3_REGION` |
|---|---|---|
| AWS S3 | | bucket region |
| MinIO | `http://localhost:9000` | `us-east-1` |
| Cloudflare R2 | `https://<account>.r2.cloudflarestorage.com` | `auto` |
| Google Cloud Storage (HMAC keys) | `https://storage.googleapis.com` | `auto` |

Signed URLs only work through a proxy that forwards the `Host` header unchanged, since it is part of the signature. R2 signs only its `r2.cloudflarestorage.com` host, not custom domains.

## Legacy Usage

The `Process` function loads `.env` files from a directory and processes env vars into a config struct:

```go
var cfg Config
if err := config.Process(os.Getenv("ENV_FILES_DIR"), &cfg); err != nil {
    panic(err)
}
```
