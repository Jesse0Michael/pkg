package errors

import (
	"testing"
)

func TestError_Error(t *testing.T) {
	tests := []struct {
		name string
		err  *Error
		want string
	}{
		{
			name: "nil error",
			err:  nil,
			want: "",
		},
		{
			name: "new error",
			err:  NewError(404, "resource not found", "test-resource"),
			want: "(404) resource not found: test-resource",
		},
		{
			name: "bare error",
			err:  &Error{Message: "resource not found"},
			want: "resource not found",
		},
		{
			name: "new error no details",
			err:  NewError(403, "forbidden", ""),
			want: "(403) forbidden",
		},
		{
			name: "new error just details",
			err:  NewError(0, "", "just details"),
			want: "just details",
		},
	}
	for i := range tests {
		tt := tests[i]
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.want {
				t.Errorf("Error.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}
