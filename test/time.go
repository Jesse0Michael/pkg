package test

import (
	"testing"
	"time"
)

// ParseTime will parse a time value in a given format while handling any errors in the test
func ParseTime(t *testing.T, layout, value string) time.Time {
	t.Helper()
	v, err := time.Parse(layout, value)
	if err != nil {
		t.Errorf("failed to parse time layout: %s value: %s", layout, value)
	}

	return v
}

// ParseTimeInLocation will parse a time value in a given format and location while handling any errors in the test
func ParseTimeInLocation(t *testing.T, layout, value string, loc *time.Location) time.Time {
	t.Helper()
	v, err := time.ParseInLocation(layout, value, loc)
	if err != nil {
		t.Errorf("failed to parse time layout: %s value: %s, location: %s", layout, value, loc.String())
	}

	return v
}
