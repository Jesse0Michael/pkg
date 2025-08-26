package server

import (
	"bytes"
	"errors"
	"io"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/go-playground/validator/v10"
)

type errReader int

func (errReader) Read(_ []byte) (n int, err error) {
	return 0, errors.New("test-error")
}

func TestDecode(t *testing.T) {
	type Thing struct {
		A string `validate:"required"`
	}
	tests := []struct {
		name    string
		r       io.Reader
		out     interface{}
		wantErr error
		want    Thing
	}{
		{
			name:    "failed to read",
			r:       errReader(0),
			out:     &Thing{},
			wantErr: errors.New("failed to read input: test-error"),
			want:    Thing{},
		},
		{
			name:    "failed to unmarshal",
			r:       bytes.NewBuffer([]byte("}{")),
			out:     &Thing{},
			wantErr: errors.New("failed to unmarshal input: readObjectStart: expect { or n, but found }, error found in #1 byte of ...|}{|..., bigger context ...|}{|..."), // nolint:revive
			want:    Thing{},
		},
		{
			name:    "successful decode and validate",
			r:       bytes.NewBuffer([]byte(`{"A":"a"}`)),
			out:     &Thing{},
			wantErr: nil,
			want:    Thing{A: "a"},
		},
		{
			name:    "successful decode and validate with nil validator",
			r:       bytes.NewBuffer([]byte(`{"A":"a"}`)),
			out:     &Thing{},
			wantErr: nil,
			want:    Thing{A: "a"},
		},
	}
	for i := range tests {
		tt := tests[i]
		t.Run(tt.name, func(t *testing.T) {
			err := Decode(tt.r, tt.out)
			if tt.wantErr == nil && err != nil {
				t.Errorf("Decode() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr != nil && (err == nil || tt.wantErr.Error() != err.Error()) {
				t.Errorf("Decode() error = %v, wantErr %v", err, tt.wantErr)
			}
			if *tt.out.(*Thing) != tt.want {
				t.Errorf("Decode() out = %v, want %v", tt.out, tt.want)
			}
		})
	}
}

func TestDecodeValidate(t *testing.T) {
	validate := validator.New()
	type Thing struct {
		A string `validate:"required"`
	}
	tests := []struct {
		name    string
		r       io.Reader
		v       *validator.Validate
		wantErr error
		want    []Thing
	}{
		{
			name:    "failed to validate",
			r:       bytes.NewBuffer([]byte(`[{}]`)),
			v:       validate,
			wantErr: errors.New("Key: '[0].A' Error:Field validation for 'A' failed on the 'required' tag"),
			want:    []Thing{{}},
		},
		{
			name:    "successful decode and validate",
			r:       bytes.NewBuffer([]byte(`[{"A":"a"}]`)),
			v:       validate,
			wantErr: nil,
			want:    []Thing{{A: "a"}},
		},
	}
	for i := range tests {
		tt := tests[i]
		t.Run(tt.name, func(t *testing.T) {
			var out []Thing
			err := DecodeValidate(tt.r, tt.v, &out)
			if tt.wantErr == nil && err != nil {
				t.Errorf("DecodeValidate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr != nil && (err == nil || tt.wantErr.Error() != err.Error()) {
				t.Errorf("DecodeValidate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(out, tt.want) {
				t.Errorf("DecodeValidate() = %v, want %v", out, tt.want)
			}
		})
	}
}

func TestEncode(t *testing.T) {
	type Thing struct {
		A string `json:"A"`
	}

	tests := []struct {
		name       string
		status     int
		value      any
		wantStatus int
		wantBody   string
		wantErr    error
	}{
		{
			name:       "nil value",
			status:     204,
			value:      nil,
			wantStatus: 204,
			wantBody:   "",
			wantErr:    nil,
		},
		{
			name:       "valid struct",
			status:     200,
			value:      Thing{A: "a"},
			wantStatus: 200,
			wantBody:   "{\"A\":\"a\"}\n",
			wantErr:    nil,
		},
		{
			name:       "invalid value (non-marshalable)",
			status:     200,
			value:      func() {}, // functions cannot be marshaled to JSON
			wantStatus: 200,
			wantBody:   "",
			wantErr:    errors.New("failed to encode json: json: unsupported type: func()"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			err := Encode(w, tt.status, tt.value)

			resp := w.Result()
			defer resp.Body.Close()

			if resp.StatusCode != tt.wantStatus {
				t.Errorf("Encode() status = %v, want %v", resp.StatusCode, tt.wantStatus)
			}

			body, _ := io.ReadAll(resp.Body)
			if tt.wantBody != "" && string(body) != tt.wantBody {
				t.Errorf("Encode() body = %q, want %q", string(body), tt.wantBody)
			}
			if tt.wantBody == "" && len(body) != 0 {
				t.Errorf("Encode() body should be empty, got %q", string(body))
			}

			if tt.wantErr != nil {
				if err == nil || err.Error() != tt.wantErr.Error() {
					t.Errorf("Encode() error = %v, want %v", err, tt.wantErr)
				}
			} else if err != nil {
				t.Errorf("Encode() error = %v, want nil", err)
			}
		})
	}
}
