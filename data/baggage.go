package data

import (
	"context"
	"fmt"
	"log/slog"

	"go.opentelemetry.io/otel/baggage"
)

// SetBaggage returns a new context with the provided key-value pairs added to the OTel baggage.
// All values are converted to strings using fmt.Sprint.
//
// Arguments can be provided in two formats:
// 1. Alternating string keys and values: SetBaggage(ctx, "user_id", 123, "region", "us-east-1")
// 2. Direct baggage.Member: SetBaggage(ctx, member)
//
// Invalid arguments (non-string keys, unpaired values) are logged and skipped.
func SetBaggage(ctx context.Context, args ...any) context.Context {
	if len(args) == 0 {
		return ctx
	}

	b := baggage.FromContext(ctx)
	b = argsToBaggage(ctx, b, args)
	return baggage.ContextWithBaggage(ctx, b)
}

// argsToBaggage recursively builds the baggage from args,
// supporting both [string, any]... pairs and baggage.Member values.
// Invalid arguments are logged and skipped.
func argsToBaggage(ctx context.Context, b baggage.Baggage, args []any) baggage.Baggage {
	if len(args) == 0 {
		return b
	}

	switch x := args[0].(type) {
	case string:
		if len(args) == 1 {
			slog.WarnContext(ctx, "SetBaggage: unpaired string key, skipping", "key", x)
			return b
		}
		member, err := baggage.NewMemberRaw(x, fmt.Sprint(args[1]))
		if err != nil {
			slog.WarnContext(ctx, "SetBaggage: failed to create baggage member, skipping", "key", x, "err", err)
			return argsToBaggage(ctx, b, args[2:])
		}
		b, _ = b.SetMember(member)
		return argsToBaggage(ctx, b, args[2:])

	case baggage.Member:
		b, _ = b.SetMember(x)
		return argsToBaggage(ctx, b, args[1:])

	default:
		slog.WarnContext(ctx, "SetBaggage: unsupported argument type, skipping",
			"arg_type", fmt.Sprintf("%T", x), "arg_value", x)
		return argsToBaggage(ctx, b, args[1:])
	}
}
