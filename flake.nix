{
  description = "foundryvtt-docker runtime and OCI image";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils, ... }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs { inherit system; };
        repoSrc = ./.;
        foundryctl = import ./nix/modules/foundryctl.nix { inherit pkgs repoSrc; };
        bun = import ./nix/bun.nix { inherit pkgs; };
        image = import ./nix/image.nix {
          inherit pkgs foundryctl bun;
        };
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
      });
}
