{ pkgs, foundryctl, bunImage }:

let
  rootfs = pkgs.buildEnv {
    name = "foundryvtt-docker-rootfs";
    paths = with pkgs; [
      cacert
      tzdata
      foundryctl
    ];
  };

  patchManifest = pkgs.runCommand "foundryvtt-patch-manifest" { } ''
    mkdir -p $out/etc/foundry/patches
    cp ${../patches/manifest.yaml} $out/etc/foundry/patches/manifest.yaml
  '';

  runtimeDirs = pkgs.runCommand "foundryvtt-runtime-dirs" { } ''
    mkdir -p $out/data $out/foundry $out/foundry/sources
  '';
in
pkgs.dockerTools.buildLayeredImage {
  name      = "foundryvtt-docker";
  tag       = "latest";
  fromImage = bunImage;
  contents  = [ rootfs patchManifest runtimeDirs ];

  config = {
    Entrypoint = [ "${foundryctl}/bin/foundryctl" ];
    Cmd        = [ "run" ];
    WorkingDir = "/";
    Env = [
      "PATH=/bin:/usr/bin:/usr/local/bin"
      "SSL_CERT_FILE=/etc/ssl/certs/ca-bundle.crt"
      "FOUNDRY_DATA_PATH=/data"
      "FOUNDRY_INSTALL_ROOT=/foundry"
      "FOUNDRY_SOURCES_DIR=/foundry/sources"
      "FOUNDRY_PATCH_MANIFEST=/etc/foundry/patches/manifest.yaml"
    ];
    ExposedPorts = { "30000/tcp" = {}; };
    Volumes = { "/data" = {}; };
    Healthcheck = {
      Test = [ "CMD" "${foundryctl}/bin/foundryctl" "healthcheck" ];
      Interval = 30000000000;
      Timeout  =  5000000000;
      Retries  = 3;
      StartPeriod = 180000000000;
    };
  };
}
