# `rules/`

Mandatory rules for `gontainer_foundryvtt`. Each file describes one decision area; together they
form the engineering contract for any new or modified Go code in the repo.

The rules are designed to be **enforceable**. Whenever possible, each rule cites the linter from
[`.github/.golangci.yml`](../../.github/.golangci.yml) that automates it. Where no linter exists,
the rule is still mandatory and is enforced via code review.

## Lint baseline

```sh
golangci-lint run --config .github/.golangci.yml ./...
```

The configuration enables (non-exhaustive): `errcheck`, `govet`, `staticcheck`, `unused`,
`cyclop`, `dupl`, `funlen`, `goconst`, `gosec`, `importas`, `interfacebloat`, `ireturn`,
`modernize`, `nakedret`, `nestif`, `noctx`, `prealloc`, `predeclared`, `sloglint`, `tagalign`,
`unconvert`, `unparam`, `whitespace`, `wrapcheck`.

## Index

| Rule | Topic |
|---|---|
| [`effective-go.md`](effective-go.md) | The Effective Go subset we enforce |
| [`naming.md`](naming.md) | Go naming for packages, types, files |
| [`imports.md`](imports.md) | Three-group import layout + aliasing |
| [`errors.md`](errors.md) | Sentinel errors, `%w` wrapping |
| [`logging.md`](logging.md) | `log/slog` + `colorlog` structured logging |
| [`commits.md`](commits.md) | Conventional commits with emoji prefix |
| [`security.md`](security.md) | Secrets, credential handling, supply chain |
| [`config.md`](config.md) | Typed `Config` via `confloader` |
| [`startup.md`](startup.md) | Zero side effects in `init()`; boot sequence |
| [`concurrency.md`](concurrency.md) | Goroutine ownership, `context` |
| [`types.md`](types.md) | Named structs and small interfaces |
| [`code-placement.md`](code-placement.md) | `apps/` vs `libs/` vs `tools/` vs `nix/` |
| [`testing.md`](testing.md) | Colocated tests, integration with build tag |
| [`wiring.md`](wiring.md) | Constructor injection and step sequence |
| [`patches.md`](patches.md) | `foundrypatch` manifest and applier |
| [`sources.md`](sources.md) | `fourcery` acquisition pipeline and sources directory |
| [`http-clients.md`](http-clients.md) | External API clients via `fourcery` |
| [`transport.md`](transport.md) | CLI command dispatch and typed HTTP |
| [`interop.md`](interop.md) | Cross-module import rules and consumer-defined interfaces |
| [`build.md`](build.md) | Nix build system, OCI rules, runtime version selection |
