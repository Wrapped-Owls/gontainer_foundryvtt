{ pkgs, src }:

let
  foundryctl       = import ./modules/foundryctl.nix { inherit pkgs; repoSrc = src; };
  bunImage         = import ./bun-image.nix { inherit pkgs; };
  image            = import ./image.nix { inherit pkgs foundryctl bunImage; };
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
    inherit foundryctl image;
    default = foundryctl;
  };

  apps = {
    foundryctl         = { type = "app"; program = "${foundryctl}/bin/foundryctl"; };
    update-vendor-hash = { type = "app"; program = "${updateVendorHash}/bin/update-vendor-hash"; };
  };

  formatter = pkgs.nixpkgs-fmt;
}
