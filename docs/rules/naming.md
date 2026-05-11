# Naming

Idiomatic Go naming. No domain‑language exceptions: code is in **English**.

## Packages

- Lowercase, single word, no underscores, no `mixedCaps`.
  - Good: `confloader`, `colorlog`, `procspawn`, `jsruntime`, `backoff`.
  - Bad: `conf_loader`, `colorLogPackage`, `myutils`, `helpers`.
- The package name is the user's first contact with the API; pick one that reads well at the call
  site (`colorlog.New(...)`, not `colorlogPkg.NewColorLogger(...)`).
- Avoid stuttering: a function exported from package `parsers` is `parsers.Parse`, not
  `parsers.ParseParser`. Similarly `colorlog.New`, not `colorlog.NewColorLog`.

## Identifiers

- Exported: `PascalCase` (`UseCase`, `NewWebServer`, `RegisterServices`).
- Unexported: `mixedCaps` (`associatedUser`, `parseToken`).
- Constants: `PascalCase` if exported; `mixedCaps` otherwise. Use `UPPER_SNAKE` only for env‑var
  keys (e.g. `EnvConfFile`), never as Go identifiers themselves.
- Acronyms keep their case: `ID`, `HTTP`, `URL`, `JWT`, `RPC`. Never `Id`, `Http`, `Url`, `Jwt`.
  - `HTTPClient`, `userID`, `parseURL`.
- Receiver names are 1–3 letters and consistent across a type's methods (`uc *UseCase`,
  `s *Server`).

## Files

- `snake_case.go`. Test files are `<source>_test.go` colocated with the source.
- One file per top‑level concept when the file would otherwise grow over ~300 lines. Group by
  action verb (`create_secret_friend.go`, `list_secret_friends.go`) — see
  [`../patterns/usecase-layout.md`](../patterns/usecase-layout.md).
- Keep types and their constructors together; don't scatter `New` into a separate file.

## Modules and Go workspace

- Module path matches the import path: `github.com/wrapped-owls/gontainer_foundryvtt/<dir>`.
- Each `apps/<name>/` and each `libs/<name>/` is its own module. The `go.work` file at the repo
  root lists every member; new modules **must** be added there.

## Private inside `internal/`

- Anything not part of the public surface of an app or library lives under `internal/`. The
  compiler enforces this. Don't leak experimental APIs by promoting them out of `internal/`
  prematurely.

## Test names

- `func TestThing_Behaviour(t *testing.T)` for table‑driven suites; subtests use a human sentence
  (`"returns ErrNotFound when user is missing"`).
- Helpers end in `Helper` and call `t.Helper()` first.

## Domain vocabulary

The domain language is **English**. FoundryVTT terms (release, build, session, admin key,
license key, data path, install root, runtime) keep their canonical English names as used in
the official FoundryVTT documentation and API. Do not coin abbreviations or synonyms.

## See also

- [`imports.md`](imports.md) — import ordering and grouping
- [`types.md`](types.md) — type definitions and aliases