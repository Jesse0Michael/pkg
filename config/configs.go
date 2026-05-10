package config

import (
	"path"
	"runtime/debug"
)

type AppConfig struct {
	Environment string `envconfig:"ENVIRONMENT"`
	Name        string `envconfig:"APP_NAME"`
	Version     string `envconfig:"VERSION"`
}

// Init populates Name and Version from Go build info
// embedded by the toolchain (module path and VCS revision).
func (c *AppConfig) Init() {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return
	}
	c.Name = path.Base(info.Main.Path)
	c.Version = buildVersion(info)
}

func buildVersion(info *debug.BuildInfo) string {
	if info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	var revision string
	var modified bool
	for _, s := range info.Settings {
		switch s.Key {
		case "vcs.revision":
			revision = s.Value
		case "vcs.modified":
			modified = s.Value == "true"
		}
	}
	if revision == "" {
		return ""
	}
	if len(revision) > 7 {
		revision = revision[:7]
	}
	if modified {
		revision += "-dirty"
	}
	return revision
}
