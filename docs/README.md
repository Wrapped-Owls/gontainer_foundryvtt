# `docs/`

Engineering documentation for `gontainer_foundryvtt`.

> **Rules tell you what. Patterns show you how.**

## How to navigate

1. Start at [`README.md`](../README.md) at the repository root.
2. For mandatory rules (what to do, what not to do), read [`rules/`](rules/).
3. For implementation recipes (how to do it, with code), read [`patterns/`](patterns/).

## Contents

### [`rules/`](rules/) — Rules (imperative, enforceable)

Mandatory standards for new and modified code. Each file describes one decision that is not
open to renegotiation per PR.

- [`effective-go.md`](rules/effective-go.md) — the Effective Go subset we enforce
- [`naming.md`](rules/naming.md) — Go identifier and package naming
- [`imports.md`](rules/imports.md) — three-group import layout and aliasing
- [`errors.md`](rules/errors.md) — sentinel errors and `%w` wrapping
- [`logging.md`](rules/logging.md) — `log/slog` + `colorlog` structured logging
- [`commits.md`](rules/commits.md) — emoji + conventional commit style
- [`security.md`](rules/security.md) — secrets, credential handling, supply chain
- [`config.md`](rules/config.md) — typed `Config` via `foundrykit/confloader`
- [`startup.md`](rules/startup.md) — zero side effects in `init()`; CLI boot sequence
- [`concurrency.md`](rules/concurrency.md) — goroutine ownership and `context` propagation
- [`types.md`](rules/types.md) — named structs over `map[string]any`; consumer-side interfaces
- [`code-placement.md`](rules/code-placement.md) — `apps/` vs `libs/` vs `tools/` vs `nix/`
- [`testing.md`](rules/testing.md) — colocated tests, integration via build tag
- [`wiring.md`](rules/wiring.md) — constructor injection + step sequence
- [`patches.md`](rules/patches.md) — `foundrypatch` manifest and applier system
- [`sources.md`](rules/sources.md) — `fourcery` acquisition pipeline and sources directory
- [`http-clients.md`](rules/http-clients.md) — external API clients via `fourcery`
- [`transport.md`](rules/transport.md) — CLI command dispatch and typed HTTP
- [`interop.md`](rules/interop.md) — cross-module import rules and consumer interfaces
- [`build.md`](rules/build.md) — Nix build system, OCI image rules, runtime version selection

### [`patterns/`](patterns/) — Patterns (cookbook)

Implementation recipes with copyable templates.

- [`bootstrap-and-di.md`](patterns/bootstrap-and-di.md) — activation sequence from `main()` to Foundry launch
- [`step-config.md`](patterns/step-config.md) — config flowing from `AppConfig` step
- [`usecase-layout.md`](patterns/usecase-layout.md) — step factory pattern and sub-package structure
- [`patch-manifest.md`](patterns/patch-manifest.md) — patch manifest format and action types
- [`jsonhttp.md`](patterns/jsonhttp.md) — `jsonhttp.Request` typed HTTP call pattern
- [`integration-tests.md`](patterns/integration-tests.md) — tests with `t.TempDir()` and `t.Setenv`
- [`auth-session.md`](patterns/auth-session.md) — `fourcery` auth session and cookie reuse
- [`app-skeleton.md`](patterns/app-skeleton.md) — new `apps/<name>/` skeleton
- [`confloader-layout.md`](patterns/confloader-layout.md) — config package layout with `confloader`
- [`nix-builds.md`](patterns/nix-builds.md) — Nix flake, Go binary, and container image build
- [`procspawn.md`](patterns/procspawn.md) — `procspawn` spec and backoff restart loop
