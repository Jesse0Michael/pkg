package errors

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"go.opentelemetry.io/otel/trace"
)

func TestNewErrorResponse(t *testing.T) {
	tests := []struct {
		name      string
		requestID string
		errs      []error
		expected  *ErrorResponse
	}{
		{
			name:      "no errors",
			requestID: "test-request-id",
			errs:      []error{},
			expected: &ErrorResponse{
				RequestID: "test-request-id",
				Code:      http.StatusInternalServerError,
			},
		},
		{
			name:      "nil error",
			requestID: "test-request-id",
			errs:      []error{nil},
			expected: &ErrorResponse{
				RequestID: "test-request-id",
				Errors:    []Error{{Message: "Internal Server Error"}},
				Code:      http.StatusInternalServerError,
			},
		},
		{
			name:      "standard error",
			requestID: "test-request-id",
			errs:      []error{errors.New("test-error")},
			expected: &ErrorResponse{
				RequestID: "test-request-id",
				Errors:    []Error{{Message: "Internal Server Error"}},
				Code:      http.StatusInternalServerError,
			},
		},
		{
			name:      "Error error",
			requestID: "test-request-id",
			errs:      []error{&Error{Message: "test-error", Code: http.StatusBadRequest}},
			expected: &ErrorResponse{
				RequestID: "test-request-id",
				Errors:    []Error{{Message: "test-error", Code: http.StatusBadRequest}},
				Code:      http.StatusBadRequest,
			},
		},
		{
			name:      "multiple Error errors with increasing codes",
			requestID: "test-request-id",
			errs: []error{
				&Error{Message: "test-error-1", Code: http.StatusBadRequest},
				&Error{Message: "test-error-2", Code: http.StatusUnauthorized},
				&Error{Message: "test-error-3", Code: http.StatusForbidden},
			},
			expected: &ErrorResponse{
				RequestID: "test-request-id",
				Errors: []Error{
					{Message: "test-error-1", Code: http.StatusBadRequest},
					{Message: "test-error-2", Code: http.StatusUnauthorized},
					{Message: "test-error-3", Code: http.StatusForbidden},
				},
				Code: http.StatusForbidden,
			},
		},
		{
			name:      "mixed Error errors and standard errors",
			requestID: "test-request-id",
			errs: []error{
				errors.New("test-error"),
				&Error{Message: "bad-request-error", Code: http.StatusBadRequest},
			},
			expected: &ErrorResponse{
				RequestID: "test-request-id",
				Errors: []Error{
					{Message: "bad-request-error", Code: http.StatusBadRequest},
				},
				Code: http.StatusBadRequest,
			},
		},
		{
			name:      "Error errors with zero code",
			requestID: "test-request-id",
			errs: []error{
				&Error{Message: "test-error-1"},
				&Error{Message: "test-error-2"},
			},
			expected: &ErrorResponse{
				RequestID: "test-request-id",
				Errors: []Error{
					{Message: "test-error-1"},
					{Message: "test-error-2"},
				},
				Code: http.StatusInternalServerError,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewErrorResponse(tt.requestID, tt.errs...)

			if result.RequestID != tt.expected.RequestID {
				t.Errorf("RequestID = %v, want %v", result.RequestID, tt.expected.RequestID)
			}

			if result.Code != tt.expected.Code {
				t.Errorf("Code = %v, want %v", result.Code, tt.expected.Code)
			}

			if !reflect.DeepEqual(result.Errors, tt.expected.Errors) {
				t.Errorf("Errors = %v, want %v", result.Errors, tt.expected.Errors)
			}
		})
	}
}

func TestErrorResponse_Error(t *testing.T) {
	tests := []struct {
		name     string
		response *ErrorResponse
		expected string
	}{
		{
			name: "nil errors",
			response: &ErrorResponse{
				RequestID: "test-request-id",
				Errors:    nil,
				Code:      500,
			},
			expected: "",
		},
		{
			name: "single error",
			response: &ErrorResponse{
				RequestID: "test-request-id",
				Errors:    []Error{{Message: "test-error"}},
				Code:      400,
			},
			expected: "test-error",
		},
		{
			name: "errors with details",
			response: &ErrorResponse{
				RequestID: "test-request-id",
				Errors: []Error{
					{Message: "test-error-1", Details: "test-details-1", Code: http.StatusForbidden},
					{Message: "test-error-2", Details: "test-details-2", Code: http.StatusInternalServerError},
				},
				Code: 500,
			},
			expected: "test-error-1\ntest-error-2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.response.Error()
			if result != tt.expected {
				t.Errorf("Error() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestWriteError(t *testing.T) {
	ctx := trace.ContextWithSpanContext(t.Context(), trace.NewSpanContext(trace.SpanContextConfig{TraceID: trace.TraceID([16]byte{128, 241, 152, 238, 86, 52, 59, 168, 100, 254, 139, 42, 87, 211, 239, 247})}))
	tests := []struct {
		name     string
		ctx      context.Context
		errs     []error
		wantCode int
		wantBody string
	}{
		{
			name:     "no errors",
			errs:     []error{},
			wantCode: http.StatusInternalServerError,
			wantBody: `{"requestID":"80f198ee56343ba864fe8b2a57d3eff7","errors":null}`,
		},
		{
			name:     "standard error",
			ctx:      t.Context(),
			errs:     []error{errors.New("test-error")},
			wantCode: http.StatusInternalServerError,
			wantBody: `{"requestID":"80f198ee56343ba864fe8b2a57d3eff7","errors":[{"message":"Internal Server Error"}]}`,
		},
		{
			name:     "error with status code",
			ctx:      t.Context(),
			errs:     []error{&Error{Message: "bad request", Details: "invalid request", Code: http.StatusBadRequest}},
			wantCode: http.StatusBadRequest,
			wantBody: `{"requestID":"80f198ee56343ba864fe8b2a57d3eff7","errors":[{"message":"bad request","details":"invalid request"}]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()

			WriteError(ctx, rec, tt.errs...)

			resp := rec.Result()
			if resp.StatusCode != tt.wantCode {
				t.Errorf("WriteError() status code = %v, want %v", resp.StatusCode, tt.wantCode)
			}

			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()

			if string(body) != tt.wantBody {
				t.Errorf("WriteError() body = %v, want %v", string(body), tt.wantBody)
			}
		})
	}
}
