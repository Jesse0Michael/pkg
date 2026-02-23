package logger

import (
	"context"
	"errors"
	"net"

	"github.com/lib/pq"
)

// ContextWarn returns true if the error is a context error.
func ContextWarn(err error) bool {
	return errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
}

// NetWarn returns true if the error is a network timeout error.
func NetWarn(err error) bool {
	var netErr net.Error
	if errors.As(err, &netErr) {
		return netErr.Timeout()
	}
	return false
}

// PQWarn returns true if the error is a Postgres cancellation error.
func PQWarn(err error) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		// 57014: "query_canceled"
		return pqErr.Code == "57014"
	}
	return false
}
