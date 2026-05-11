{ pkgs }:

# Recomputes the vendorHash in nix/modules/foundryctl.nix by temporarily
# setting it to fakeHash, letting Nix report the real hash, then patching
# the file. Used locally via `nix run .#update-vendor-hash` and invoked by
# the update-vendor-hash CI workflow after go.mod / go.sum changes.
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
    nix_file="$root/nix/modules/foundryctl.nix"

    echo "==> Resetting vendorHash to fakeHash..."
    sed -i 's|vendorHash = .*|vendorHash = pkgs.lib.fakeHash;|' "$nix_file"

    echo "==> Building to discover correct hash (expected to fail)..."
    build_output=$(nix build .#foundryctl 2>&1 || true)
    hash=$(printf '%s\n' "$build_output" | awk '/got:/ { print $2; exit }')

    if [[ -z "$hash" ]]; then
      echo "ERROR: could not extract hash from Nix output" >&2
      printf '%s\n' "$build_output" >&2
      exit 1
    fi

    echo "==> Patching vendorHash = \"$hash\"..."
    sed -i "s|vendorHash = .*|vendorHash = \"$hash\";|" "$nix_file"

    echo "Done."
  '';
}
