{ pkgs, repoSrc }:

# foundryctl is a Go module living at apps/foundryctl in a multi-module
# workspace. We point buildGoModule at the leaf module via `modRoot`
# and treat the whole repo as the source so the `replace` directives
# can resolve sibling libs/.
#
# vendorHash = null: uses the vendor/ directory created by the Docker
# go-vendor stage (GOWORK=off go mod vendor). Not committed to git.

pkgs.buildGoModule {
  pname = "foundryctl";
  version = "0.0.0";
  src = repoSrc;
  modRoot = "apps/foundryctl";
  vendorHash = null;
  subPackages = [ "." ];
  # Static, CGO-free binary so it can run on a distroless or scratch base.
  env = { CGO_ENABLED = "0"; GOWORK = "off"; };
  ldflags = [ "-s" "-w" ];
  doCheck = false;
  meta = with pkgs.lib; {
    description = "PID 1 controller for the foundryvtt-docker container";
    license = licenses.mit;
    mainProgram = "foundryctl";
  };
}
