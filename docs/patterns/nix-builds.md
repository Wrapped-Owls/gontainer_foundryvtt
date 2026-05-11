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

`nix/image.nix` wraps the Go binary into an OCI layer. The binary is the only layer content — no downloaded runtime resources are embedded here.

```nix
{ pkgs, gontainerBin }:
pkgs.dockerTools.buildLayeredImage {
  name = "gontainer_foundryvtt";
  tag = "latest";
  contents = [ gontainerBin ];
  config.Cmd = [ "/bin/gontainer" ];
}
```

Runtime resources (e.g. FoundryVTT) are downloaded at container startup by the Go application, never baked in.

## Dockerfile as a thin Nix wrapper

When CI or users lack a local Nix installation, a Dockerfile may wrap the Nix build:

```dockerfile
FROM nixos/nix AS builder
COPY . /src
WORKDIR /src
RUN nix build --extra-experimental-features "nix-command flakes" .#default

FROM scratch
COPY --from=builder /src/result/bin/gontainer /gontainer
ENTRYPOINT ["/gontainer"]
```

The Dockerfile must not re-implement build logic that already exists in `flake.nix`.

## Rules recap

- Nix is the single source of build truth; Dockerfiles are wrappers only.
- No build steps live in shell scripts or `Makefile` targets if a Nix derivation handles them.
- Output is always the same binary regardless of whether you build via `nix build` or `docker build`.

## See also

- [`../rules/build.md`](../rules/build.md) — build system rules.
- [`../rules/code-placement.md`](../rules/code-placement.md) — `nix/` directory placement.
