package test

import (
	"fmt"
	"testing"
	"time"
)

func ExampleParse() {
	t := &testing.T{}

	fmt.Println(ParseTime(t, time.RFC3339, "1990-06-04T01:02:03Z"))
	// Output:
	// 1990-06-04 01:02:03 +0000 UTC
}

func ExampleParseInLocation() {
	t := &testing.T{}

	fmt.Println(ParseTimeInLocation(t, time.RFC3339, "1990-06-04T01:02:03Z", time.UTC))
	// Output:
	// 1990-06-04 01:02:03 +0000 UTC
}

func TestParseTime_Failure(t *testing.T) {
	tt := &testing.T{}
	ParseTime(tt, time.RFC3339, "June 4 1990")
	if !tt.Failed() {
		t.Error("expected ParseTime() to fail to parse time")
	}
}

func TestParseTimeInLocation_Failure(t *testing.T) {
	tt := &testing.T{}
	ParseTimeInLocation(tt, time.RFC3339, "June 4 1990", time.UTC)
	if !tt.Failed() {
		t.Error("expected ParseTimeInLocation() to fail to parse time")
	}
}
