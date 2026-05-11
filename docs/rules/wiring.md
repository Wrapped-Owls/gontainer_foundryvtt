# Wiring — constructor injection and the activation step sequence

This project uses **constructor injection** throughout: every dependency (logger, config, HTTP
client) is passed as a function parameter, never read from a package-level global. For the
startup activation sequence, the `step.Run` function orchestrates a pipeline of `Step`
implementations, each filling in the shared `State`.

## Constructor injection

Every function or struct that depends on an external collaborator receives it as a parameter:

```go
func Run(_ []string, logger *slog.Logger) int {
    state, err := activate.Prepare(ctx, logger)
    ...
}
```

No global logger, no package-level config. The dependency graph is explicit and visible in
function signatures.

## The activation step sequence

`apps/foundryctl/internal/activate/` orchestrates startup through a pipeline of `Step`
implementations defined in `internal/activate/step/`:

```go
// step.go — the strategy interface
type Step interface {
    Apply(ctx context.Context, s *State, logger *slog.Logger) error
}

// State accumulates the result of each preparation step.
type State struct {
    App       appconfig.Config
    Runtime   runtimecfg.Config
    JSRuntime jsruntime.Runtime
    Install   install.Install
}

// Run executes steps in order; the first error stops the pipeline.
func Run(ctx context.Context, logger *slog.Logger, steps ...Step) (State, error) {
    var s State
    for _, step := range steps {
        if err := step.Apply(ctx, &s, logger); err != nil {
            return State{}, err
        }
    }
    return s, nil
}
```

`activate.Prepare` assembles the pipeline:

```go
func Prepare(ctx context.Context, logger *slog.Logger) (State, error) {
    return step.Run(ctx, logger,
        step.AppConfig(),   // loads config + env
        step.Secrets(),     // loads secfuse secrets
        step.Install(),     // resolves or downloads Foundry installation
        step.Options(),     // prepares runtime options
        step.Patches(),     // applies foundrypatch manifest
        step.JSRuntime(),   // detects bun or node
    )
}
```

Each step reads from and writes only the `State` fields it owns. Steps are independent of each
other and receive everything they need through `*State` and the `*slog.Logger`.

## Adding a new step

1. Create `apps/foundryctl/internal/activate/step/<name>.go`.
2. Implement `Step` — either a struct with `Apply` or a factory function returning one.
3. Read from `s` for context; write only to the fields this step owns.
4. Add it to `activate.Prepare` at the correct position in the pipeline.

## Testing

In unit tests, construct a `State` directly and call `Apply`:

```go
func TestInstallStep(t *testing.T) {
    t.Parallel()
    s := &step.State{App: config.Default()}
    if err := step.Install().Apply(context.Background(), s, slog.Default()); err != nil {
        t.Fatal(err)
    }
    // assert s.Install fields are populated
}
```

For end-to-end activation, call `activate.Prepare` in an integration test against a test
environment.

## Forbidden

- Package-level `var` holding mutable state shared between steps.
- Steps reading from global variables instead of `*State`.
- Calling `Apply` outside of `step.Run` except in tests.
- Storing the logger or config on a package-level variable.

## See also

- [`../patterns/bootstrap-and-di.md`](../patterns/bootstrap-and-di.md) — activation sequence
  recipe.
- [`startup.md`](startup.md) — the full boot sequence from `main()` to process launch.
