# Step layout

How to structure an activation step and its sub-packages.

## One step, one file

Each step lives in `apps/foundryctl/internal/activate/step/<name>.go`. The file exports a
single factory function returning a `Step`:

```
step/
├── step.go       # Step interface, State, Run
├── config.go     # step.AppConfig()
├── install.go    # step.Install()
├── jsruntime.go  # step.JSRuntime()
├── options.go    # step.Options()
├── patches.go    # step.Patches()
└── secrets.go    # step.Secrets()
```

## Factory pattern

```go
// step/install.go
package step

import (
    "context"
    "log/slog"
)

type installStep struct{}

// Install returns the step that resolves or downloads the FoundryVTT installation.
func Install() Step { return installStep{} }

func (installStep) Apply(ctx context.Context, s *State, logger *slog.Logger) error {
    // 1. Read s.App.Paths.InstallRoot and s.App.Install.Version
    // 2. Detect or download the installation
    // 3. Write s.Install
    return nil
}
```

## Sub-packages for complex steps

When a step's logic is large enough to split, put the helpers in a sibling sub-package rather
than bloating the step file:

```
internal/activate/
├── activate.go
├── install/         # helpers for install detection + download
│   ├── candidates.go
│   ├── download.go
│   ├── install.go
│   └── session.go
└── step/
    └── install.go   # thin: calls install.Resolve(...)
```

The step file stays thin — it delegates to the sub-package:

```go
func (installStep) Apply(ctx context.Context, s *State, logger *slog.Logger) error {
    result, err := install.Resolve(ctx, s.App, logger)
    if err != nil {
        return fmt.Errorf("install step: %w", err)
    }
    s.Install = result
    return nil
}
```

## State ownership

Each step documents which `State` fields it **reads** and which it **writes**. A step must not
write fields it doesn't own.

| Step | Reads | Writes |
|---|---|---|
| `AppConfig` | — | `s.App` |
| `Secrets` | `s.App` | `s.App.Secrets` (mutates) |
| `Install` | `s.App` | `s.Install` |
| `Options` | `s.App`, `s.Install` | `s.Runtime` |
| `Patches` | `s.App`, `s.Install` | — (side effect: applies patches) |
| `JSRuntime` | `s.App`, `s.Install` | `s.JSRuntime` |

## See also

- [`../rules/wiring.md`](../rules/wiring.md) — step interface and Run.
- [`../rules/code-placement.md`](../rules/code-placement.md) — `internal/activate/` placement.
