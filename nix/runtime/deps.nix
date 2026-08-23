{ pkgs, runtimes }:

pkgs.runCommand "foundry-runtime-deps" { } ''
  mkdir -p $out
  ln -s ${runtimes.bun}    $out/bun
  ln -s ${runtimes.node22} $out/node22
  ln -s ${runtimes.node24} $out/node24
  ln -s ${pkgs.cacert}     $out/cacert
  ln -s ${pkgs.tzdata}     $out/tzdata
''
