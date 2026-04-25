# data

Generic, concurrency-safe data structures and OpenTelemetry baggage helpers for Go.

## Map

A type-safe concurrent map with first-class support for the `maps`, `slices`, and `iter` packages.

```go
m := data.NewMap[string, int]()
m.Set("alpha", 1)
m.Set("beta", 2)
m.Set("gamma", 3)

v, ok := m.Get("alpha") // 1, true

// range over all key-value pairs
for k, v := range m.All() {
    fmt.Println(k, v)
}

// collect into a plain map
snapshot := maps.Collect(m.All())

// collect sorted keys
keys := slices.Sorted(m.Keys())

// collect values
values := slices.Collect(m.Values())
```

## Ref

A self-refreshing reference that periodically calls a function to keep its value current.

```go
ref := data.NewRef[Config](ctx, fetchConfig,
    data.WithInterval[Config](30*time.Second),
    data.WithInitialValue[Config](defaultCfg),
    data.WithOnError[Config](func(err error) {
        slog.Error("config refresh failed", "err", err)
    }),
)

cfg := ref.Value() // always up-to-date
```

## Baggage

Helpers for writing typed values into OpenTelemetry baggage.

```go
ctx = data.SetBaggage(ctx, "user_id", 12345, "tenant", "acme", "region", "us-east-1")
```

To propagate baggage into logs, wrap your `slog.Handler` with [`logger.NewBaggageHandler`](../logger/baggagehandler.go), which copies baggage members from the context onto each log record. For traces, use the [baggagecopy](https://pkg.go.dev/go.opentelemetry.io/contrib/processors/baggagecopy) processors from the OTel contrib library.

## Usage

```bash
go get github.com/jesse0michael/pkg/data
```
