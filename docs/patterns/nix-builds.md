# Nix build pattern

Nix is the canonical build system: it resolves every package that ends up in the
image. The `Containerfile` assembles them, so a build is one ordinary docker command
and needs no local Nix installation.

## Flake structure

```
flake.nix               # top-level: inputs + outputs
flake.lock              # pins nixpkgs; without it the build is not reproducible
nix/
├── default.nix         # wires the packages, apps and devShell; no derivations of its own
├── overlay.nix         # nixpkgs overrides (Bun, until nixpkgs catches up)
├── packages/           # what this repo builds: the Go apps and the vendorHash updater
└── runtime/            # what the image ships
    ├── runtimes.nix    # the JS runtimes, named once for root.nix and deps.nix
    ├── root.nix        # the runtime filesystem root the image is built from
    └── deps.nix        # the closure the Containerfile caches before copying source
```

`default.nix` only imports and wires; every derivation lives in the file named after
what it produces. A version and the hashes that go with it stay in the same file, so a
bump is one edit.

## Building the Go binaries

`nix/packages/foundryctl.nix` and `nix/packages/taverncord.nix` are `buildGoModule`
derivations over the whole workspace. They vendor with `go work vendor` (plain
`go mod vendor` fails when `go.work` is present) and pin a `vendorHash`. After any
`go.mod` / `go.sum` change, regenerate it:

```sh
make nix-hash
```

## Building the container image

One `Containerfile` builds both images. The `builder` stage — flake, closure prefetch,
source — is shared, so whichever image builds second reuses it:

```sh
docker build -f Containerfile .                     # the Foundry runtime (default)
docker build -f Containerfile --target taverncord . # the Discord bot
```

`nix/runtime/root.nix` builds the runtime root: `bin/` symlinks plus the directories the
controller expects. The `Containerfile` runs that derivation in a `nixos/nix` builder
stage, copies the closure and the root onto `scratch`, and sets the OCI config.

Both splits in the `Containerfile` exist to keep rebuilds cheap, and both were
measured:

- The builder builds `.#runtimeDeps` from `flake.nix`/`flake.lock`/`nix/` alone before
  copying the source, so a Go change reuses that layer instead of re-fetching the
  whole closure from `cache.nixos.org`.
- The final image is three layers: the shared runtime closure (~394 MiB, moves only
  on a dependency bump), `foundryctl` (~9 MiB, moves on every Go change), and the
  symlink tree. Copying the whole store in one layer made a one-line Go fix
  re-transfer 402 MiB.

Stripping the Node binaries into leaner derivations was tried and rejected: the copy
keeps a store reference to its source, so the image carried both the original and the
stripped binary and grew by 118 MiB.

```sh
docker build -f Containerfile -t foundryvtt-docker:dev .
make image   # the same command
```

The image carries `foundryctl`, Bun, Node 22 and Node 24. Every nodejs derivation
installs its binary as `bin/node`, so the two majors are exposed as `node22` and
`node24` through the symlink tree in `root.nix`; merging the two packages directly
would let one shadow the other. `jsruntime.Resolve` probes those names by Foundry
major.

There is no shell in the image. That is deliberate (nothing but the runtimes and the
controller ships), but it means `docker exec ... sh` does not work; use
`--entrypoint` to run a specific binary.

## Pinning ahead of nixpkgs

`nix/overlay.nix` overrides the official nixpkgs derivation with the official upstream
release artifacts, for versions that have not landed in nixpkgs yet. Keep the override
to `version`, `src` and `passthru.sources`, keep the URLs pointing at the project's own
releases, and drop the override once nixpkgs ships the version. CI reads the pinned
version straight out of that file and fails if the built image disagrees.

## Rules recap

- Nix is the source of truth for which packages exist and at which versions.
- The Containerfile only assembles and configures; it resolves nothing itself.
- No build steps live in shell scripts.
- Runtime resources (FoundryVTT itself) are downloaded at container startup, never
  baked into a layer.

## See also

- [`../rules/build.md`](../rules/build.md) - build system rules.
- [`../rules/code-placement.md`](../rules/code-placement.md) - `nix/` directory placement.
