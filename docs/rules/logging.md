# Logging

**`log/slog` only.** No `fmt.Println`, no `log.Printf`, no third‑party loggers. The `forbidigo`
rule blocks `fmt.Print*` outside `*_test.go`, and `sloglint` enforces the call shape below.

## Call shape

- Use the package-level convenience funcs (`slog.Info`, `slog.Error`, ...) only at the very top
  of `main()`.
- Inside libraries and command functions, take a `*slog.Logger` as a parameter or from a
  constructor argument.
- Always use **typed attributes**, never positional `Sprintf`:

```go
logger.Info("foundry installed",
    slog.String("version", info.Version),
    slog.String("install_root", cfg.Paths.InstallRoot),
)

if err != nil {
    logger.Error("activation failed",
        slog.String("error", err.Error()),
        slog.String("install_root", state.App.Paths.InstallRoot),
    )
}
```

`sloglint` (with `attr-only`) rejects `slog.Info("msg", "key", value)` — always use
`slog.String`, `slog.Int`, `slog.Any`, `slog.Group`, etc.

## Levels

| Level | Meaning |
|---|---|
| `Debug` | Verbose tracing, off by default in production. |
| `Info` | Normal operation, business‑level events (request received, message sent). |
| `Warn` | Unexpected but recoverable (retryable downstream failure). |
| `Error` | Action failed; operator attention may be needed. |

Reserve `Error` for things you would page on. Use `Warn` for transient failures that the system
already retries.

## Context

Pass the `*slog.Logger` through the call chain as a function parameter, not on
`context.Context`. The activation step sequence receives the logger at the top of `step.Run`
and passes it to each `Step.Apply`.

## Secrets

- Never log credentials: `FOUNDRY_PASSWORD`, `FOUNDRY_SESSION`, `FOUNDRY_ADMIN_KEY`,
  `FOUNDRY_LICENSE_KEY`. Log only presence/absence, not values.
- Path values and version strings are safe to log.

## Format

- `colorlog.New("foundryctl", colorlog.LevelFromEnv())` is the handler used in production. It
  writes to stderr with ANSI colour when the terminal is a TTY and plain text otherwise.
- Log message strings are short, lowercase, no trailing punctuation: `"foundry installed"`,
  not `"Foundry was successfully installed."`.
- Keys are `snake_case`, lowercase.

## Don't log + return

When a function returns an error, **do not log it** unless this is the top of the call chain. The
caller will log once, with the wrapped context — see [`errors.md`](errors.md).

## See also

- [`errors.md`](errors.md) — error handling and wrapping
- [`startup.md`](startup.md) — where logging is configured