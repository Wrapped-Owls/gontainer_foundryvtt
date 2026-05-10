{ pkgs, repoSrc }:

# foundryctl is a Go module living at apps/foundryctl in a multi-module
# workspace. The whole repo is the source so the `replace` directives
# can resolve sibling libs/.
#
# We use an impure derivation (__impure = true) so the build itself
# runs `go mod download` over the network. Integrity is enforced by
# the committed go.sum (GOSUMDB=off; go verifies every module hash
# against go.sum, so a tampered proxy / network mitm cannot inject
# bad code).
#
# Why impure instead of buildGoModule's FOD vendor? buildGoModule
# requires a `vendorHash` that has to be regenerated and committed
# every time go.mod changes. We deliberately trade reproducibility
# (this derivation is not pushable to a binary cache) for the
# simplicity of "go.mod + go.sum are the only sources of truth".
#
# Requires the `impure-derivations` experimental feature, enabled in
# the flake's nixConfig.

pkgs.stdenv.mkDerivation {
  pname = "foundryctl";
  version = "0.0.0";
  src = repoSrc;

  __impure = true;

  nativeBuildInputs = with pkgs; [ go cacert gitMinimal ];

  env = {
    CGO_ENABLED = "0";
    GOWORK = "off";
    GOSUMDB = "off";
    GOTOOLCHAIN = "local";
  };

  buildPhase = ''
    runHook preBuild

    export GOCACHE=$TMPDIR/go-cache
    export GOPATH=$TMPDIR/go
    export HOME=$TMPDIR

    cd apps/foundryctl
    go build \
      -trimpath \
      -ldflags="-s -w -buildid=" \
      -o $TMPDIR/foundryctl \
      .

    runHook postBuild
  '';

  installPhase = ''
    runHook preInstall
    install -Dm755 $TMPDIR/foundryctl $out/bin/foundryctl
    runHook postInstall
  '';

  meta = with pkgs.lib; {
    description = "PID 1 controller for the foundryvtt-docker container";
    license = licenses.mit;
    mainProgram = "foundryctl";
  };
}
