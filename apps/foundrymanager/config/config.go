package config

type Config struct {
	ProfilesFile  string
	DashboardAddr string
}

func Default() Config {
	return Config{
		ProfilesFile:  "/etc/foundry/profiles.json",
		DashboardAddr: "0.0.0.0:30002",
	}
}
