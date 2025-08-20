package errors

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"go.opentelemetry.io/otel/trace"
)

// ErrorResponse is a collection of errors intended to standardize error responses
type ErrorResponse struct {
	RequestID string  `json:"requestID,omitempty"`
	Errors    []Error `json:"errors"`
	Code      int     `json:"-"`
}

// NewErrorResponse creates a new error response with the given request ID and errors
// it returns the highest HTTP status code found in the errors, defaulting to 500 Internal Server Error
// Errors from this package are included in the response
// If there are no errors that are from this package a default error  is used
func NewErrorResponse(requestID string, errs ...error) *ErrorResponse {
	var group []Error
	var code int
	for _, err := range errs {
		var e *Error
		if ok := errors.As(err, &e); ok {
			group = append(group, *e)
			if e.Code > 0 {
				code = e.Code
			}
		}
	}
	if len(group) == 0 && len(errs) > 0 {
		group = append(group, Error{Message: "Internal Server Error"})
	}
	if code == 0 {
		code = http.StatusInternalServerError
	}
	return &ErrorResponse{
		RequestID: requestID,
		Errors:    group,
		Code:      code,
	}
}

// Error joins the messages from the inner errors in the ErrorResponse
func (e *ErrorResponse) Error() string {
	errs := make([]error, len(e.Errors))
	for i, err := range e.Errors {
		errs[i] = errors.New(err.Message)
	}
	return errors.Join(errs...).Error()
}

// Write sets the HTTP response status code and body on the response
// A request ID will be included if available
// If the error response code is 500 or higher, the error will be recorded in the span
func WriteError(ctx context.Context, w http.ResponseWriter, errs ...error) {
	if len(errs) == 0 {
		slog.ErrorContext(ctx, "WriteError called with no errors")
	}
	var requestID string
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().HasTraceID() {
		requestID = span.SpanContext().TraceID().String()
	}

	errResponse := NewErrorResponse(requestID, errs...)

	if errResponse.Code >= http.StatusInternalServerError {
		span.RecordError(errResponse)
	}
	w.WriteHeader(errResponse.Code)
	b, _ := json.Marshal(errResponse)
	_, _ = w.Write(b)
}
