package test

import (
	"bytes"
	"encoding/json"
	"html/template"
	"os"
	"testing"
)

// LoadFile will read test data file handling errors through the testing interface
func LoadFile(t *testing.T, file string) []byte {
	t.Helper()
	b, err := os.ReadFile(file)
	if err != nil {
		t.Errorf("failed to load file: %s", file)
	}

	return b
}

// LoadJSONFile reads a JSON test data file and unmarshals it into a new value of type T.
func LoadJSONFile[T any](t *testing.T, file string) T {
	t.Helper()
	var zero T

	b := LoadFile(t, file)

	var result T
	if err := json.Unmarshal(b, &result); err != nil {
		t.Errorf("failed to unmarshal JSON file: %s", string(b))
		return zero
	}

	return result
}

// LoadTemplate loads and executes a template file with the provided data, returning the result as bytes.
func LoadTemplate(t *testing.T, filename string, data any) []byte {
	t.Helper()
	tmpl, err := template.ParseFiles(filename)
	if err != nil {
		t.Errorf("failed to parse template file: %s", filename)
		return nil
	}
	var out bytes.Buffer
	err = tmpl.Execute(&out, data)
	if err != nil {
		t.Errorf("failed to execute template file: %s", filename)
	}
	return out.Bytes()
}

// LoadJSONTemplate loads and executes a template file with the provided data,
// then unmarshals the result into a new value of type T.
func LoadJSONTemplate[T any](t *testing.T, filename string, data any) T {
	t.Helper()
	var zero T

	b := LoadTemplate(t, filename, data)

	var result T
	if err := json.Unmarshal(b, &result); err != nil {
		t.Errorf("failed to unmarshal JSON template: %s", string(b))
		return zero
	}

	return result
}
