# Code placement

Where each file goes in the repository.

> ## ⛔ STOP
>
> New code goes in `apps/`, `libs/`, `tools/`, or `nix/`. Nothing else. Cross-app imports are
> forbidden. Libraries never import apps.

## Top level

| Directory | Role |
|---|---|
| `apps/<name>/` | A deployable binary. Each is its own Go module (`go.mod`). Owns its `main.go`, `config/`, `cmd/`, and `internal/`. |
| `libs/<name>/` | A reusable Go module imported by one or more apps. Each is its own `go.mod`. **Cannot import from `apps/`.** |
| `tools/` | Build/dev tooling: Taskfile, `golangci-lint` config, codegen helpers. Not a runtime dependency of any app. |
| `nix/` | Nix flakes, packages, and image derivations. |
| `patches/` | FoundryVTT patch manifests (`manifest.yaml`). Read at runtime by the activation sequence. |
| `examples/` | Runnable reference deployments (Docker Compose, etc.). |
| `docs/` | This documentation tree. |
| `resources/` | Downloaded runtime artefacts (gitignored content — never committed). |

`go.work` at the repo root lists every module. New modules are added there.

## Inside an app

Each app uses a flat top-level layout of public packages (`config/`, `cmd/`) and an `internal/`
subtree for everything not exported:

```
apps/foundryctl/
├── main.go              # thin: parse subcommand, set up logger, dispatch to cmd/
├── go.mod
├── config/              # typed Config + env loaders — see config.md
│   ├── config.go        # Config struct + sub-structs + Default()
│   ├── env_vars.go      # unexported env var name constants
│   └── load.go          # Load() + LoadFromEnv() + per-domain binders
├── cmd/                 # one file per subcommand
│   ├── run.go           # foundryctl run — main process loop
│   ├── healthcheck.go
│   ├── options.go
│   ├── version.go
│   └── signal.go        # shared OS signal helper
└── internal/
    ├── activate/         # activation pipeline (step sequence)
    │   ├── activate.go   # Prepare() — assembles and runs the step pipeline
    │   ├── install/      # install candidate detection and download
    │   └── step/         # Step interface, State, and individual step implementations
    └── secfuse/          # secret fuse config loader (runtime credentials)
```

HTTP handlers, gRPC servers, and database repositories do not exist in this project — it is a
CLI. Logic that would go in those layers instead lives in `internal/activate/step/`.

## Inside a lib

Libraries expose **types and functions**. They do not start goroutines, read environment
variables outside their own `LoadFromEnv`, or open connections on import.

```
libs/<name>/
├── go.mod
└── <package>/           # one sub-directory per concern
    ├── <concern>.go     # exported API surface
    ├── <concern>_test.go
    └── internal/        # implementation details not exported outside this package
```

Examples from this repo:

```
libs/foundrykit/
├── backoff/      — exponential backoff with jitter
├── colorlog/     — slog handler with ANSI colour output
├── confloader/   — typed JSON+env config loader
├── jsonhttp/     — generic typed HTTP request helper
└── procspawn/    — subprocess launch and signal management

libs/foundryruntime/
├── config/       — typed Foundry runtime config + env binders
├── health/       — health-check HTTP probe
├── jsruntime/    — bun/node runtime detection
└── lifecycle/    — installed version detection + Options file writer
```

## Workspace topology

Required top-level folders:

| Directory | Role |
|---|---|
| `apps/` | Deployable binaries |
| `libs/` | Reusable libraries |
| `tools/` | Developer/build/codegen tooling |
| `test/` | Integration/e2e-only modules and dependencies (added when needed) |
| `docs/` | Engineering documentation |
| `examples/` | Runnable examples and reference deployments |
| `nix/` | Nix flakes/modules/packages |
| `patches/` | Runtime patch manifests |

Forbidden top-level folders unless explicitly approved via PR:

- `deploy/`
- `assets/`
- `data/`
- Legacy migration folders
- Ad-hoc scripts directories

## Workspace/module rules

- Each workspace member owns its own `go.mod`.
- The repository root owns a single `go.work`.
- Cross-module local references use `replace` directives during workspace development.
- Do not create excessively granular modules. Prefer one module with multiple packages over many
  tiny modules with one package each. Modules exist for independently reusable/versionable
  boundaries, not for every package.

## Package granularity

- Packages may contain multiple files.
- Do not collapse unrelated logic into giant files.
- Prefer one responsibility per package and one action/concern per file.
- Use `internal/` subpackages for complex domains.
- Avoid single-file god packages and avoid splitting modules when packages are sufficient.

## Internal vs library boundaries

- Libraries under `libs/` represent genuinely reusable components.
- Application-specific logic belongs under `apps/<app>/internal/...`.
- Do not promote code into `libs/` unless it is reused by multiple apps, independently
  composable, and stable enough to act as shared infrastructure.

## Forbidden moves

- `apps/foundryctl` importing from another `apps/` module. Extract the shared piece into a lib.
- `libs/<name>` importing from `apps/...`. Direction is one-way.
- A file at the root of a module that is not exported API. Push it into `internal/`.
- New top-level directories outside the table above. Discuss in a PR before adding one.

## See also

- [`../patterns/app-skeleton.md`](../patterns/app-skeleton.md) — new app skeleton
- [`../patterns/bootstrap-and-di.md`](../patterns/bootstrap-and-di.md) — activation sequence example
