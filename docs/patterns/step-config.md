# Config from step state

How activation steps pass configuration to downstream steps and to command code.

## The pattern

`step.AppConfig()` is always the first step. It calls `config.Load()`, validates the result,
and stores it in `s.App`. All subsequent steps read from `s.App` rather than calling
`config.Load()` again.

```go
// step/config.go
type appConfigStep struct{}

func AppConfig() Step { return appConfigStep{} }

func (appConfigStep) Apply(_ context.Context, s *State, _ *slog.Logger) error {
    cfg, err := appconfig.Load()
    if err != nil {
        return fmt.Errorf("app config: %w", err)
    }
    s.App = cfg
    return nil
}
```

## Reading config downstream

Subsequent steps receive config through `s.App`:

```go
func (installStep) Apply(ctx context.Context, s *State, logger *slog.Logger) error {
    root := s.App.Paths.InstallRoot
    version := s.App.Install.Version
    // ...
}
```

Command functions receive the final `State` from `activate.Prepare` and pull values from it:

```go
func Run(_ []string, logger *slog.Logger) int {
    state, err := activate.Prepare(ctx, logger)
    // ...
    mainScript := filepath.Join(state.Install.Root, state.App.Paths.MainScript)
}
```

## No re-loading

`config.Load()` is called exactly once, inside `AppConfig`. No other step or command function
calls `config.Load()` directly, except `cmd.Version()` which needs a best-effort default for
display purposes only.

## See also

- [`../rules/config.md`](../rules/config.md) — config package layout and loading rules.
- [`../rules/wiring.md`](../rules/wiring.md) — step State ownership.
