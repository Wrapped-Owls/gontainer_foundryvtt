# Nix build pattern

Nix is the canonical build system. This pattern describes how to wire a Go binary and a container image.

## Flake structure

```
flake.nix               # top-level: inputs + outputs
nix/
├── image.nix           # OCI/Docker image derivation
└── bun.nix             # optional: JS/TS tooling overlay
```

## Building the Go binary

In `flake.nix`, expose the binary as a package:

```nix
packages.default = pkgs.buildGoModule {
  pname = "gontainer";
  version = "0.1.0";
  src = ./.;
  vendorHash = "sha256-...";
};
```

Use `nix build` to produce a reproducible binary in `result/`.

## Building the container image

The container image is built via `Containerfile` using Docker Buildx — this is the canonical path for both CI and local builds:

```sh
docker build -f Containerfile -t foundryvtt-docker:dev .
```

The `Containerfile` uses a `nixos/nix` builder stage to compile `foundryctl` via `nix build .#foundryctl`, then copies only the binary into a minimal `oven/bun:1-debian` runtime image. Runtime resources (e.g. FoundryVTT) are downloaded at container startup, never baked in.

## Rules recap

- Nix is the source of truth for the Go binary (`nix build .#foundryctl`); the Containerfile assembles the final image.
- No build steps live in shell scripts or `Makefile` targets if a Nix derivation handles them.
- Output is always the same binary regardless of whether you build via `nix build .#foundryctl` or via the Containerfile builder stage.

## See also

- [`../rules/build.md`](../rules/build.md) — build system rules.
- [`../rules/code-placement.md`](../rules/code-placement.md) — `nix/` directory placement.
