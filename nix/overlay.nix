final: prev:

let
  version = "1.4.0";

  fetchRelease =
    file: hash:
    prev.fetchurl {
      url = "https://github.com/oven-sh/bun/releases/download/bun-v${version}/${file}";
      inherit hash;
    };

  sources = {
    "x86_64-linux" = fetchRelease
      "bun-linux-x64-baseline.zip"
      "sha256-GE+0WV8NQBohfPfHjBvEMLqDMU2reouUgFurv3+nCX8=";
    "aarch64-linux" = fetchRelease
      "bun-linux-aarch64.zip"
      "sha256-SxozLuhhmD65O8/m93D/+U4+MbLDiL2uo8jtNeWO7Q4=";
    "aarch64-darwin" = fetchRelease
      "bun-darwin-aarch64.zip"
      "sha256-xmnpf2Fk4cluBwF0jbmN+ndJKQjL2DlMdVcTSnNd44E=";
  };

  system = prev.stdenvNoCC.hostPlatform.system;
in
{
  bun = prev.bun.overrideAttrs (prevAttrs: {
    inherit version;
    src = sources.${system} or (throw "bun ${version}: unsupported system ${system}");
    passthru = prevAttrs.passthru // { inherit sources; };
  });
}
