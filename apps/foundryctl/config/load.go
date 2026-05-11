package config

import (
	"strconv"

	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundryctl/internal/secfuse"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrykit/backoff"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrykit/colorlog"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrykit/confloader"
	runtimecfg "github.com/wrapped-owls/gontainer_foundryvtt/libs/foundryruntime/config"
)

const DefaultFileName = "foundryctl.json"

func Load() (Config, error) {
	return confloader.Load(DefaultFileName, Default(), LoadFromEnv)
}

func LoadFromEnv(c *Config) error {
	return confloader.BindEnv(
		loadPathsFromEnv(&c.Paths),
		loadInstallFromEnv(&c.Install),
		loadAdminFromEnv(&c.Admin),
		loadRuntimeFromEnv(&c.Runtime),
		func() error { return colorlog.LoadFromEnv(&c.Log) },
		func() error { return backoff.LoadFromEnv(&c.Backoff) },
		func() error { return secfuse.LoadFromEnv(&c.Secrets) },
	)
}

func loadPathsFromEnv(c *PathsConfig) confloader.Binder {
	return func() error {
		return confloader.BindEnv(
			confloader.BindField(&c.DataPath, envDataPath, nil),
			confloader.BindField(&c.InstallRoot, envInstallRoot, nil),
			confloader.BindField(&c.SourcesDir, envSourcesDir, nil),
			confloader.BindField(&c.ManifestPath, envManifestPath, nil),
			confloader.BindField(&c.MainScript, envMainScript, nil),
			confloader.BindField(&c.HealthAddr, envHealthAddr, nil),
		)
	}
}

func loadInstallFromEnv(c *InstallConfig) confloader.Binder {
	return func() error {
		return confloader.BindEnv(
			confloader.BindField(&c.Version, envVersion, nil),
			confloader.BindField(&c.ReleaseURL, envReleaseURL, nil),
			confloader.BindField(&c.Session, envSession, nil),
			confloader.BindField(&c.Username, envUsername, nil),
			confloader.BindField(&c.Password, envPassword, nil),
		)
	}
}

func loadAdminFromEnv(c *AdminConfig) confloader.Binder {
	return func() error {
		return confloader.BindEnv(
			confloader.BindField(&c.Key, envAdminKey, nil),
			confloader.BindField(&c.PasswordSalt, envPasswordSalt, nil),
		)
	}
}

func loadRuntimeFromEnv(c *RuntimeConfig) confloader.Binder {
	return func() error {
		return confloader.BindEnv(
			confloader.BindField(&c.Port, envPort, func(v string) (int, error) {
				p, err := strconv.Atoi(v)
				if err != nil {
					return c.Port, nil
				}
				if p < runtimecfg.MinPort || p > runtimecfg.MaxPort {
					return c.Port, nil
				}
				return p, nil
			}),
		)
	}
}
