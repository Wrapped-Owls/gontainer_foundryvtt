package config

import "github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrykit/confloader"

const DefaultFileName = "foundrymanager.json"

// Load reads config from the default JSON file overlaid with env vars.
func Load() (Config, error) {
	return confloader.Load(DefaultFileName, Default(), LoadFromEnv)
}

// LoadFromEnv overlays Config fields from environment variables.
func LoadFromEnv(c *Config) error {
	return confloader.BindEnv(
		confloader.BindField(&c.ProfilesFile, envProfilesFile, nil),
		confloader.BindField(&c.DashboardAddr, envDashboardAddr, nil),
	)
}
