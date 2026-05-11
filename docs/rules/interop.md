# Interop — cross-module calls

Rules for when code in one module (app or lib) calls into another module.

## Import direction

```
apps/foundryctl
  ├─► libs/fourcery
  ├─► libs/foundrykit
  ├─► libs/foundrypatch
  └─► libs/foundryruntime

libs/fourcery
  └─► libs/foundrykit  (jsonhttp, backoff)

libs/foundrypatch
  └─► libs/foundrykit  (jsonhttp, backoff)

libs/foundryruntime
  └─► (no lib dependencies within this workspace)
```

`libs/` packages never import from `apps/`. Direction is one-way.

## Consumer-defined interfaces

When an `internal/` package needs to call a lib function but wants to remain testable without
importing the concrete type, declare a minimal interface in the **consumer** package:

```go
// internal/activate/step/install.go

// downloader is the subset of release.Fetch this step needs.
type downloader interface {
    Fetch(ctx context.Context, sess *auth.Session, version string, opts release.FetchOptions) (string, error)
}
```

`libs/fourcery/release.Fetch` satisfies this structurally. Tests inject a fake without
importing the real implementation.

## Cross-lib dependencies

When a lib (`fourcery`) depends on another lib (`foundrykit`), the dependency is declared
in its `go.mod` with a `replace` directive pointing to the workspace path. No lib may pull in
`apps/` through a transitive chain.

## Forbidden

- A `libs/` module importing from `apps/foundryctl` (or any `apps/` module).
- Importing `internal/` packages from outside the parent module — the Go compiler enforces this.
- Cross-app imports (single app in this workspace, but the rule applies if others are added).
- Circular dependencies between libs — resolve by extracting a shared primitive into a third lib
  or moving the symbol into the consumer package.

## See also

- [`code-placement.md`](code-placement.md) — workspace topology and module rules.
- [`types.md`](types.md) — consumer-defined interfaces.
- [`imports.md`](imports.md) — three-group import ordering.
