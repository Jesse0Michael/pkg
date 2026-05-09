package boot

import "testing"

func TestWithConfigPrefix(t *testing.T) {
	var o options
	WithConfigPrefix("test-prefix")(&o)

	if len(o.configOpts) != 1 {
		t.Fatalf("expected 1 config option, got %d", len(o.configOpts))
	}
}

func TestWithConfigFile(t *testing.T) {
	var o options
	WithConfigFile("test-path")(&o)

	if len(o.configOpts) != 1 {
		t.Fatalf("expected 1 config option, got %d", len(o.configOpts))
	}
}
