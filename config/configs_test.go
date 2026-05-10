package config

import (
	"os"
	"runtime/debug"
	"testing"
)

func TestBuildVersion(t *testing.T) {
	tests := []struct {
		name string
		info *debug.BuildInfo
		want string
	}{
		{
			name: "module version set",
			info: &debug.BuildInfo{
				Main: debug.Module{Version: "v1.2.3"},
			},
			want: "v1.2.3",
		},
		{
			name: "devel uses vcs revision",
			info: &debug.BuildInfo{
				Main: debug.Module{Version: "(devel)"},
				Settings: []debug.BuildSetting{
					{Key: "vcs.revision", Value: "abc1234def5678"},
				},
			},
			want: "abc1234",
		},
		{
			name: "dirty revision",
			info: &debug.BuildInfo{
				Main: debug.Module{Version: "(devel)"},
				Settings: []debug.BuildSetting{
					{Key: "vcs.revision", Value: "abc1234def5678"},
					{Key: "vcs.modified", Value: "true"},
				},
			},
			want: "abc1234-dirty",
		},
		{
			name: "short revision unchanged",
			info: &debug.BuildInfo{
				Main: debug.Module{Version: "(devel)"},
				Settings: []debug.BuildSetting{
					{Key: "vcs.revision", Value: "abc"},
				},
			},
			want: "abc",
		},
		{
			name: "no version no revision",
			info: &debug.BuildInfo{
				Main: debug.Module{Version: "(devel)"},
			},
			want: "",
		},
		{
			name: "empty version uses revision",
			info: &debug.BuildInfo{
				Main: debug.Module{Version: ""},
				Settings: []debug.BuildSetting{
					{Key: "vcs.revision", Value: "abc1234def5678"},
				},
			},
			want: "abc1234",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildVersion(tt.info)
			if got != tt.want {
				t.Errorf("buildVersion() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNewWithInitializer(t *testing.T) {
	orig := os.Args
	os.Args = []string{"test"}
	t.Cleanup(func() { os.Args = orig })

	cfg, err := New[testDefaulterConfig]()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if cfg.Value != "defaulted" {
		t.Errorf("New() Value = %q, want %q", cfg.Value, "defaulted")
	}
}

type testDefaulterConfig struct {
	Value string `envconfig:"TEST_DEFAULTER_VALUE"`
}

func (c *testDefaulterConfig) Init() {
	c.Value = "defaulted"
}

func TestNewInitializerOverriddenByEnv(t *testing.T) {
	t.Setenv("TEST_DEFAULTER_VALUE", "from-env")

	orig := os.Args
	os.Args = []string{"test"}
	t.Cleanup(func() { os.Args = orig })

	cfg, err := New[testDefaulterConfig]()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if cfg.Value != "from-env" {
		t.Errorf("New() Value = %q, want %q", cfg.Value, "from-env")
	}
}
