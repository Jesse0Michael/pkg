package config

import (
	"crypto/rsa"
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/golang-jwt/jwt/v5"
)

func TestFile_Decode(t *testing.T) {
	tmp, _ := os.CreateTemp("", "")
	fmt.Println(tmp.Name())
	defer os.Remove(tmp.Name())
	_, _ = tmp.Write([]byte("test-file"))

	tests := []struct {
		name    string
		value   string
		want    []byte
		wantErr bool
	}{
		{
			name:    "decode file",
			value:   tmp.Name(),
			want:    []byte("test-file"),
			wantErr: false,
		},
		{
			name:    "missing file",
			value:   "/does-not-exist",
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := File{}
			if err := f.Decode(tt.value); (err != nil) != tt.wantErr {
				t.Errorf("File.Decode() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual([]byte(f), tt.want) {
				t.Errorf("File.Decode() = %v, want %v", f, tt.want)
			}
		})
	}
}

func TestRSAPublicKey_Decode(t *testing.T) {
	b, _ := os.ReadFile("testdata/jwtRS256.key.pub")
	pubickKey, _ := jwt.ParseRSAPublicKeyFromPEM(b)

	tests := []struct {
		name    string
		value   string
		want    rsa.PublicKey
		wantErr bool
	}{
		{
			name:    "decode file",
			value:   "testdata/jwtRS256.key.pub",
			want:    *pubickKey,
			wantErr: false,
		},
		{
			name:    "missing file",
			value:   "/does-not-exist",
			want:    rsa.PublicKey{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := RSAPublicKey{}
			if err := k.Decode(tt.value); (err != nil) != tt.wantErr {
				t.Errorf("RSAPublicKey.Decode() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(rsa.PublicKey(k), tt.want) {
				t.Errorf("RSAPublicKey.Decode() = %v, want %v", k, tt.want)
			}
		})
	}
}
