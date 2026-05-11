# Build

## Build system

Nix is the canonical build system for this repository.

- Dockerfiles must not duplicate build logic already implemented in Nix.
- Containers may use `nixos/nix` images purely as Nix execution wrappers for environments without a local Nix installation.
- Reproducibility is a hard requirement: build outputs must be cacheable and deterministic.

## OCI/image rules

- Downloaded runtime resources must never be embedded in the final image layer.
- Build outputs must remain reproducible and cacheable.
- Version resolution logic belongs in runtime/application code, not in shell wrappers or Dockerfile `RUN` steps.

## Runtime version selection

The runtime selection flow is:

1. If an explicit version env var exists:
   - Resolve a compatible local installation.
   - Ignore the patch component when it is unspecified.
2. If the local install is missing:
   - Attempt remote download.
3. If a custom runtime URL is configured:
   - Use the custom URL.
4. If an admin key is present:
   - Use authenticated download.
5. Otherwise:
   - Fall back to the latest installed runtime.

This logic lives in Go application code. It must not be re-implemented in shell or Dockerfile `RUN` steps.

## See also

- [`../patterns/nix-builds.md`](../patterns/nix-builds.md) — Nix flake and image build recipe.
- [`code-placement.md`](code-placement.md) — `nix/` directory placement rules.
