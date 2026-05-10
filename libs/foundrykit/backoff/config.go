package backoff

import (
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrykit/confloader"
)

const (
	envCacheDir     = "CONTAINER_CACHE"
	envKubeHost     = "KUBERNETES_SERVICE_HOST"
	defaultCacheDir = "/data/container_cache"
)

type Config struct {
	CacheDir         string
	KubernetesBypass bool
}

func Default() Config {
	return Config{CacheDir: defaultCacheDir}
}

func LoadFromEnv(c *Config) error {
	return confloader.BindEnv(
		confloader.BindFieldPresent(&c.CacheDir, envCacheDir, nil),
		confloader.BindField(&c.KubernetesBypass, envKubeHost, func(v string) (bool, error) {
			return v != "", nil
		}),
	)
}

func NewFromConfig(cfg Config) *Tracker {
	return &Tracker{
		CacheDir:         cfg.CacheDir,
		KubernetesBypass: cfg.KubernetesBypass,
	}
}
