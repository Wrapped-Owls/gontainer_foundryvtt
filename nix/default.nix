{ pkgs, src }:

let
  foundryctl = import ./modules/foundryctl.nix { inherit pkgs; repoSrc = src; };
  bun = import ./bun.nix { inherit pkgs; };
  image = import ./image.nix { inherit pkgs foundryctl bun; };
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
    inherit foundryctl bun image;
    default = foundryctl;
  };

  apps = {
    foundryctl = { type = "app"; program = "${foundryctl}/bin/foundryctl"; };
  };

  formatter = pkgs.nixpkgs-fmt;
}
