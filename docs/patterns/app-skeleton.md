# App skeleton

How to bootstrap a new binary under `apps/`.

## Directory layout

```
apps/<name>/
├── main.go              # dispatch: set up logger, parse subcommand, call cmd.*
├── go.mod               # module github.com/wrapped-owls/gontainer_foundryvtt/apps/<name>
├── config/
│   ├── config.go        # Config struct + sub-structs + Default()
│   ├── env_vars.go      # unexported env var name constants
│   └── load.go          # Load() + LoadFromEnv() + per-domain binders
├── cmd/
│   └── run.go           # first subcommand
└── internal/
    └── activate/        # add an activation pipeline if needed
        ├── activate.go
        └── step/
```

## `go.mod`

```
module github.com/wrapped-owls/gontainer_foundryvtt/apps/<name>

go 1.26.2

require (
    github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrykit v0.0.0
)

replace (
    github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrykit => ../../libs/foundrykit
)
```

Register in `go.work`:

```
use (
    apps/<name>
    ...
)
```

## `main.go`

```go
package main

import (
    "fmt"
    "log/slog"
    "os"

    "<module>/apps/<name>/cmd"
    "github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrykit/colorlog"
)

func main() {
    logger := colorlog.New("<name>", colorlog.LevelFromEnv())
    slog.SetDefault(logger)

    args := os.Args[1:]
    sub := "run"
    if len(args) > 0 && args[0][0] != '-' {
        sub, args = args[0], args[1:]
    }

    switch sub {
    case "run":
        os.Exit(cmd.Run(args, logger))
    default:
        fmt.Fprintf(os.Stderr, "<name>: unknown subcommand %q\n", sub)
        os.Exit(2)
    }
}
```

## `config/config.go`

```go
package config

type Config struct {
    // add sub-structs here
}

func Default() Config {
    return Config{}
}
```

## `config/load.go`

```go
package config

import "github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrykit/confloader"

const DefaultFileName = "<name>.json"

func Load() (Config, error) {
    return confloader.Load(DefaultFileName, Default(), LoadFromEnv)
}

func LoadFromEnv(c *Config) error {
    return confloader.BindEnv(
        // add binders here
    )
}
```

## `cmd/run.go`

```go
package cmd

import (
    "log/slog"

    "<module>/apps/<name>/config"
)

func Run(_ []string, logger *slog.Logger) int {
    cfg, err := config.Load()
    if err != nil {
        logger.Error("config load failed", "err", err)
        return 1
    }
    _ = cfg
    return 0
}
```

## See also

- [`../rules/code-placement.md`](../rules/code-placement.md) — module topology.
- [`../rules/config.md`](../rules/config.md) — config package rules.
- [`bootstrap-and-di.md`](bootstrap-and-di.md) — activation sequence.
