# Effective Go — the enforced subset

We follow [Effective Go](https://go.dev/doc/effective_go) and the
[Go Code Review Comments](https://go.dev/wiki/CodeReviewComments) wholesale. This file pins the
points that are **non‑negotiable** in this repo and maps each one to the linter that automates it.

## Formatting

- Source is formatted by **`golines`** (line-length enforcement) and **`gofumpt`** (stricter
  `gofmt`). Both are wired in `tools/.golangci.yml` under the `formatters:` block and run via:
  ```sh
  task tools:fmt
  ```
- Never hand‑format. PRs that bypass the formatter will be rejected.
- `goimports` also runs as part of the formatter chain to keep import groups sorted and
  deduplicated.

## Naming

- Identifiers use `MixedCaps` or `mixedCaps`, never `snake_case`. Exported names start uppercase.
- Package names are short, lowercase, single‑word, no underscores or `mixedCaps`. Avoid generic
  names like `util`, `common`, `helpers`. See [`naming.md`](naming.md).
- Do not stutter: `colorlog.New`, not `colorlog.NewColorLog`.

## Doc comments — `godoclint`

- Every exported identifier has a doc comment that starts with the identifier name.
- Package-level doc lives in a file named `doc.go` when the package needs more than a one-liner
  (see `libs/foundrypatch/applier/doc.go`, `libs/foundrypatch/manifest/doc.go`).
- Don't restate the signature in the doc; describe behaviour, contracts, and edge cases.

## Accept interfaces, return structs

- Constructors return concrete types (`*UseCase`, `*Server`).
- Function/method parameters take the **smallest** interface that satisfies their needs, declared
  in the **consumer** package, not exported from the implementor. See [`types.md`](types.md).
- Linter: `ireturn` flags returning interfaces — only allowed via documented exceptions.

## Errors are values

- Functions that can fail return `(T, error)` as the last return value.
- Wrap with `fmt.Errorf("context: %w", err)` when crossing a layer; otherwise return as is.
- Compare with `errors.Is` / `errors.As`. Never `err.Error() == "..."`.
- Linter: `errcheck` (no ignored errors), `wrapcheck` (no naked external errors), `staticcheck`.
- See [`errors.md`](errors.md).

## No `panic` outside `main`

- Libraries never `panic`. Programmer-error invariants use `panic` only in `main.go` boot
  paths where there is no recoverable state.
- Don't use `recover` to mask bugs. Recover only at the top of long‑lived goroutines that own a
  supervised loop, and re‑surface the panic as a structured error log + restart decision.

## Receivers

- Choose value or pointer receiver consistently across **all** methods of a type.
- Use pointer receivers when the method mutates state, when the struct is large, or when the type
  has a sync primitive embedded.
- Method sets matter for interface satisfaction — prefer pointer receivers when in doubt.

## Control flow

- Early returns over deeply nested `if`/`else`. Linter: `nestif`.
- Function length cap: 80 lines (`funlen`); cyclomatic complexity cap: 15 (`cyclop`). Refactor
  before silencing.
- No naked returns in functions longer than 40 lines (`nakedret`).

## Concurrency

- Goroutines have a clear owner and respond to `context.Context` cancellation. See
  [`concurrency.md`](concurrency.md).
- No `time.Sleep` in production code paths to "wait for" something — use channels, `sync.Cond`,
  or `errgroup`.

## Modernisation

- Use `any` instead of `interface{}`. Linter: `modernize`.
- Use `errors.Is`/`As` instead of type assertions on `error`.
- Use `slices`, `maps`, `cmp`, `min`, `max` from the standard library where they fit.

## Forbidden

- `fmt.Println` / `log.Println` in non‑test code. Use `log/slog` ([`logging.md`](logging.md)). The
  `forbidigo` rule is enabled outside `*_test.go`.
- Empty interfaces in public APIs (`any` parameters) without a documented reason.
- Predeclared identifier shadowing (`new`, `len`, `error`, ...). Linter: `predeclared`.

---

## Shell replacement policy

Prefer Go implementations over shell scripts. Operational, runtime, and build orchestration must be implemented in Go whenever feasible.

Shell is permitted only when:
- Unavoidable external tooling integration exists.
- No stable Go alternative is available.

Rationale: Go provides structured error handling, typed execution flow, improved observability, and better OOM diagnostics compared to shell.

## Dependency minimization

Minimize third-party Go dependencies.

- Prefer stdlib and small focused libraries with typed APIs.
- Avoid framework-heavy abstractions, reflection-heavy libraries, and transitive dependency explosions.

## Package naming conventions

Avoid placeholder or `x`-suffixed package names.

Forbidden patterns: `logx`, `procx`, `secretsx`.

Prefer semantic names: `colorlog`, `runtimelog`, `foundrylog`.
