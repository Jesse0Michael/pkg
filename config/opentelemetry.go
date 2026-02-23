package config

import (
	"context"
	"fmt"
	"log/slog"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/contrib/instrumentation/host"
	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

// OpenTelemetryConfig is the configuration for the OpenTelemetry exporters.
// It uses the environment variables defined by the OpenTelemetry Collector
// but with different defaults, to make it easier to integrate into services.
// https://github.com/open-telemetry/opentelemetry-specification/blob/v1.20.0/specification/protocol/exporter.md
type OpenTelemetryConfig struct {
	OpenTelemetryEndpoint   string  `envconfig:"OTEL_EXPORTER_OTLP_ENDPOINT" default:"localhost:4317"`
	OpenTelemetryInsecure   bool    `envconfig:"OTEL_EXPORTER_OTLP_INSECURE" default:"true"`
	OpenTelemetrySampleRate float64 `envconfig:"OTEL_TRACES_SAMPLER_ARG" default:"1.0"`
}

func (cfg OpenTelemetryConfig) MetricOptions() []otlpmetricgrpc.Option {
	opts := []otlpmetricgrpc.Option{
		otlpmetricgrpc.WithEndpoint(cfg.OpenTelemetryEndpoint),
	}
	if cfg.OpenTelemetryInsecure {
		opts = append(opts, otlpmetricgrpc.WithInsecure())
	}
	return opts
}

func (cfg OpenTelemetryConfig) TracerOptions() []otlptracegrpc.Option {
	opts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(cfg.OpenTelemetryEndpoint),
	}
	if cfg.OpenTelemetryInsecure {
		opts = append(opts, otlptracegrpc.WithInsecure())
	}
	return opts
}

func (cfg OpenTelemetryConfig) LogOptions() []otlploggrpc.Option {
	opts := []otlploggrpc.Option{
		otlploggrpc.WithEndpoint(cfg.OpenTelemetryEndpoint),
	}
	if cfg.OpenTelemetryInsecure {
		opts = append(opts, otlploggrpc.WithInsecure())
	}
	return opts
}

// OtelResource creates a otel resource from an app config.
// The purpose of this function standardize on the way open telemetry is configured
// and not have to repeat the same boilerplate code in every service
// or end up in a situation where every service configures open telemetry differently.
func OtelResource(ctx context.Context, cfg AppConfig, attributes ...attribute.KeyValue) (*resource.Resource, error) {
	attributes = append(attributes,
		semconv.ServiceName(cfg.Name),
		semconv.ServiceVersion(cfg.Version),
		semconv.DeploymentEnvironment(cfg.Environment))

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	return resource.New(ctx,
		resource.WithAttributes(attributes...),
		resource.WithContainer(),
		resource.WithHost(),
	)
}

// OtelTraceProvider creates and sets an otel trace provider from an otel config and resource.
// The purpose of this function is to standardize on the way open telemetry is configured
// and not have to repeat the same boilerplate code in every service
// or end up in a situation where every service configures open telemetry differently.
func OtelTraceProvider(ctx context.Context, cfg OpenTelemetryConfig, r *resource.Resource,
) (*trace.TracerProvider, error) {
	traceExporter, err := otlptracegrpc.New(ctx, cfg.TracerOptions()...)
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	tp := trace.NewTracerProvider(
		trace.WithSampler(trace.ParentBased(trace.TraceIDRatioBased(cfg.OpenTelemetrySampleRate))),
		trace.WithResource(r),
		trace.WithSpanProcessor(trace.NewBatchSpanProcessor(traceExporter)),
	)
	otel.SetTracerProvider(tp)

	return tp, nil
}

// OtelMeterProvider creates and sets an otel meter provider from an otel config and resource.
// The purpose of this function is to standardize on the way open telemetry is configured
// and not have to repeat the same boilerplate code in every service
// or end up in a situation where every service configures open telemetry differently.
func OtelMeterProvider(ctx context.Context, cfg OpenTelemetryConfig, r *resource.Resource,
) (*metric.MeterProvider, error) {
	metricExporter, err := otlpmetricgrpc.New(ctx, cfg.MetricOptions()...)
	if err != nil {
		return nil, fmt.Errorf("failed to create metric exporter: %w", err)
	}

	mp := metric.NewMeterProvider(
		metric.WithResource(r),
		metric.WithReader(metric.NewPeriodicReader(metricExporter)),
	)
	otel.SetMeterProvider(mp)

	if err := runtime.Start(); err != nil {
		return mp, fmt.Errorf("failed to start runtime metric collection: %w", err)
	}

	if err := host.Start(); err != nil {
		return mp, fmt.Errorf("failed to start host metric collection: %w", err)
	}
	return mp, nil
}

func OtelLogProvider(ctx context.Context, cfg OpenTelemetryConfig, r *resource.Resource) (*log.LoggerProvider, error) {
	logExporter, err := otlploggrpc.New(ctx, cfg.LogOptions()...)
	if err != nil {
		return nil, fmt.Errorf("failed to create log exporter: %w", err)
	}

	provider := log.NewLoggerProvider(
		log.WithResource(r),
		log.WithProcessor(log.NewBatchProcessor(logExporter)),
	)

	handler := otelslog.NewHandler("",
		otelslog.WithLoggerProvider(provider),
		otelslog.WithSource(true),
	)

	slog.SetDefault(slog.New(
		slog.NewMultiHandler(
			slog.Default().Handler(),
			handler,
		),
	))

	return provider, nil
}
