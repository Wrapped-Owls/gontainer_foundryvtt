package config

import "github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrykit/confloader"

const DefaultFileName = "taverncord.json"

// Load reads config from the default JSON file overlaid with env vars.
func Load() (Config, error) {
	return confloader.Load(DefaultFileName, Default(), LoadFromEnv)
}

// LoadFromEnv overlays Config fields from environment variables.
func LoadFromEnv(c *Config) error {
	return confloader.BindEnv(
		confloader.BindField(&c.Discord.Token, envDiscordToken, nil),
		confloader.BindField(&c.Discord.ApplicationID, envDiscordApplicationID, nil),
		confloader.BindField(&c.Discord.GuildID, envDiscordGuildID, nil),
		confloader.BindField(&c.Discord.GMRoleID, envDiscordGMRoleID, nil),
		confloader.BindField(&c.Foundry.DashboardURL, envFoundryDashboardURL, nil),
	)
}
