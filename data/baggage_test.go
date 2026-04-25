package data

import (
	"testing"

	"go.opentelemetry.io/otel/baggage"
)

func TestSetBaggage(t *testing.T) {
	tests := []struct {
		name     string
		args     []any
		wantKeys map[string]string
	}{
		{
			name:     "string key-value pair",
			args:     []any{"test-key", "test-value"},
			wantKeys: map[string]string{"test-key": "test-value"},
		},
		{
			name:     "int value",
			args:     []any{"test-key", 42},
			wantKeys: map[string]string{"test-key": "42"},
		},
		{
			name:     "bool value",
			args:     []any{"test-key", true},
			wantKeys: map[string]string{"test-key": "true"},
		},
		{
			name:     "float value",
			args:     []any{"test-key", 3.14},
			wantKeys: map[string]string{"test-key": "3.14"},
		},
		{
			name:     "multiple pairs",
			args:     []any{"test-key-1", "test-value-1", "test-key-2", 42},
			wantKeys: map[string]string{"test-key-1": "test-value-1", "test-key-2": "42"},
		},
		{
			name:     "baggage.Member argument",
			args:     []any{must(baggage.NewMemberRaw("test-key", "test-value"))},
			wantKeys: map[string]string{"test-key": "test-value"},
		},
		{
			name:     "mixed pairs and members",
			args:     []any{"test-key-1", "test-value-1", must(baggage.NewMemberRaw("test-key-2", "test-value-2"))},
			wantKeys: map[string]string{"test-key-1": "test-value-1", "test-key-2": "test-value-2"},
		},
		{
			name:     "no args",
			args:     nil,
			wantKeys: map[string]string{},
		},
		{
			name:     "unpaired key is skipped",
			args:     []any{"test-key"},
			wantKeys: map[string]string{},
		},
		{
			name:     "unsupported type is skipped",
			args:     []any{123, "test-key", "test-value"},
			wantKeys: map[string]string{"test-key": "test-value"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := SetBaggage(t.Context(), tt.args...)

			b := baggage.FromContext(ctx)
			for k, v := range tt.wantKeys {
				if got := b.Member(k).Value(); got != v {
					t.Errorf("baggage[%q] = %q, want %q", k, got, v)
				}
			}
		})
	}
}

func TestSetBaggage_PreservesExisting(t *testing.T) {
	ctx := SetBaggage(t.Context(), "test-key-1", "test-value-1")
	ctx = SetBaggage(ctx, "test-key-2", "test-value-2")

	b := baggage.FromContext(ctx)
	if got := b.Member("test-key-1").Value(); got != "test-value-1" {
		t.Errorf(`baggage["test-key-1"] = %q, want "test-value-1"`, got)
	}
	if got := b.Member("test-key-2").Value(); got != "test-value-2" {
		t.Errorf(`baggage["test-key-2"] = %q, want "test-value-2"`, got)
	}
}

func must(m baggage.Member, err error) baggage.Member {
	if err != nil {
		panic(err)
	}
	return m
}
