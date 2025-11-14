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
	ctx    context.Context
	cancel context.CancelCauseFunc
	cfg    T
}

func NewApp[T any]() *App[T] {
	logger.NewLogger()

	ctx, cancel := Context()

	cfg, err := config.New[T]()
	if err != nil {
		cancel(err)
	}

	return &App[T]{
		ctx:    ctx,
		cancel: cancel,
		cfg:    *cfg,
	}
}

func (a *App[T]) Context() context.Context {
	return a.ctx
}

func (a *App[T]) Cancel(cause error) {
	a.cancel(cause)
}

func (a *App[T]) Run(runners ...Runner[T]) {
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

	handlers.ServeHealthCheckMetrics(a.ctx)

	<-a.ctx.Done()
	if err := a.ctx.Err(); err != nil && !errors.Is(err, context.Canceled) {
		slog.Error("app done with error", "error", err)
	}

	for _, r := range runners {
		r.Close()
	}

	wg.Wait()
	slog.Info("exiting")
}
