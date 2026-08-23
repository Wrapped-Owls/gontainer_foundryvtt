{ pkgs, foundryctl, runtimes, repoSrc }:

pkgs.runCommand "foundry-runtime-root" { } ''
  mkdir -p $out/bin $out/data $out/foundry/sources $out/etc/foundry/patches $out/etc/ssl/certs $out/tmp

  ln -s ${foundryctl}/bin/foundryctl  $out/bin/foundryctl
  ln -s ${runtimes.bun}/bin/bun       $out/bin/bun
  ln -s ${runtimes.node22}/bin/node   $out/bin/node22
  ln -s ${runtimes.node24}/bin/node   $out/bin/node24

  cp ${pkgs.dockerTools.fakeNss}/etc/passwd  $out/etc/passwd
  cp ${pkgs.dockerTools.fakeNss}/etc/group   $out/etc/group

  ln -s ${pkgs.cacert}/etc/ssl/certs/ca-bundle.crt $out/etc/ssl/certs/ca-bundle.crt
  ln -s ${pkgs.tzdata}/share/zoneinfo              $out/etc/zoneinfo

  cp ${repoSrc}/patches/manifest.yaml $out/etc/foundry/patches/manifest.yaml
  touch $out/etc/foundry/profiles.json

  chmod -R u+w $out/data $out/foundry $out/tmp $out/etc/foundry
''
