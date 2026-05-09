{ pkgs }:

# Bun is available in nixpkgs; expose it as a stand-alone derivation so
# image.nix can reference exactly one Bun across all variants and so the
# dev shell stays consistent with the runtime.
pkgs.bun
