# Sources — the `fourcery` acquisition pipeline

`libs/fourcery` is the unified library for obtaining a Foundry installation. It handles
authentication, release URL resolution, archive acquisition, and local artefact sourcing.

## Sources directory layout

```
/foundry/sources/                  ← FOUNDRY_SOURCES_DIR (operator-mounted)
├── foundryvtt_v14.361.2.zip       ← local zip artefact
└── foundryvtt_v14.361.2/          ← pre-extracted folder artefact
    └── resources/app/package.json
```

Filename convention — the controller recognises files and directories whose names match:

```
(?i)^foundryvtt[_\-]?v?(\d+\.\d+(?:\.\d+)?)(?:\.zip)?$
```

Examples of valid names: `foundryvtt_v14.361.2.zip`, `FoundryVTT-14.361.2.zip`,
`foundryvtt14.361.zip`, `foundryvtt_v14.361.2/`.

## Version probe order

For each candidate artefact the controller probes the version in this order:

1. **Filename** — regex match above.
2. **`resources/app/package.json`** inside the artefact — `"version"` field.
   - Zip: read via central-directory scan (no extraction).
   - Folder: read from disk.

If neither probe succeeds the source returns `source.ErrVersionUnknown`; the resolver may still
use it (e.g. a presigned URL) but only as a last resort.

## Strategy interface contract

Every source implements `source.Source`:

```go
type Source interface {
    Kind() Kind
    Describe() string
    Probe(ctx context.Context) (string, error)
    Materialise(ctx context.Context, dst string) (Result, error)
}
```

- `Probe` must be **read-only** and **fast** (no network if avoidable). Return
  `ErrVersionUnknown` rather than making a network call.
- `Materialise` receives an **empty staging directory** managed by `Forge`. It writes into `dst`
  and returns. It must not touch any path outside `dst`.
- Both methods must be **safe to call concurrently** from different goroutines (implementations
  are expected to be stateless after construction).

## Adding a new source kind

1. Create `libs/fourcery/source/<kind>.go` implementing `source.Source`.
2. Add a `Kind<Name>` constant in `source/source.go`.
3. Register the new source in `source/registry.go`:`Registry.Enumerate`.
4. Add unit tests in `source/<kind>_test.go`; add an integration test (build tag
   `integration`) that materialises into `t.TempDir()`.

## Forbidden

- Writing into `FOUNDRY_SOURCES_DIR` from the controller — it is operator-owned.
- In-place extraction (modifying the zip or folder in sources dir instead of staging to dst).
- Materialise calling `os.MkdirTemp` itself — that is Forge's responsibility.
- Network calls inside `Probe` (except `sessionSource` which knows its version from config).

## `FOUNDRY_VERSION=latest`

The special sentinel value `latest` changes the resolver priority so local artefacts are
**preferred over remote sources**:

1. Highest-version zip or folder in `FOUNDRY_SOURCES_DIR`.
2. Newest installed candidate in `FOUNDRY_INSTALL_ROOT`.
3. `FOUNDRY_RELEASE_URL` presigned URL — only if steps 1–2 found nothing.
4. Session source — only if steps 1–3 found nothing.

Use `latest` when you have a local zip or pre-installed folder and do not want an expired or
slow presigned URL to take precedence. If no local artefact exists, the container still downloads
as normal.

## See also

- [`../patterns/auth-session.md`](../patterns/auth-session.md) — auth session and cookie reuse.
- [`patches.md`](patches.md) — post-install patch system.
- [`code-placement.md`](code-placement.md) — `libs/` placement rules.
