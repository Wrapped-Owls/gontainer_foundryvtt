// Package config holds the foundrymanager runtime configuration.
package config

// Config holds foundrymanager-specific configuration.
type Config struct {
	ProfilesFile  string
	DashboardAddr string
}

// Default returns the default config with container-friendly values.
func Default() Config {
	return Config{
		ProfilesFile:  "/etc/foundry/profiles.json",
		DashboardAddr: "0.0.0.0:30002",
	}
}
