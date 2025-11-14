package boot

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func Context() (context.Context, context.CancelCauseFunc) {
	ctx, cancel := context.WithCancelCause(context.Background())
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-sig
		slog.Info("termination signaled")
		cancel(nil)
	}()

	return ctx, cancel
}
