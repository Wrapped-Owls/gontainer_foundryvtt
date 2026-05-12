{ pkgs }:

# Pinned oven/bun:1-debian base image.
# Run `nix build .#image` with fakeHash to discover the correct sha256,
# then update it here.
(pkgs.dockerTools.pullImage {
  imageName    = "oven/bun";
  imageDigest  = "sha256:e95356cb8e1de62ad69ab3bd3584ba947013d27650a226804d2fc0af4e17dac2";
  sha256       = "sha256-0rDNv0/o+vlVlKCjR6uA4XyrncHOG2as0lRuuKySqec=";
  finalImageTag = "1-debian";
}).overrideAttrs (old: {
  # skopeo looks for an auth file even for public registries; on CI runners
  # the system path (/run/containers/<uid>/auth.json) exists but is not
  # readable by the Nix build user.  Point it at an empty temp file instead.
  buildCommand = ''
    export REGISTRY_AUTH_FILE="$TMPDIR/auth.json"
    echo '{}' > "$REGISTRY_AUTH_FILE"
  '' + old.buildCommand;
})
