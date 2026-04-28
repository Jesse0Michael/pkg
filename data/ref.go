package data

import (
	"context"
	"reflect"
	"sync"
	"time"
)

// RefOption configures a Ref.
type RefOption[T any] func(*Ref[T])

// WithInitialValue sets the initial value of the Ref before the first refresh.
func WithInitialValue[T any](v T) RefOption[T] {
	return func(r *Ref[T]) {
		r.value = v
	}
}

// WithOnError sets a callback invoked when the refresh function returns an error.
func WithOnError[T any](f func(error)) RefOption[T] {
	return func(r *Ref[T]) {
		r.onError = f
	}
}

// WithInterval overrides the default 10s refresh interval.
func WithInterval[T any](d time.Duration) RefOption[T] {
	return func(r *Ref[T]) {
		r.interval = d
	}
}

// WithOnChange sets a callback invoked when the refreshed value differs from the previous one.
func WithOnChange[T any](f func(T)) RefOption[T] {
	return func(r *Ref[T]) {
		r.onChange = f
	}
}

// WithLazy disables the background ticker. Instead, Load() checks whether the
// cached value is stale (older than the configured interval) and triggers an
// async refresh, returning the current value immediately.
func WithLazy[T any]() RefOption[T] {
	return func(r *Ref[T]) {
		r.lazy = true
	}
}

// Ref holds a value of type T that is periodically refreshed by calling a function.
// It is safe for concurrent reads.
type Ref[T any] struct {
	mu          sync.RWMutex
	value       T
	interval    time.Duration
	fetch       func(context.Context) (T, error)
	onError     func(error)
	onChange    func(T)
	cancel      context.CancelFunc
	lazy        bool
	lastRefresh time.Time
	ctx         context.Context
}

// NewRef creates a Ref that refreshes its value by calling fetch at a regular interval (default 10s).
// The refresh loop runs in a goroutine until ctx is cancelled or Stop is called.
func NewRef[T any](ctx context.Context, fetch func(context.Context) (T, error), opts ...RefOption[T]) *Ref[T] {
	r := &Ref[T]{
		interval: 10 * time.Second,
		fetch:    fetch,
	}
	for _, o := range opts {
		o(r)
	}
	ctx, r.cancel = context.WithCancel(ctx)
	r.ctx = ctx
	if !r.lazy {
		go r.start(ctx)
	}
	return r
}

func (r *Ref[T]) start(ctx context.Context) {
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r.refresh(ctx)
		}
	}
}

// Stop cancels the refresh loop.
func (r *Ref[T]) Stop() {
	if r.cancel != nil {
		r.cancel()
	}
}

// Load returns the current value. In lazy mode, if the cached value is stale
// it triggers an async refresh and returns the current value immediately.
func (r *Ref[T]) Load() T {
	r.mu.RLock()
	v := r.value
	stale := r.lazy && time.Since(r.lastRefresh) > r.interval
	r.mu.RUnlock()
	if stale {
		go r.refresh(r.ctx)
	}
	return v
}

func (r *Ref[T]) refresh(ctx context.Context) {
	v, err := r.fetch(ctx)
	r.mu.Lock()
	r.lastRefresh = time.Now()
	if err != nil {
		r.mu.Unlock()
		if r.onError != nil {
			r.onError(err)
		}
		return
	}
	old := r.value
	r.value = v
	r.mu.Unlock()
	if r.onChange != nil && !reflect.DeepEqual(old, v) {
		r.onChange(v)
	}
}
