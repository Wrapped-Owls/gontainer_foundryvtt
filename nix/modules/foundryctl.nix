{ pkgs, repoSrc }:

# foundryctl lives in apps/foundryctl inside a multi-module workspace.
# We use buildGoModule (pure, cacheable) instead of __impure = true.
#
# The FOD (fixed-output derivation) phase runs `go work vendor` from the
# workspace root so all five modules are vendored together. The main build
# then compiles only apps/foundryctl with -mod=vendor (set automatically
# by buildGoModule when vendorHash is non-null).
#
# After any go.mod / go.sum change, regenerate vendorHash:
#   nix build .#foundryctl 2>&1 | grep 'got:' | awk '{print $2}'
# then paste the printed hash below.

pkgs.buildGoModule {
  pname = "foundryctl";
  version = "0.0.0";
  src = repoSrc;

  vendorHash = "sha256-UNmsR9O6M7C1ySfo6BaDf7/kI4bangMrl1+iEAgyUSE=";

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
