package backoff

import (
	"time"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrykit/confloader"
)

const (
	envCacheDir     = "CONTAINER_CACHE"
	envKubeHost     = "KUBERNETES_SERVICE_HOST"
	defaultCacheDir = "/data/container_cache"
)

// Config configures the backoff manager.
type Config struct {
	// CacheDir is where backoff_state.json is persisted.
	// Empty string disables persistence. Default: /data/container_cache.
	CacheDir string
	// KubernetesBypass, when true, makes OnFailure return immediately
	// (CrashLoopBackOff handles restart throttling).
	KubernetesBypass bool
}

// Default returns a Config with the standard cache directory.
func Default() Config {
	return Config{CacheDir: defaultCacheDir}
}

// LoadFromEnv reads CONTAINER_CACHE and KUBERNETES_SERVICE_HOST.
func LoadFromEnv(c *Config) error {
	return confloader.BindEnv(
		// BindFieldPresent so that an explicit empty CONTAINER_CACHE="" disables persistence.
		confloader.BindFieldPresent(&c.CacheDir, envCacheDir, nil),
		confloader.BindField(&c.KubernetesBypass, envKubeHost, func(v string) (bool, error) {
			return v != "", nil
		}),
	)
}

// NewFromConfig constructs a Manager from a Config.
func NewFromConfig(cfg Config) *Manager {
	return &Manager{
		CacheDir:         cfg.CacheDir,
		KubernetesBypass: cfg.KubernetesBypass,
		Now:              time.Now,
	}
}
