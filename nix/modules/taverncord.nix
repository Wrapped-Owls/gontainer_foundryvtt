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

  vendorHash = "sha256-+Jp6wWnET6O3w/GG56CNoDc4TlrYXNTo60vs5puPRUc=";

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
