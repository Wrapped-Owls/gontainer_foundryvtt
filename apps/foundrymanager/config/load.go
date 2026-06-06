package config

import "github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrykit/confloader"

const DefaultFileName = "foundrymanager.json"

func Load() (Config, error) {
	return confloader.Load(DefaultFileName, Default(), LoadFromEnv)
}

func LoadFromEnv(c *Config) error {
	return confloader.BindEnv(
		confloader.BindField(&c.ProfilesFile, envProfilesFile, nil),
		confloader.BindField(&c.DashboardAddr, envDashboardAddr, nil),
	)
}
