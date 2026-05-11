# `patterns/`

Implementation recipes for `gontainer_foundryvtt`. Each file shows **how** to apply the rules
with working, self-contained Go code.

## Index

| Pattern | What it solves |
|---|---|
| [`bootstrap-and-di.md`](bootstrap-and-di.md) | Activation sequence from `main()` to Foundry launch |
| [`step-config.md`](step-config.md) | Config flowing from `AppConfig` step into the pipeline |
| [`usecase-layout.md`](usecase-layout.md) | Step factory pattern and sub-package structure |
| [`patch-manifest.md`](patch-manifest.md) | Patch manifest format and action types |
| [`jsonhttp.md`](jsonhttp.md) | `jsonhttp.Request` typed HTTP call pattern |
| [`integration-tests.md`](integration-tests.md) | Integration tests with `t.TempDir()` and `t.Setenv` |
| [`auth-session.md`](auth-session.md) | `foundryacquire` auth session and cookie reuse |
| [`app-skeleton.md`](app-skeleton.md) | New `apps/<name>/` skeleton |
| [`confloader-layout.md`](confloader-layout.md) | Config package layout with `confloader` |
| [`nix-builds.md`](nix-builds.md) | Nix flake, Go binary, and container image build |
| [`procspawn.md`](procspawn.md) | `procspawn` spec and backoff restart loop |
