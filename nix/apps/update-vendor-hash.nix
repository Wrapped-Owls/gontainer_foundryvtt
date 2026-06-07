{ pkgs }:

# Recomputes the shared vendorHash used by all Go workspace apps (foundryctl,
# taverncord, …). Both modules run `go work vendor` on the same workspace so
# they always produce an identical vendor directory — one hash covers all.
# Used locally via `nix run .#update-vendor-hash` and invoked by the CI
# workflow after go.mod / go.sum changes.
pkgs.writeShellApplication {
  name = "update-vendor-hash";

  runtimeInputs = with pkgs; [
    nix
    git
    gnused
    gawk
  ];

  text = ''
    root=$(git rev-parse --show-toplevel)

    # All nix modules that embed the workspace vendorHash.
    nix_files=(
      "$root/nix/modules/foundryctl.nix"
      "$root/nix/modules/taverncord.nix"
    )

    echo "==> Resetting vendorHash to fakeHash in all modules..."
    for f in "''${nix_files[@]}"; do
      sed -i 's|vendorHash = .*|vendorHash = pkgs.lib.fakeHash;|' "$f"
    done

    echo "==> Building foundryctl to discover correct hash (expected to fail)..."
    build_output=$(nix build .#foundryctl 2>&1 || true)
    hash=$(printf '%s\n' "$build_output" | awk '/got:/ { print $2; exit }')

    if [[ -z "$hash" ]]; then
      echo "ERROR: could not extract hash from Nix output" >&2
      printf '%s\n' "$build_output" >&2
      exit 1
    fi

    echo "==> Patching vendorHash = \"$hash\" in all modules..."
    for f in "''${nix_files[@]}"; do
      sed -i "s|vendorHash = .*|vendorHash = \"$hash\";|" "$f"
    done

    echo "Done."
  '';
}
