{ pkgs, repoSrc }:

# taverncord lives in apps/taverncord inside a multi-module workspace.
# Mirrors foundryctl.nix: uses go work vendor at the workspace root so all
# modules are vendored together, then builds only apps/taverncord.
#
# vendorHash must equal foundryctl's — both run go work vendor on the same
# workspace, producing an identical vendor directory.
# Regenerate with: nix run .#update-vendor-hash

pkgs.buildGoModule {
  pname = "taverncord";
  version = "0.0.0";
  src = repoSrc;

  vendorHash = "sha256-wzd/RDhOf+dr/R9hNakva6ILao+fzEeka8rhd193oI4=";

  overrideModAttrs = _: {
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

  subPackages = [ "apps/taverncord" ];

  env = {
    CGO_ENABLED = "0";
    GOTOOLCHAIN = "local";
  };

  ldflags = [ "-s" "-w" "-buildid=" ];
  buildFlags = [ "-trimpath" ];

  meta = with pkgs.lib; {
    description = "Discord bot for FoundryVTT profile switching";
    license = licenses.mit;
    mainProgram = "taverncord";
  };
}
