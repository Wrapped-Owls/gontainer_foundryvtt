{ pkgs, src }:

let
  foundryctl       = import ./modules/foundryctl.nix { inherit pkgs; repoSrc = src; };
  taverncord       = import ./modules/taverncord.nix  { inherit pkgs; repoSrc = src; };
  updateVendorHash = import ./apps/update-vendor-hash.nix { inherit pkgs; };
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
    ];
  };

  packages = {
    inherit foundryctl taverncord;
    default = foundryctl;
  };

  apps = {
    foundryctl         = { type = "app"; program = "${foundryctl}/bin/foundryctl"; };
    taverncord         = { type = "app"; program = "${taverncord}/bin/taverncord"; };
    update-vendor-hash = { type = "app"; program = "${updateVendorHash}/bin/update-vendor-hash"; };
  };

  formatter = pkgs.nixpkgs-fmt;
}
