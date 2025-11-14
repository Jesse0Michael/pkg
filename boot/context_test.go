package boot

import (
	"context"
	"errors"
	"os"
	"syscall"
	"testing"
	"time"
)

func TestContext_Signal(t *testing.T) {
	ctx, _ := Context()

	if ctx == nil {
		t.Error("Context() should return a non-nil context")
	}

	select {
	case <-ctx.Done():
		t.Error("context should not be done initially")
	default:
	}

	// Send signal to current process
	pid := os.Getpid()
	process, err := os.FindProcess(pid)
	if err != nil {
		t.Fatalf("failed to find current process: %v", err)
	}

	// Send SIGTERM signal
	err = process.Signal(syscall.SIGTERM)
	if err != nil {
		t.Fatalf("failed to send signal: %v", err)
	}

	select {
	case <-ctx.Done():
		// context was properly cancelled by signal
		if ctx.Err() != context.Canceled {
			t.Errorf("expected context.Canceled, got %v", ctx.Err())
		}
	case <-time.After(1 * time.Second):
		t.Error("context should have been cancelled by signal within 1 second")
	}
}

func TestContext_CancelFunc(t *testing.T) {
	// Get context from the function
	ctx, cancel := Context()

	if ctx == nil {
		t.Error("Context() should return a non-nil context")
	}

	select {
	case <-ctx.Done():
		t.Error("context should not be done initially")
	default:
	}

	err := errors.New("test-error")
	cancel(err)

	// Check that context is done
	select {
	case <-ctx.Done():
		// context was properly cancelled
		if context.Cause(ctx) != err {
			t.Errorf("expected context cause to be %v, got %v", err, context.Cause(ctx))
		}
		if ctx.Err() != context.Canceled {
			t.Errorf("expected context error to be context.Canceled, got %v", ctx.Err())
		}
	case <-time.After(10 * time.Millisecond):
		t.Error("context should have been cancelled immediately")
	}
}
