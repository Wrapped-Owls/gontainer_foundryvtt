package config

import (
	"strings"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrykit/confloader"
)

const DefaultFileName = "foundrymanager.json"

func Load() (Config, error) {
	return confloader.Load(DefaultFileName, Default(), LoadFromEnv)
}

func LoadFromEnv(c *Config) error {
	return confloader.BindEnv(
		confloader.BindField(&c.ProfilesFile, envProfilesFile, nil),
		confloader.BindField(&c.DashboardAddr, envDashboardAddr, nil),
		confloader.BindField(&c.LogAlertPatterns, envLogPatterns, parsePatterns),
	)
}

func parsePatterns(v string) ([]string, error) {
	var out []string
	for p := range strings.SplitSeq(v, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out, nil
}
