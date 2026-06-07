// Package config holds runtime configuration for the foundrydiscord bot.
package config

// DiscordConfig holds Discord API credentials and bot settings.
type DiscordConfig struct {
	Token         string
	ApplicationID string
	GuildID       string // empty = register commands globally
	GMRoleID      string // empty = unrestricted access
}

// FoundryConfig holds connection settings for the foundrymanager dashboard.
type FoundryConfig struct {
	DashboardURL string
}

// Config is the top-level configuration container.
type Config struct {
	Discord DiscordConfig
	Foundry FoundryConfig
}

// Default returns configuration with container-friendly defaults.
func Default() Config {
	return Config{
		Foundry: FoundryConfig{
			DashboardURL: "http://foundryvtt:30002",
		},
	}
}
