# Integration tests

Integration tests require real I/O (network, filesystem, subprocess). They use the
`//go:build integration` tag and run only in CI's integration job.

## Build tag

Every integration test file begins with:

```go
//go:build integration
```

Run them with:

```sh
go test -tags=integration ./...
```

## Patch applier integration test

Tests for `libs/foundrypatch` use a temp filesystem tree:

```go
//go:build integration

package applier_test

import (
    "os"
    "path/filepath"
    "testing"

    "github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrypatch/applier"
    "github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrypatch/manifest"
)

func TestApplier_FileReplace(t *testing.T) {
    t.Parallel()
    root := t.TempDir()

    // Set up the source file
    src := filepath.Join(root, "patch.txt")
    if err := os.WriteFile(src, []byte("patched"), 0o644); err != nil {
        t.Fatal(err)
    }

    m := manifest.Manifest{
        Patches: []manifest.Entry{{
            Actions: []manifest.Action{{
                Type:   "filereplace",
                Target: "resources/app/target.txt",
                Source: src,
            }},
        }},
    }

    a := applier.New(root)
    if err := a.Apply(t.Context(), m.Patches); err != nil {
        t.Fatal(err)
    }

    got, err := os.ReadFile(filepath.Join(root, "resources/app/target.txt"))
    if err != nil {
        t.Fatal(err)
    }
    if string(got) != "patched" {
        t.Fatalf("got %q, want %q", got, "patched")
    }
}
```

## Activation integration test

Call `activate.Prepare` against a test environment built with `t.TempDir()`:

```go
//go:build integration

package activate_test

func TestPrepare_NoNetwork(t *testing.T) {
    t.Parallel()
    root := t.TempDir()

    t.Setenv("FOUNDRY_INSTALL_ROOT", root)
    t.Setenv("FOUNDRY_DATA_PATH", filepath.Join(root, "data"))

    // Place a fake main.mjs so the install step sees a present installation
    script := filepath.Join(root, "resources/app/main.mjs")
    if err := os.MkdirAll(filepath.Dir(script), 0o755); err != nil {
        t.Fatal(err)
    }
    _ = os.WriteFile(script, []byte(""), 0o644)

    state, err := activate.Prepare(t.Context(), slog.Default())
    if err != nil {
        t.Fatal(err)
    }
    if state.Install.Root != root {
        t.Fatalf("install root = %q, want %q", state.Install.Root, root)
    }
}
```

## Rules

- Use `t.TempDir()` for any filesystem work — never write to real paths in tests.
- Use `t.Setenv` to override env vars — it is automatically cleaned up after the test.
- Use `t.Context()` (Go 1.24+) as the context; it is cancelled when the test ends.
- Do not hardcode ports or paths that may conflict with parallel tests.

## See also

- [`../rules/testing.md`](../rules/testing.md) — test layout rules.
- [`../rules/patches.md`](../rules/patches.md) — `foundrypatch` test fixtures.
