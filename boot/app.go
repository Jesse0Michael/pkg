package boot

import (
	"context"
	"errors"
	"log/slog"
	"sync"

	"github.com/jesse0michael/pkg/config"
	"github.com/jesse0michael/pkg/http/handlers"
	"github.com/jesse0michael/pkg/logger"
)

type Runner[T any] interface {
	Run(ctx context.Context, cfg T) error
	Close() error
}

type App[T any] struct {
	ctx     context.Context
	cancel  context.CancelCauseFunc
	cfg     T
	closers []func(context.Context) error
}

func NewApp[T any]() *App[T] {
	logger.NewLogger()

	ctx, cancel := Context()
	cfg, err := config.New[T]()
	if err != nil {
		cancel(err)
	}

	app := &App[T]{
		ctx:    ctx,
		cancel: cancel,
		cfg:    cfg,
	}

	appConfig, hasApp := structHas[config.AppConfig](cfg)
	otelConfig, hasOtel := structHas[config.OpenTelemetryConfig](cfg)
	if hasApp && hasOtel {
		resource, err := config.OtelResource(ctx, appConfig)
		if err != nil {
			slog.ErrorContext(ctx, "failed to create otel resource", "err", err)
			cancel(err)
		}

		lp, err := config.OtelLogProvider(ctx, otelConfig, resource)
		if err != nil {
			slog.ErrorContext(ctx, "failed to create otel log provider", "err", err)
			cancel(err)
		} else {
			app.closers = append(app.closers, lp.Shutdown)
		}

		tp, err := config.OtelTraceProvider(ctx, otelConfig, resource)
		if err != nil {
			slog.ErrorContext(ctx, "failed to create otel trace provider", "err", err)
			cancel(err)
		} else {
			app.closers = append(app.closers, tp.Shutdown)
		}

		mp, err := config.OtelMeterProvider(ctx, otelConfig, resource)
		if err != nil {
			slog.ErrorContext(ctx, "failed to create otel meter provider", "err", err)
			cancel(err)
		} else {
			app.closers = append(app.closers, mp.Shutdown)
		}

		slog.InfoContext(ctx, "OpenTelemetry initialized", "resource", resource)
	}

	return app
}

func (a *App[T]) Context() context.Context {
	return a.ctx
}

func (a *App[T]) Cancel(cause error) {
	a.cancel(cause)
}

func (a *App[T]) Run(runners ...Runner[T]) error {
	if err := a.ctx.Err(); err != nil {
		return err
	}

	var wg sync.WaitGroup
	for _, runner := range runners {
		wg.Add(1)
		go func(r Runner[T]) {
			defer wg.Done()
			if err := r.Run(a.ctx, a.cfg); err != nil {
				a.cancel(err)
			}
		}(runner)
	}
	wg.Wait()

	handlers.ServeHealthCheckMetrics(a.ctx)

	<-a.ctx.Done()
	if err := a.ctx.Err(); err != nil && !errors.Is(err, context.Canceled) {
		slog.Error("app done with error", "error", err)
		return err
	}

	for _, r := range runners {
		r.Close()
	}

	for _, shutdown := range a.closers {
		shutdownCtx := context.Background()
		if err := shutdown(shutdownCtx); err != nil {
			slog.Error("failed to shutdown provider", "err", err)
		}
	}

	slog.Info("exiting")
	return nil
}
