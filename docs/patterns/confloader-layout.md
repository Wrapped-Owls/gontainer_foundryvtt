# Config package layout — `confloader` pattern

How to organise a config package using `libs/foundrykit/confloader`.

## Directory layout

Split by concern, not by type:

```
apps/<name>/
├── config.go      # Config + all sub-struct types + Default()
├── env_vars.go    # unexported const env* = "FOUNDRY_..."
└── load.go        # Load(), LoadFromEnv(), per-domain binder functions
```

## Example: `config.go`

```go
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

func Default() Config {
    return Config{
        Paths: PathsConfig{
            DataPath:     "/data",
            InstallRoot:  "/foundry",
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
```

## Example: `env_vars.go`

```go
package config

const (
    envDataPath     = "FOUNDRY_DATA_PATH"
    envInstallRoot  = "FOUNDRY_INSTALL_ROOT"
    envManifestPath = "FOUNDRY_PATCH_MANIFEST"
    envHealthAddr   = "FOUNDRY_HEALTH_ADDR"
    envPort         = "FOUNDRY_PORT"
    envVersion      = "FOUNDRY_VERSION"
    envAdminKey     = "FOUNDRY_ADMIN_KEY"
)
```

## Example: `load.go`

```go
package config

import (
    "github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrykit/confloader"
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
    )
}

func loadPathsFromEnv(c *PathsConfig) confloader.Binder {
    return func() error {
        return confloader.BindEnv(
            confloader.BindField(&c.DataPath, envDataPath, nil),
            confloader.BindField(&c.InstallRoot, envInstallRoot, nil),
            confloader.BindField(&c.ManifestPath, envManifestPath, nil),
            confloader.BindField(&c.HealthAddr, envHealthAddr, nil),
        )
    }
}
```

## Rules recap

- Never call `os.Getenv` directly outside `env_vars.go`-sourced constants and `load.go`.
- Env var name constants are unexported (`envXxx`).
- `Default()` returns a zero-value-safe struct; `Load()` overlays the JSON file then env vars.
- Config is read-only after `Load()` returns.

## See also

- [`../rules/config.md`](../rules/config.md) — full config rules.
- [`../rules/startup.md`](../rules/startup.md) — boot-time loading order.
