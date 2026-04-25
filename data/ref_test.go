package data

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"
)

func TestRef_InitialValue(t *testing.T) {
	ref := NewRef(t.Context(), func(ctx context.Context) (string, error) {
		return "test-fetched", nil
	}, WithInitialValue("test-initial"), WithInterval[string](time.Hour))

	if got := ref.Load(); got != "test-initial" {
		t.Errorf("Load() = %q, want %q", got, "test-initial")
	}
}

func TestRef_DefaultInterval(t *testing.T) {
	ref := NewRef(t.Context(), func(ctx context.Context) (string, error) {
		return "test-value", nil
	})

	if ref.interval != 10*time.Second {
		t.Errorf("default interval = %v, want 10s", ref.interval)
	}
}

func TestRef_Refresh(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var count atomic.Int32
		ref := NewRef(t.Context(), func(ctx context.Context) (int, error) {
			return int(count.Add(1)), nil
		}, WithInterval[int](10*time.Millisecond))

		time.Sleep(25 * time.Millisecond)
		synctest.Wait()

		if got := ref.Load(); got < 2 {
			t.Errorf("Load() = %d after 2 ticks, want >= 2", got)
		}
	})
}

func TestRef_OnError(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var gotErr atomic.Value
		ref := NewRef(t.Context(), func(ctx context.Context) (string, error) {
			return "", errors.New("test-error")
		},
			WithInitialValue("test-initial"),
			WithInterval[string](10*time.Millisecond),
			WithOnError[string](func(err error) {
				gotErr.Store(err)
			}),
		)

		time.Sleep(15 * time.Millisecond)
		synctest.Wait()

		if gotErr.Load() == nil {
			t.Error("onError was not called")
		}
		// Value should remain the initial value on error.
		if got := ref.Load(); got != "test-initial" {
			t.Errorf("Load() = %q, want %q", got, "test-initial")
		}
	})
}

func TestRef_OnChange(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var count atomic.Int32
		var changed atomic.Value
		ref := NewRef(t.Context(), func(ctx context.Context) (int, error) {
			return int(count.Add(1)), nil
		},
			WithInitialValue(0),
			WithInterval[int](10*time.Millisecond),
			WithOnChange[int](func(v int) {
				changed.Store(v)
			}),
		)
		_ = ref

		time.Sleep(15 * time.Millisecond)
		synctest.Wait()

		if changed.Load() == nil {
			t.Error("onChange was not called")
		}
	})
}

func TestRef_OnChange_NotCalledWhenUnchanged(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var calls atomic.Int32
		ref := NewRef(t.Context(), func(ctx context.Context) (string, error) {
			return "test-static", nil
		},
			WithInitialValue("test-static"),
			WithInterval[string](10*time.Millisecond),
			WithOnChange[string](func(v string) {
				calls.Add(1)
			}),
		)
		_ = ref

		time.Sleep(25 * time.Millisecond)
		synctest.Wait()

		if got := calls.Load(); got != 0 {
			t.Errorf("onChange called %d times, want 0 (value did not change)", got)
		}
	})
}

func TestRef_Stop(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var count atomic.Int32
		ref := NewRef(t.Context(), func(ctx context.Context) (int, error) {
			return int(count.Add(1)), nil
		}, WithInterval[int](10*time.Millisecond))

		time.Sleep(15 * time.Millisecond)
		synctest.Wait()
		if got := count.Load(); got < 1 {
			t.Fatalf("count = %d after 1 tick, want >= 1", got)
		}

		ref.Stop()
		stopped := count.Load()

		// After Stop, the count should not continue to increase.
		time.Sleep(time.Second)
		synctest.Wait()
		if got := count.Load(); got != stopped {
			t.Errorf("count = %d, want %d (no increase after Stop)", got, stopped)
		}
	})
}
