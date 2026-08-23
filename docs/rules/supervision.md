# Supervision and the JS runtime

## The runtime is chosen once, per install

The image ships Bun plus Node 22 and Node 24 (`nix/runtime/runtimes.nix`, assembled into the
image by `nix/runtime/root.nix` and the `Containerfile`). Foundry v13 boots only on Node 22
and v14 only on Node 24, so `jsruntime.Resolve` takes the resolved Foundry major and
probes exactly one binary. There is no unversioned fallback: guessing a Node major
for an unparseable version is how a v14 install ends up refusing to boot on Node 22.

Consequences:

- `step.JSRuntime()` must run **after** `step.Install()`; the version decides the
  binary.
- `activate.PrepareProfile` re-runs it whenever the version changes, or a profile
  switch across the v13/v14 boundary would keep the wrong runtime.
- The **kind** (`bun` or `node`) is global, from `FOUNDRY_JS_RUNTIME`. It is not a
  per-profile setting; only the Node major adapts per profile.
- `FOUNDRY_JS_RUNTIME_PATH` short-circuits the whole lookup with an explicit path.

The pinned versions live in Nix and never reach the Go code, so the only honest place
to assert them is against the built image; CI does that after the build.

## When the supervisor restarts, and when it exits

`procloop` owns the child process and `backoff` schedules its restarts.

`backoff.Manager.OnFailure` takes the child's uptime: one that outlived
`HealthyUptime` counts as recovered and its history is cleared before the failure is
recorded. Without that reset the counter only grows, and a server that crashed nine
times over the life of its volume pays `MaxDelay` on every crash forever after.

The supervisor exits, letting whatever manages the container recreate it, only
when respawning cannot plausibly help:

- the child failed to *start* (a missing or non-executable runtime binary),
- `MaxConsecutiveFailures` was reached,
- the Kubernetes bypass is on, and CrashLoopBackOff owns the throttling.

Everything else respawns in process, including an unwritable backoff cache, so the
dashboard and the Discord bot stay reachable while the server is down.
