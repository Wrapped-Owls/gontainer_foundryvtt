package config

type Config struct {
	ProfilesFile     string
	DefaultProfile   string
	DashboardAddr    string
	LogAlertPatterns []string
}

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
