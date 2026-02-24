package boot

import (
	"context"
	"log/slog"

	"github.com/jesse0michael/pkg/config"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
)

// SetupOpenTelemetry initializes OpenTelemetry resources, log provider, trace provider,
// and meter provider based on the provided configuration.
func SetupOpenTelemetry(ctx context.Context, appConfig config.AppConfig, otelConfig config.OpenTelemetryConfig) (
	*resource.Resource,
	*log.LoggerProvider,
	*trace.TracerProvider,
	*metric.MeterProvider,
	error,
) {
	res, err := config.OtelResource(ctx, appConfig)
	if err != nil {
		slog.ErrorContext(ctx, "failed to create otel resource", "err", err)
		return nil, nil, nil, nil, err
	}

	lp, err := config.OtelLogProvider(ctx, otelConfig, res)
	if err != nil {
		slog.ErrorContext(ctx, "failed to create otel log provider", "err", err)
		return nil, nil, nil, nil, err
	}

	tp, err := config.OtelTraceProvider(ctx, otelConfig, res)
	if err != nil {
		slog.ErrorContext(ctx, "failed to create otel trace provider", "err", err)
		return nil, nil, nil, nil, err
	}

	mp, err := config.OtelMeterProvider(ctx, otelConfig, res)
	if err != nil {
		slog.ErrorContext(ctx, "failed to create otel meter provider", "err", err)
		return nil, nil, nil, nil, err
	}

	return res, lp, tp, mp, nil
}
