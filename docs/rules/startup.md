# Startup — no side effects on import

Importing a package must **only define** symbols. It must not open connections, start goroutines,
read files, register global handlers, or schedule timers. All of that happens inside `main()`.

## Why

- Imports stay fast and predictable.
- Unit tests can import any package without paying for I/O.
- `go vet`, `gopls`, `golangci-lint`, `staticcheck` all import every package they analyse — they
  must not trigger network calls.
- Boot order becomes explicit and traceable.

## The rule

```go
// ❌ Bad — I/O at package load
var cfg = config.Load()

// ❌ Bad — background goroutine at import
func init() {
    go monitorProcess()
}

// ❌ Bad — env read at import
var dataPath = os.Getenv("FOUNDRY_DATA_PATH")
```

```go
// ✅ Good — define only
type Installer struct {
    root string
}

func New(root string) *Installer {
    return &Installer{root: root}
}
```

## Boot sequence

`apps/foundryctl/main.go` performs only these steps:

```go
func main() {
    logger := colorlog.New("foundryctl", colorlog.LevelFromEnv())
    slog.SetDefault(logger)

    args := os.Args[1:]
    sub := "run"
    if len(args) > 0 && !startsWithFlag(args[0]) {
        sub, args = args[0], args[1:]
    }

    switch sub {
    case "run":
        os.Exit(cmd.Run(args, logger))
    case "healthcheck":
        os.Exit(cmd.Healthcheck(args, logger))
    case "options":
        os.Exit(cmd.Options(args, logger))
    case "version":
        cmd.Version()
    default:
        fmt.Fprintf(os.Stderr, "foundryctl: unknown subcommand %q\n", sub)
        os.Exit(2)
    }
}
```

Config is loaded inside the command functions (via `config.Load()`), not at startup. The logger
is the only shared resource created in `main()`.

## `init()` is almost always wrong

`init()` is acceptable in exactly two cases:

1. **Driver registration** in a `_` import inside `main.go`. Document each blank import with a
   one-line comment.
2. **Compile-time invariant checks** that have no I/O:

   ```go
   var _ step.Step = (*installStep)(nil) // interface assertion
   ```

Anything else — env reads, file loads, goroutines — goes in a constructor or in the function
the caller invokes explicitly.

## Test collection imports too

`go test ./...` imports every package under the current directory. A package with import-time I/O
makes the test collector pay that cost (and may break CI on restricted networks). Keep imports
pure.

## Where structured logging is configured

In `main.go`, **before** any other call:

```go
logger := colorlog.New("foundryctl", colorlog.LevelFromEnv())
slog.SetDefault(logger)
```

Libraries never call `slog.SetDefault`. They receive a `*slog.Logger` via constructor or
function parameter.

## Legacy references

Do not preserve comments, docs, or compatibility references to removed legacy implementations.
When a subsystem is replaced:

- Remove migration-era comments.
- Remove references to deleted scripts/tools.
- Remove historical wording from doc files.

The repository should describe the current architecture only. History lives in git.

## Dead code policy

Unused code must be removed immediately. Indicators include:

- Unreachable packages.
- Unreferenced commands.
- Unused tooling.
- Stale migration helpers.
- Compatibility wrappers with zero call sites.

## See also

- [`config.md`](config.md) — how configuration is loaded at command invocation.
- [`wiring.md`](wiring.md) — the step activation sequence.
