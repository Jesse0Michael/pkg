package config

import (
	"fmt"
	"reflect"
	"testing"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

func TestOpenTelemetryConfig_MetricOptions(t *testing.T) {
	tests := []struct {
		name string
		cfg  OpenTelemetryConfig
		want []otlpmetricgrpc.Option
	}{
		{
			name: "empty config",
			cfg:  OpenTelemetryConfig{},
			want: []otlpmetricgrpc.Option{
				otlpmetricgrpc.WithEndpoint(""),
			},
		},
		{
			name: "insecure config",
			cfg: OpenTelemetryConfig{
				OpenTelemetryEndpoint: "localhost:4317",
				OpenTelemetryInsecure: true,
			},
			want: []otlpmetricgrpc.Option{
				otlpmetricgrpc.WithEndpoint("localhost:4317"),
				otlpmetricgrpc.WithInsecure(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cfg.MetricOptions(); len(tt.want) == len(got) {
				for i, v := range tt.want {
					fmt.Println(reflect.TypeOf(v))
					if reflect.TypeOf(v) != reflect.TypeOf(got[i]) {
						t.Errorf("OpenTelemetryConfig.MetricOptions() = %v, want %v", got, tt.want)
					}
				}
			} else {
				t.Errorf("OpenTelemetryConfig.MetricOptions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOpenTelemetryConfig_TracerOptions(t *testing.T) {
	tests := []struct {
		name string
		cfg  OpenTelemetryConfig
		want []otlptracegrpc.Option
	}{
		{
			name: "empty config",
			cfg:  OpenTelemetryConfig{},
			want: []otlptracegrpc.Option{
				otlptracegrpc.WithEndpoint(""),
			},
		},
		{
			name: "insecure config",
			cfg: OpenTelemetryConfig{
				OpenTelemetryEndpoint: "localhost:4317",
				OpenTelemetryInsecure: true,
			},
			want: []otlptracegrpc.Option{
				otlptracegrpc.WithEndpoint("localhost:4317"),
				otlptracegrpc.WithInsecure(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cfg.TracerOptions(); len(tt.want) == len(got) {
				for i, v := range tt.want {
					fmt.Println(reflect.TypeOf(v))
					if reflect.TypeOf(v) != reflect.TypeOf(got[i]) {
						t.Errorf("OpenTelemetryConfig.TracerOptions() = %v, want %v", got, tt.want)
					}
				}
			} else {
				t.Errorf("OpenTelemetryConfig.TracerOptions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOpenTelemetryConfig_LogOptions(t *testing.T) {
	tests := []struct {
		name string
		cfg  OpenTelemetryConfig
		want []otlploggrpc.Option
	}{
		{
			name: "empty config",
			cfg:  OpenTelemetryConfig{},
			want: []otlploggrpc.Option{
				otlploggrpc.WithEndpoint(""),
			},
		},
		{
			name: "insecure config",
			cfg: OpenTelemetryConfig{
				OpenTelemetryEndpoint: "localhost:4317",
				OpenTelemetryInsecure: true,
			},
			want: []otlploggrpc.Option{
				otlploggrpc.WithEndpoint("localhost:4317"),
				otlploggrpc.WithInsecure(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cfg.LogOptions(); len(tt.want) == len(got) {
				for i, v := range tt.want {
					fmt.Println(reflect.TypeOf(v))
					if reflect.TypeOf(v) != reflect.TypeOf(got[i]) {
						t.Errorf("OpenTelemetryConfig.LogOptions() = %v, want %v", got, tt.want)
					}
				}
			} else {
				t.Errorf("OpenTelemetryConfig.LogOptions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_otelResource(t *testing.T) {
	tests := []struct {
		name       string
		cfg        AppConfig
		attributes []attribute.KeyValue
		want       *resource.Resource
		wantErr    bool
	}{
		{
			name:       "empty config",
			cfg:        AppConfig{},
			attributes: []attribute.KeyValue{},
			want: func() *resource.Resource {
				r, _ := resource.New(t.Context(),
					resource.WithAttributes(
						semconv.ServiceName(""),
						semconv.ServiceVersion(""),
						semconv.DeploymentEnvironment("")),
					resource.WithContainer(),
					resource.WithHost(),
				)
				return r
			}(),
			wantErr: false,
		},
		{
			name:       "with config",
			cfg:        AppConfig{Environment: "local", Name: "app", Version: "1.0.0", LogLevel: "debug"},
			attributes: []attribute.KeyValue{},
			want: func() *resource.Resource {
				r, _ := resource.New(t.Context(),
					resource.WithAttributes(
						semconv.ServiceName("app"),
						semconv.ServiceVersion("1.0.0"),
						semconv.DeploymentEnvironment("local")),
					resource.WithContainer(),
					resource.WithHost(),
				)
				return r
			}(),
			wantErr: false,
		},
		{
			name:       "with attributes",
			cfg:        AppConfig{Environment: "local", Name: "app", Version: "1.0.0", LogLevel: "debug"},
			attributes: []attribute.KeyValue{attribute.String("test.id", "test")},
			want: func() *resource.Resource {
				r, _ := resource.New(t.Context(),
					resource.WithAttributes(
						attribute.String("test.id", "test"),
						semconv.ServiceName("app"),
						semconv.ServiceVersion("1.0.0"),
						semconv.DeploymentEnvironment("local")),
					resource.WithContainer(),
					resource.WithHost(),
				)
				return r
			}(),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := OtelResource(t.Context(), tt.cfg, tt.attributes...)
			if (err != nil) != tt.wantErr {
				t.Errorf("OtelResource() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("OtelResource() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOtelTraceProvider(t *testing.T) {
	tests := []struct {
		name    string
		cfg     OpenTelemetryConfig
		wantTP  bool
		wantErr bool
	}{
		{
			name: "successful trace provider",
			cfg: OpenTelemetryConfig{
				OpenTelemetryEndpoint: "localhost:4317",
			},
			wantTP:  true,
			wantErr: false,
		},
		{
			name:    "failed trace provider",
			cfg:     OpenTelemetryConfig{},
			wantTP:  false,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := OtelTraceProvider(t.Context(), tt.cfg, &resource.Resource{})
			if (err != nil) != tt.wantErr {
				t.Errorf("OtelTraceProvider() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantTP {
				tp := otel.GetTracerProvider()
				if !reflect.DeepEqual(got, tp) {
					t.Errorf("OtelTraceProvider() = %v, want %v", got, tp)
				}
			} else if got != nil {
				t.Errorf("OtelTraceProvider() tp = %v, want %v", got, nil)
			}
		})
	}
}

func TestOtelMeterProvider(t *testing.T) {
	tests := []struct {
		name    string
		cfg     OpenTelemetryConfig
		wantMP  bool
		wantErr bool
	}{
		{
			name: "successful meter provider",
			cfg: OpenTelemetryConfig{
				OpenTelemetryEndpoint: "localhost:4317",
			},
			wantMP:  true,
			wantErr: false,
		},
		{
			name:    "failed meter provider",
			cfg:     OpenTelemetryConfig{},
			wantMP:  false,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &resource.Resource{}
			got, err := OtelMeterProvider(t.Context(), tt.cfg, r)
			if (err != nil) != tt.wantErr {
				t.Errorf("OtelMeterProvider() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantMP {
				mp := otel.GetMeterProvider()
				if !reflect.DeepEqual(got, mp) {
					t.Errorf("OtelMeterProvider() = %v, want %v", got, mp)
				}
			} else if got != nil {
				t.Errorf("OtelMeterProvider() mp = %v, want %v", got, nil)
			}
		})
	}
}

func TestOtelLogProvider(t *testing.T) {
	tests := []struct {
		name    string
		cfg     OpenTelemetryConfig
		wantErr bool
	}{
		{
			name: "successful log provider",
			cfg: OpenTelemetryConfig{
				OpenTelemetryEndpoint: "localhost:4317",
			},
			wantErr: false,
		},
		{
			name:    "empty config",
			cfg:     OpenTelemetryConfig{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := OtelLogProvider(t.Context(), tt.cfg, &resource.Resource{})
			if (err != nil) != tt.wantErr {
				t.Errorf("OtelLogProvider() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got == nil {
				t.Errorf("OtelLogProvider() = %v, want non-nil", got)
			}
		})
	}
}
