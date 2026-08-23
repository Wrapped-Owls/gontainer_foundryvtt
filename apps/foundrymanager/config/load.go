package config

import (
	"strings"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrykit/confloader"
)

const DefaultFileName = "foundrymanager.json"

// Load reads config from the default JSON file overlaid with env vars.
func Load() (Config, error) {
	return confloader.Load(DefaultFileName, Default(), LoadFromEnv)
}

// LoadFromEnv overlays Config fields from environment variables.
func LoadFromEnv(c *Config) error {
	return confloader.BindEnv(
		confloader.BindField(&c.ProfilesFile, envProfilesFile, nil),
		confloader.BindField(&c.DefaultProfile, envDefaultProfile, nil),
		confloader.BindField(&c.DashboardAddr, envDashboardAddr, nil),
		confloader.BindField(&c.LogAlertPatterns, envLogPatterns, parsePatterns),
	)
}

// parsePatterns splits a comma-separated list into trimmed, non-empty patterns.
func parsePatterns(v string) ([]string, error) {
	var out []string
	for p := range strings.SplitSeq(v, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out, nil
}
