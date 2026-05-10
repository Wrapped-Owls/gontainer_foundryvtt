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
in
pkgs.dockerTools.buildLayeredImage {
  name = "foundryvtt-docker";
  tag  = "latest";
  contents = [ rootfs ];

  fakeRootCommands = ''
    mkdir -p data foundry etc/foundry/patches
    cp ${../patches/manifest.yaml} etc/foundry/patches/manifest.yaml
  '';
  enableFakechroot = true;

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
