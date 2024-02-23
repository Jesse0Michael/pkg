# Logger 

Standardizing logging for GO applications

Uses [slog](https://pkg.go.dev/log/slog) from the standard library to have consistent structured logging across applications.

The logger should be configured and set up in the main function of the application and accessed through the slog package.

## Configuration
The logger will take the following environment variables to configure the logger.
| Environment Variable | Description | Default |
| --- | --- | --- |
| LOG_LEVEL | The log level to use. (DEBUG, INFO, WARN, ERROR) | INFO |
| LOG_OUTPUT | The output to write the logs to. (STDOUT, STDERR) | stdout |
| HOSTNAME | The hostname to set in the logs. |  |
| ENVIRONMENT | The environment to set in the logs. |  |

## Usage

``` go
import (
    "errors"
    "log/slog"

    "github.com/jesse0michael/pkg/logger"
)

func main() {
    logger.SetLog()
    ctx := context.Background()

    // ...


    slog.With("key", "value").InfoContext(ctx, "writing logs")
    // {"time":"2023-08-21T09:15:18.111628-07:00","level":"INFO","msg":"writing logs","host":"local","environment":"test","key":"value"}

    slog.With("error", errors.New("test-error")).ErrorContext(ctx, "writing errors") 
    // {"time":"2023-08-21T09:15:18.111716-07:00","level":"ERROR","msg":"writing errors","host":"local","environment":"test","error":"test-error"}
}
```

## Context Handler
The default logger is configured with a [ContextHandler](context_handler.go) that allow you to set attributes in the GO context that will be logged with every log message.

``` go
import (
    "errors"
    "log/slog"

    "github.com/jesse0michael/pkg/logger"
)

func main() {
    _ = logger.NewLogger()
    ctx := context.Background()
    ctx = logger.AddAttrs(ctx, slog.String("accountID", "12345"))

    // ...

    slog.InfoContext(ctx, "writing logs")
    // {"time":"2023-08-21T09:15:18.111628-07:00","level":"INFO","msg":"writing logs","host":"local","environment":"test","accountID":"12345"}

    slog.With("error", errors.New("test-error")).ErrorContext(ctx, "writing errors") 
    // {"time":"2023-08-21T09:15:18.111628-07:00","level":"ERROR","msg":"writing errors","host":"local","environment":"test","error":"test-error","accountID":"12345"}
}
```

## Error Warn Handler
The logger can be configured with an [ErrorWarnHandler](errorwarnhandler.go) that will log errors as warnings if an `error` attribute matches a specified function.

``` go
import (
    "context"
    "errors"
    "log/slog"

    "github.com/jesse0michael/pkg/logger"
)

func main() {
    warnCheck := func(err error) bool { return err == context.Canceled }
    logger.NewLogger()
    logger.SetErrorWarnHandler(warnCheck)
    ctx := context.Background()

    // ...

    slog.With("error", context.Canceled).ErrorContext(ctx, "writing errors") 
    // {"time":"2023-08-21T09:15:18.111628-07:00","level":"WARN","msg":"writing errors","host":"local","environment":"test","error":"context canceled"}
}
```
