# Config

A single typed `Config` struct per app, loaded once when a command runs, from a JSON file plus
environment variable overrides via `libs/foundrykit/confloader`. No `os.Getenv` calls scattered
across the codebase.

## Where config lives

Each app declares its config in a top-level `config/` package:

```
apps/foundryctl/config/
├── config.go      # Config struct + sub-structs + Default()
├── env_vars.go    # unexported env var name constants
└── load.go        # Load(), LoadFromEnv(), per-domain binders
```

The JSON default file name is exposed as a constant:

```go
const DefaultFileName = "foundryctl.json"
```

Cross-module shared config types (logging level, backoff parameters) are embedded into the
app's `Config` struct by importing the relevant library config package directly:

```go
type Config struct {
    Paths   PathsConfig
    Install InstallConfig
    Admin   AdminConfig
    Runtime RuntimeConfig
    Log     colorlog.Config
    Backoff backoff.Config
    Secrets secfuse.Config
}
```

## Loading

`config.Load` in `apps/foundryctl/config/load.go`:

```go
const DefaultFileName = "foundryctl.json"

func Load() (Config, error) {
    return confloader.Load(DefaultFileName, Default(), LoadFromEnv)
}
```

`confloader.Load`:

1. Honours `CONF_FILE` env var if set; otherwise uses the supplied filename.
2. Treats a missing file as not-an-error (defaults remain).
3. Calls the `updater` callback so env vars overlay file values.

## Binding env vars

Env binding is split by subdomain using `confloader.BindField` / `BindEnv`:

```go
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
```

Env var name constants are unexported (`env*`) in `env_vars.go` — never inline string literals.

## Default values

`Default()` returns a zero-value-safe struct with sensible container defaults:

```go
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

## Config system

- All configuration uses the typed `confloader` pattern.
- Environment variables are declared through typed config structs and binders.
- Direct `os.Getenv` usage outside `config/` (or a library's own `LoadFromEnv`) is forbidden.
- Each subdomain config is split into dedicated structs and a dedicated binder function rather
  than one giant `LoadFromEnv`.

## Config organization

Split config concerns across files in the `config/` package:

```
config/
├── config.go      # Config + sub-struct types + Default()
├── env_vars.go    # const env* = "FOUNDRY_..."
└── load.go        # Load() + LoadFromEnv() + per-domain binders
```

Avoid a single monolithic `config.go` that holds types, binders, and constants together.

## Config naming

Name the config package after the domain it configures, not generically. This project uses
`config` (unambiguous for a single-app workspace); in multi-app workspaces prefer a name that
identifies the domain (e.g. `foundryconf`, `runtimeconf`) so imports read clearly at the call
site.

## Forbidden

- `os.Getenv("...")` outside `config/` or library `LoadFromEnv` functions.
- Reading config inside business logic. The activation step or command receives an already-parsed
  `Config` value.
- Mutating `Config` after `Load()` returns.
- Top-level `var cfg = config.Load()` — see [`startup.md`](startup.md).

## Defaults and validation

- Defaults live in `Default()` on the config package, returning a zero-value-safe struct.
- Validation (e.g. port range) runs inside the per-domain binder or immediately after `Load`.
  Fail fast with a descriptive error; the command then logs and exits.

## See also

- [`startup.md`](startup.md) — how config is loaded at command invocation.
- [`wiring.md`](wiring.md) — how the loaded config flows into the
  activation step sequence.
- [`../patterns/confloader-layout.md`](../patterns/confloader-layout.md) — full config package
  layout recipe.
