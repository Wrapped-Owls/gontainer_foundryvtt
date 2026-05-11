# Security

## Secrets

- Secrets (`FOUNDRY_PASSWORD`, `FOUNDRY_SESSION`, `FOUNDRY_ADMIN_KEY`, `FOUNDRY_LICENSE_KEY`,
  `FOUNDRY_PASSWORD_SALT`) come from environment variables loaded by `libs/foundrykit/confloader`
  into the typed `Config` struct (see [`config.md`](config.md)).
- Never commit secrets. `.env` files are local-only and listed in `.gitignore`. Container
  secrets are provided via environment variables at runtime.
- The `gosec` linter is enabled; obey its findings. Suppressions need a `// #nosec G<n>:
  <justification>` comment.
- The `secfuse` package (`apps/foundryctl/internal/secfuse`) handles runtime credential
  injection into the FoundryVTT options file. The loaded secret values must not be logged —
  see [`logging.md`](logging.md).

## Logs

- Never log a secret, an auth credential, or the raw value of any `envPassword`, `envSession`,
  `envAdminKey`, or `envLicenseKey` field.
- Paths, version strings, and port numbers are safe to log.

## Input validation

- All values that come from environment variables are validated in the per-domain binder (e.g.
  port range checked in `loadRuntimeFromEnv`) or immediately after `config.Load()` returns.
- Downloaded archives are extracted into a controlled target directory. Validate archive entry
  paths before extraction to prevent path traversal.

## Container and supply chain

- Base images are pinned by digest in `Dockerfile`. No `:latest` tags.
- Go dependencies are pinned in `go.mod`; `go mod tidy` is run via the Makefile/Taskfile,
  not by hand.
- Downloaded FoundryVTT releases are verified against the presigned URL returned by the
  authenticated FoundryVTT API — do not fetch from arbitrary URLs.

## See also

- [`config.md`](config.md) — typed `Config` and env var loading.
- [`logging.md`](logging.md) — what must not appear in log output.
- [`../rules/http-clients.md`](http-clients.md) — authenticated HTTP client for `fourcery`.
