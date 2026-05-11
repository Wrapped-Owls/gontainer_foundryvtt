# Activation sequence

How `foundryctl run` initialises the container before launching FoundryVTT.

## Overview

`main.go` dispatches to `cmd.Run`, which calls `activate.Prepare`. `activate.Prepare` runs a
pipeline of `step.Step` implementations that fill in a shared `step.State`.

```
main.go
  └─ cmd.Run(args, logger)
       └─ activate.Prepare(ctx, logger)
            └─ step.Run(ctx, logger,
                 step.AppConfig(),
                 step.Secrets(),
                 step.Install(),
                 step.Options(),
                 step.Patches(),
                 step.JSRuntime(),
               )
```

Each step is independent, reads from `*State`, and writes only to the fields it owns.

## `main.go` — full boot sequence

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

## `cmd.Run` — process loop

After activation, `cmd.Run` enters a supervised restart loop using `libs/foundrykit/backoff`:

```go
func Run(_ []string, logger *slog.Logger) int {
    ctx, cancel := context.WithCancelCause(context.Background())
    defer cancel(nil)
    ctx, stop := signalNotifyContext(ctx, syscall.SIGTERM, syscall.SIGINT)
    defer stop()

    state, err := activate.Prepare(ctx, logger)
    if err != nil {
        logger.Error("activation failed", "err", err)
        return 1
    }
    startHealthServer(ctx, logger, state.App.Paths.HealthAddr, state.Runtime.Port)

    mgr := backoff.NewFromConfig(state.App.Backoff)
    for {
        code, err := procspawn.Run(ctx, spec)
        if errors.Is(ctx.Err(), context.Canceled) {
            return code
        }
        dec, _ := mgr.OnFailure(code)
        if err = backoff.Sleep(ctx, dec.Delay); err != nil {
            return code
        }
    }
}
```

## Adding a step

1. Create `apps/foundryctl/internal/activate/step/<name>.go`.
2. Implement `step.Step` (a struct with `Apply`, or a factory function).
3. Add it to `activate.Prepare` at the right position.

```go
// step/minstep.go
type myStep struct{}

func MyStep() Step { return myStep{} }

func (myStep) Apply(_ context.Context, st *State, logger *slog.Logger) error {
    // read from st, write to st
    logger.Info("my step applied")
    return nil
}
```

## See also

- [`../rules/wiring.md`](../rules/wiring.md) — step interface and State.
- [`../rules/startup.md`](../rules/startup.md) — boot sequence rules.
- [`../rules/config.md`](../rules/config.md) — config loading in AppConfig step.
