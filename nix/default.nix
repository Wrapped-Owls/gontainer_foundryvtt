{ pkgs, src }:

let
  foundryctl       = import ./packages/foundryctl.nix         { inherit pkgs; repoSrc = src; };
  taverncord       = import ./packages/taverncord.nix         { inherit pkgs; repoSrc = src; };
  updateVendorHash = import ./packages/update-vendor-hash.nix { inherit pkgs; };
  runtimes         = import ./runtime/runtimes.nix { inherit pkgs; };
  runtimeRoot      = import ./runtime/root.nix     { inherit pkgs foundryctl runtimes; repoSrc = src; };
  runtimeDeps      = import ./runtime/deps.nix     { inherit pkgs runtimes; };
in {
  devShells.default = pkgs.mkShell {
    packages = with pkgs; [
      go
      gopls
      gotools
      golangci-lint
      git
      gnumake
      bun
      nodejs-slim_22
    ];
  };

  packages = {
    inherit foundryctl taverncord runtimeRoot runtimeDeps;
    default = foundryctl;
  };

  apps = {
    foundryctl         = { type = "app"; program = "${foundryctl}/bin/foundryctl"; };
    taverncord         = { type = "app"; program = "${taverncord}/bin/taverncord"; };
    update-vendor-hash = { type = "app"; program = "${updateVendorHash}/bin/update-vendor-hash"; };
  };

  formatter = pkgs.nixpkgs-fmt;
}
