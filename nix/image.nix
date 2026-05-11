{ pkgs, foundryctl, bun }:

let
  rootfs = pkgs.buildEnv {
    name = "foundryvtt-docker-rootfs";
    paths = with pkgs; [
      cacert
      tzdata
      bun
      foundryctl
      busybox
    ];
  };

  # Separate derivations so buildLayeredImage never needs to write into
  # existing store-owned directories (which are read-only and cause
  # permission errors with both fakeRootCommands and extraCommands).
  patchManifest = pkgs.runCommand "foundryvtt-patch-manifest" { } ''
    mkdir -p $out/etc/foundry/patches
    cp ${../patches/manifest.yaml} $out/etc/foundry/patches/manifest.yaml
  '';

  runtimeDirs = pkgs.runCommand "foundryvtt-runtime-dirs" { } ''
    mkdir -p $out/data $out/foundry
  '';
in
pkgs.dockerTools.buildLayeredImage {
  name = "foundryvtt-docker";
  tag  = "latest";
  contents = [ rootfs patchManifest runtimeDirs ];

  config = {
    Entrypoint = [ "${foundryctl}/bin/foundryctl" ];
    Cmd        = [ "run" ];
    WorkingDir = "/";
    Env = [
      "PATH=/bin:/usr/bin"
      "SSL_CERT_FILE=/etc/ssl/certs/ca-bundle.crt"
      "FOUNDRY_DATA_PATH=/data"
      "FOUNDRY_INSTALL_ROOT=/foundry"
      "FOUNDRY_PATCH_MANIFEST=/etc/foundry/patches/manifest.yaml"
    ];
    ExposedPorts = { "30000/tcp" = {}; };
    Volumes = { "/data" = {}; };
    Healthcheck = {
      Test = [ "CMD" "${foundryctl}/bin/foundryctl" "healthcheck" ];
      Interval = 30000000000;   # 30s in ns
      Timeout  =  5000000000;   #  5s
      Retries  = 3;
      StartPeriod = 180000000000; # 3m
    };
  };
}
