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

Lives in `libs/foundrypatch/applier/applier_integration_test.go`, same package as the
code under test. Drive the real `manifest` types against a `t.TempDir()` tree:

```go
//go:build integration

package applier

patches := []manifest.Patch{{
    ID:       "core-fix",
    Versions: ">=1.0.0",
    Actions: []manifest.Action{{
        Type:    manifest.ActionFileReplace,
        Dest:    "resources/app/target.txt",
        Content: "patched",
    }},
}}

applier := &Applier{Root: root}
if err := applier.Apply(t.Context(), patches, nil); err != nil {
    t.Fatalf("apply: %v", err)
}
```

`Applier` is a struct literal, not a constructor. `Apply` takes three arguments; the
third is a `logf`. A `file-replace` carries its payload inline in `Content` - there is
no file-to-file copy action. Worth proving beyond the happy path: a `Dest` that escapes
the root is refused and leaves the root empty, and a second `Apply` of the same patch
does not re-apply.

## Activation integration test

Lives in `apps/foundryctl/internal/activate/activate_test.go`, same package. It proves
the boot path works offline against a populated volume, which is what happens on every
container restart.

```go
//go:build integration

package activate

state, err := Prepare(t.Context(), slog.New(slog.DiscardHandler))
```

Four things the test has to set up, each learned the hard way:

- **Do not call `t.Parallel()`.** The test mutates process-global env, and `t.Setenv`
  panics outright in a parallel test.
- **Point every path-bearing env var into `t.TempDir()`**, not just the install root and
  data path. `SourcesDir` defaults to the literal `/foundry/sources` and is not derived
  from `FOUNDRY_INSTALL_ROOT`, so `EnsureDirs` will `MkdirAll` on the real host without
  an override. The same holds for the manifest, license cache, profiles file, secret
  file and `CONF_FILE`.
- **Plant the install as `<root>/foundryvtt_v<semver>/resources/app/` with a
  `package.json`**, and set `FOUNDRY_VERSION`. Only then does the resolver exercise the
  documented "already-installed candidate matching the version" path instead of falling
  through to the latest candidate.
- **Give the JS runtime a deterministic `PATH`**: a temp `bin/` holding executable stubs
  named `bun`, `node22` and `node24`. Otherwise `exec.LookPath` reads the host and the
  result depends on what happens to be installed.

Assert on more than a nil error: the resolved install root and version, the data path,
and the runtime the Foundry major selects. `PrepareProfile` with a different version is
worth the same check, since it re-runs the runtime resolution.

## Rules

- Use `t.TempDir()` for any filesystem work; never write to real paths in tests.
- Use `t.Setenv` to override env vars; it is cleaned up automatically.
- Use `t.Context()` (Go 1.24+) as the context; it is cancelled when the test ends.
- Never combine `t.Parallel()` with `t.Setenv` or `t.Chdir`: the runtime panics.
- Tests live in the same package as the code under test. No black-box `<pkg>_test`
  package, here or anywhere - see [`../rules/testing.md`](../rules/testing.md).
- Do not hardcode ports or paths that may conflict with parallel tests.

## See also

- [`../rules/testing.md`](../rules/testing.md): test layout rules.
- [`../rules/patches.md`](../rules/patches.md): `foundrypatch` test fixtures.
