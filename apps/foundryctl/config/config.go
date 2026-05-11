package config

import (
	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundryctl/internal/secfuse"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrykit/backoff"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrykit/colorlog"
)

type Config struct {
	Paths   PathsConfig
	Install InstallConfig
	Admin   AdminConfig
	Runtime RuntimeConfig
	Log     colorlog.Config
	Backoff backoff.Config
	Secrets secfuse.Config
}

type PathsConfig struct {
	DataPath     string
	InstallRoot  string
	SourcesDir   string
	ManifestPath string
	MainScript   string
	HealthAddr   string
}

type InstallConfig struct {
	Version    string
	ReleaseURL string
	Session    string
	Username   string
	Password   string
}

type AdminConfig struct {
	Key          string
	PasswordSalt string
}

type RuntimeConfig struct {
	Port int
}

func Default() Config {
	return Config{
		Paths: PathsConfig{
			DataPath:     "/data",
			InstallRoot:  "/foundry",
			SourcesDir:   "/foundry/sources",
			ManifestPath: "/etc/foundry/patches/manifest.yaml",
			MainScript:   "resources/app/main.mjs",
			HealthAddr:   "127.0.0.1:30001",
		},
		Runtime: RuntimeConfig{Port: 30000},
		Log:     colorlog.Default(),
		Backoff: backoff.Default(),
		Secrets: secfuse.Default(),
	}
}
