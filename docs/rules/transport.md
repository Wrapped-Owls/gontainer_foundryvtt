# Transport — CLI command dispatch and typed HTTP

`foundryctl` is a CLI binary. There is no HTTP server. This rule covers the command dispatch
pattern and the typed HTTP client rule for outbound API calls.

## Command dispatch

`main.go` routes the first positional argument to a function in `apps/foundryctl/cmd/`:

```go
switch sub {
case "run":
    os.Exit(cmd.Run(args, logger))
case "healthcheck":
    os.Exit(cmd.Healthcheck(args, logger))
case "options":
    os.Exit(cmd.Options(args, logger))
case "version":
    cmd.Version()
}
```

Each non-trivial command signature is `func(args []string, logger *slog.Logger) int`. The
return value is the process exit code. `cmd.Version` is the exception — it always exits 0 and
takes no logger.

## Anatomy of a command

```
cmd/<name>.go
  │
  ├─ Parse flags from args            ← cmd concern
  ├─ Validate surface-level inputs    ← cmd concern
  │
  ├─ Call activate.Prepare or         ← cross the boundary
  │   internal/... helpers
  │
  ├─ Run process or return result     ← cmd concern
  └─ Return exit code
```

Business logic (version selection, patch application, JS runtime detection) lives in `internal/`
packages and library modules. Command files contain only parsing, dispatch, and exit-code
translation — no business rules.

## Typed HTTP

HTTP interactions must remain strongly typed.

- Prefer `libs/foundrykit/jsonhttp.Request[Resp, Body]` over raw `http.Client` with manual
  decode.
- Avoid `map[string]any` payloads, untyped JSON handling, and dynamic response decoding.
- Request and response types must be named structs, never anonymous maps.

```go
result, err := jsonhttp.Request[releaseURLResp, struct{}](ctx,
    jsonhttp.ClientConfig{BaseURL: auth.BaseURL, HTTP: sess.Client()},
    jsonhttp.RequestConfig[struct{}]{Method: http.MethodGet, Path: path},
)
```

## File placement

```
apps/foundryctl/
├── main.go           # dispatch switch
└── cmd/
    ├── run.go        # foundryctl run
    ├── healthcheck.go
    ├── options.go
    ├── version.go
    └── signal.go     # OS signal wiring shared by run
```

## Forbidden

- ❌ Business rule logic in `cmd/` files — activation, version resolution, patch application
  belong in `internal/` or `libs/`.
- ❌ `os.Exit` called from `internal/` packages — only `main.go` and `cmd/` functions exit.
- ❌ Shared mutable state between command functions.
- ❌ Raw `http.Get`/`http.Post` — use `jsonhttp.Request`.
- ❌ `map[string]any` request or response bodies.

## See also

- [`startup.md`](startup.md) — full boot sequence.
- [`../rules/http-clients.md`](http-clients.md) — `fourcery` typed HTTP client.
- [`../patterns/app-skeleton.md`](../patterns/app-skeleton.md) — adding a new subcommand.
