package errors

import (
	"fmt"
)

type Error struct {
	Code    int    `json:"-"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// NewError creates a new error with HTTP status code, message, and (optional) details
func NewError(code int, message, details string) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Details: details,
	}
}

// Error converts the Error to a string
func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	result := e.Message
	if e.Code > 0 {
		result = fmt.Sprintf("(%d) %s", e.Code, e.Message)
	}
	if e.Details != "" {
		if result == "" {
			return e.Details
		}
		result += fmt.Sprintf(": %s", e.Details)
	}
	return result
}
