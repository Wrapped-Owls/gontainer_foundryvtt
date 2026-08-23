# `procspawn` - subprocess configuration

How to configure and launch a child process using `libs/foundrykit/procspawn`.

## Spec

A `procspawn.Spec` describes the process to launch:

```go
spec := procspawn.Spec{
    Path: state.JSRuntime.Path,   // absolute path to bun or node
    Args: []string{
        mainScript,
        "--dataPath=" + dataPath,
        "--port=" + strconv.Itoa(port),
    },
    Dir: state.Install.Root,      // working directory for the child
}
```

`procspawn.Run(ctx, spec)` forks the process, wires its stdin/stdout/stderr to the parent, and
returns when the child exits. It forwards `SIGTERM`/`SIGINT` to the child on context
cancellation.

## Bun vs Node argument shape

The FoundryVTT `main.mjs` entry point requires a `run` prefix when launched via `bun`.
`procloop.BuildArgs` owns that difference; it also appends `--world=` when the profile
names a world to boot straight into.

```go
args := []string{mainScript, "--dataPath=" + dataPath, "--port=" + strconv.Itoa(port)}
if kind == jsruntime.Bun {
    return append([]string{"run"}, args...)
}
return args
```

## Which binary

`jsruntime.Resolve` decides; the Foundry-major rule lives in
[`../rules/supervision.md`](../rules/supervision.md). `FOUNDRY_JS_RUNTIME_PATH`
short-circuits the whole lookup.

The `JSRuntime` step therefore has to run AFTER `Install`, and has to re-run whenever
the resolved version changes - `activate.PrepareProfile` re-runs it in the
version-changed branch for exactly that reason.

## Backoff loop

`cmd/run.go` wraps `procspawn.Run` in a restart loop managed by `libs/foundrykit/backoff`:

```go
mgr := backoff.NewFromConfig(state.App.Backoff)
for {
    code, err := procspawn.Run(ctx, spec)
    if err != nil {
        logger.Error("child failed to start", "err", err)
        return 1
    }
    if errors.Is(ctx.Err(), context.Canceled) {
        logger.Info("shutdown requested; exiting", "exit_code", code)
        return code
    }
    dec, err := mgr.OnFailure(code)
    if err != nil { return code }
    switch dec.Mode {
    case backoff.ModeKubernetes:
        return code
    case backoff.ModeBackoff:
        if dec.Delay == 0 { return code }
        _ = backoff.Sleep(ctx, dec.Delay)
    }
}
```

## Forbidden

- Calling `os/exec.Command` directly - use `procspawn.Run`.
- Hard-coding the `bun` or `node` binary path - detect it via the `JSRuntime` step.
- Mutating `Spec` fields inside the restart loop.

## See also

- [`../rules/concurrency.md`](../rules/concurrency.md) - long-running process loop rules.
- [`bootstrap-and-di.md`](bootstrap-and-di.md) - how the backoff loop fits in `cmd.Run`.
- [`usecase-layout.md`](usecase-layout.md) - `JSRuntime` step that populates `state.JSRuntime`.
