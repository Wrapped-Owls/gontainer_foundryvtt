package config

// Config holds foundrymanager-specific configuration.
type Config struct {
	ProfilesFile string
	// DefaultProfile activates when no last-active one is recorded.
	DefaultProfile   string
	DashboardAddr    string
	LogAlertPatterns []string
}

// Default returns the default config with container-friendly values.
func Default() Config {
	return Config{
		ProfilesFile:  "/etc/foundry/profiles.json",
		DashboardAddr: "0.0.0.0:30002",
		LogAlertPatterns: []string{
			"lacks permission",
			"does not have permission",
			"permission denied",
		},
	}
}
