package config

type DiscordConfig struct {
	Token         string
	ApplicationID string
	GuildID       string
	GMRoleID      string
}

type FoundryConfig struct {
	DashboardURL string
}

type Config struct {
	Discord DiscordConfig
	Foundry FoundryConfig
}

func Default() Config {
	return Config{
		Foundry: FoundryConfig{
			DashboardURL: "http://foundryvtt:30002",
		},
	}
}
