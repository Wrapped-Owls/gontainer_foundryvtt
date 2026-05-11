# Errors

**Errors are values.** Functions return `error` as their last result; `panic` is reserved for
unrecoverable boot‑time failures inside `cmd/<app>/main.go`.

## The wrap rule — `wrapcheck`

Whenever you forward an error across a package boundary, wrap it with context:

- Use `%w`, never `%v` or `%s`, so `errors.Is` / `errors.As` keep working.
- Lead with the action you tried, not "error while ...". The wrapper supplies the trailing
  `: <cause>` when printed.
- The `wrapcheck` linter rejects naked returns of errors that originated outside the current
  module/package.

## Sentinel errors

Define sentinel errors as exported package vars when callers need to branch on them:

- Use `errors.Is` for equality, `errors.As` for typed unwrapping.
- Don't compare on `err.Error()` strings — ever.

## Typed sentinel errors

For errors that callers need to branch on (e.g. "version not found", "no local install"),
define exported sentinel vars in the package that owns the concept:

```go
var (
    ErrNoBuildNumber = errors.New("release: no build number in version string")
    ErrEmptyURL      = errors.New("release: server returned an empty URL")
)
```

Return them directly or wrapped: `fmt.Errorf("fetch: %w", ErrNoBuildNumber)`. Callers test
with `errors.Is`.

## Forbidden

- Returning `nil` together with a non‑nil result and **no** error to signal "soft failure". Pick
  one: either `(zero, error)` or use a typed result.
- `panic` in libraries.
- `recover()` to mask bugs. The single legitimate use is at the top of a supervised long‑running
  goroutine (see [`concurrency.md`](concurrency.md)).
- Logging an error then returning the same error — log **or** return, not both. The top of the
  call chain logs.

## Boundaries — `recover`

`recover` is allowed at exactly one boundary: a goroutine supervisor that restarts the process
loop (e.g. `cmd/run.go`'s backoff loop). Anywhere else, an unwound panic is a bug.

## Logging an error

Once, at the top of the call chain, with the structured pair `slog.String("error", err.Error())`
or `slog.Any("error", err)`. See [`logging.md`](logging.md).

## See also

- [`logging.md`](logging.md) — structured logging
- [`concurrency.md`](concurrency.md) — goroutine supervision and recovery