package config

import (
	"crypto/rsa"
	"os"

	"github.com/golang-jwt/jwt/v4"
)

// File is a custom envconfig decoder that reads a file from disk
// using the assigned environment variable as the path to the file
type File []byte

// Decode a file using the environment variable value as the path to the file
func (f *File) Decode(value string) error {
	b, err := os.ReadFile(value)
	*f = b
	return err
}

// RSAPublicKey is a RSA public key PEM file read from the os
type RSAPublicKey rsa.PublicKey

// Decode a RSA public key using the environment variable value as the path to the file
func (k *RSAPublicKey) Decode(value string) error {
	b, err := os.ReadFile(value)
	if err != nil {
		return err
	}
	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(b)
	*k = RSAPublicKey(*publicKey)
	return err
}
