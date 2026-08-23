{ pkgs, repoSrc }:

# Vendors the whole go.work workspace; `make nix-hash` regenerates vendorHash.

pkgs.buildGoModule {
  pname = "foundryctl";
  version = "0.0.0";
  src = repoSrc;

  vendorHash = "sha256-wzd/RDhOf+dr/R9hNakva6ILao+fzEeka8rhd193oI4=";

  overrideModAttrs = _: {
    # go mod vendor fails when go.work is present (nixpkgs #347998).
    # go work vendor creates vendor/ at the workspace root and works
    # correctly with workspace replace directives.
    buildPhase = ''
      runHook preBuild
      export HOME=$TMPDIR
      go work vendor
      runHook postBuild
    '';
    installPhase = ''
      runHook preInstall
      cp -r vendor "$out"
      runHook postInstall
    '';
  };

  subPackages = [ "apps/foundryctl" ];

  env = {
    CGO_ENABLED = "0";
    GOTOOLCHAIN = "local";
  };

  ldflags = [ "-s" "-w" "-buildid=" ];
  buildFlags = [ "-trimpath" ];

  meta = with pkgs.lib; {
    description = "PID 1 controller for the foundryvtt-docker container";
    license = licenses.mit;
    mainProgram = "foundryctl";
  };
}
