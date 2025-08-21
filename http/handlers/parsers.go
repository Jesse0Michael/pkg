package handlers

import (
	"fmt"
	"io"

	"github.com/go-playground/validator/v10"
	jsoniter "github.com/json-iterator/go"
)

// Decode decodes a reader, unmarshalls the bytes into the out interface as json
func Decode(r io.Reader, out any, processors ...func(any) error) error {
	b, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}

	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	if err = json.Unmarshal(b, &out); err != nil {
		return fmt.Errorf("failed to unmarshal input: %w", err)
	}

	for _, p := range processors {
		if err := p(out); err != nil {
			return err
		}
	}

	return nil
}

// DecodeValidate decodes a reader, unmarshalls the bytes into the out interface as json, and validates it
func DecodeValidate(r io.Reader, v *validator.Validate, out interface{}) error {
	if v == nil {
		v = validator.New()
	}

	return Decode(r, out, Validate(v))
}

// Validate is a processor function that will validate an interface with a given validator
func Validate(v *validator.Validate) func(interface{}) error {
	return func(o interface{}) error { return v.Var(o, "dive") }
}
