package auth

import (
	"context"
	"testing"
)

func TestAuthorization(t *testing.T) {
	tests := []struct {
		name   string
		ctx    context.Context
		want   string
		wantOK bool
	}{
		{
			name:   "authorization ID found",
			ctx:    context.WithValue(context.TODO(), AuthorizationContextKey, "test-auth"),
			want:   "test-auth",
			wantOK: true,
		},
		{
			name:   "authorization ID invalid",
			ctx:    context.WithValue(context.TODO(), AuthorizationContextKey, 12345),
			want:   "",
			wantOK: false,
		},
		{
			name:   "authorization ID not found",
			ctx:    context.TODO(),
			want:   "",
			wantOK: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := Authorization(tt.ctx)
			if got != tt.want {
				t.Errorf("Authorization() got = %v, want %v", got, tt.want)
			}
			if ok != tt.wantOK {
				t.Errorf("Authorization() ok = %v, want %v", ok, tt.wantOK)
			}
		})
	}
}

func TestSubject(t *testing.T) {
	tests := []struct {
		name   string
		ctx    context.Context
		want   string
		wantOK bool
	}{
		{
			name:   "subject found",
			ctx:    context.WithValue(context.TODO(), SubjectContextKey, "test-account"),
			want:   "test-account",
			wantOK: true,
		},
		{
			name:   "subject invalid",
			ctx:    context.WithValue(context.TODO(), SubjectContextKey, 12345),
			want:   "",
			wantOK: false,
		},
		{
			name:   "subject not found",
			ctx:    context.TODO(),
			want:   "",
			wantOK: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := Subject(tt.ctx)
			if got != tt.want {
				t.Errorf("Subject() got = %v, want %v", got, tt.want)
			}
			if ok != tt.wantOK {
				t.Errorf("Subject() ok = %v, want %v", ok, tt.wantOK)
			}
		})
	}
}

func TestAdmin(t *testing.T) {
	tests := []struct {
		name   string
		ctx    context.Context
		want   bool
		wantOK bool
	}{
		{
			name:   "admin found - true",
			ctx:    context.WithValue(context.TODO(), AdminContextKey, true),
			want:   true,
			wantOK: true,
		},
		{
			name:   "admin found - false",
			ctx:    context.WithValue(context.TODO(), AdminContextKey, false),
			want:   false,
			wantOK: true,
		},
		{
			name:   "admin invalid",
			ctx:    context.WithValue(context.TODO(), AdminContextKey, "truth"),
			want:   false,
			wantOK: false,
		},
		{
			name:   "admin not found",
			ctx:    context.TODO(),
			want:   false,
			wantOK: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := Admin(tt.ctx)
			if got != tt.want {
				t.Errorf("Admin() got = %v, want %v", got, tt.want)
			}
			if ok != tt.wantOK {
				t.Errorf("Admin() ok = %v, want %v", ok, tt.wantOK)
			}
		})
	}
}

func TestReadOnly(t *testing.T) {
	tests := []struct {
		name   string
		ctx    context.Context
		want   bool
		wantOK bool
	}{
		{
			name:   "readOnly found - true",
			ctx:    context.WithValue(context.TODO(), ReadOnlyContextKey, true),
			want:   true,
			wantOK: true,
		},
		{
			name:   "readOnly found - false",
			ctx:    context.WithValue(context.TODO(), ReadOnlyContextKey, false),
			want:   false,
			wantOK: true,
		},
		{
			name:   "readOnly invalid",
			ctx:    context.WithValue(context.TODO(), ReadOnlyContextKey, "truth"),
			want:   false,
			wantOK: false,
		},
		{
			name:   "readOnly not found",
			ctx:    context.TODO(),
			want:   false,
			wantOK: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := ReadOnly(tt.ctx)
			if got != tt.want {
				t.Errorf("ReadOnly() got = %v, want %v", got, tt.want)
			}
			if ok != tt.wantOK {
				t.Errorf("ReadOnly() ok = %v, want %v", ok, tt.wantOK)
			}
		})
	}
}

func TestCheck(t *testing.T) {
	tests := []struct {
		name    string
		subject string
		ctx     context.Context
		want    bool
	}{
		{
			name:    "empty context",
			subject: "test-account",
			ctx:     context.TODO(),
			want:    false,
		},
		{
			name:    "non admin",
			subject: "test-account",
			ctx:     context.WithValue(context.TODO(), AdminContextKey, false),
			want:    false,
		},
		{
			name:    "admin",
			subject: "test-account",
			ctx:     context.WithValue(context.TODO(), AdminContextKey, true),
			want:    true,
		},
		{
			name:    "empty subject",
			subject: "",
			ctx:     context.WithValue(context.TODO(), SubjectContextKey, ""),
			want:    false,
		},
		{
			name:    "matching subject",
			subject: "test-account",
			ctx:     context.WithValue(context.TODO(), SubjectContextKey, "test-account"),
			want:    true,
		},
		{
			name:    "not matching subject",
			subject: "test-account",
			ctx:     context.WithValue(context.TODO(), SubjectContextKey, "non-account"),
			want:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Check(tt.ctx, tt.subject)
			if got != tt.want {
				t.Errorf("Check() got = %v, want %v", got, tt.want)
			}
		})
	}
}
