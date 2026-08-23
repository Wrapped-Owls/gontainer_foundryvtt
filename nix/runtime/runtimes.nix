{ pkgs }:

{
  inherit (pkgs) bun;
  node22 = pkgs.nodejs-slim_22;
  node24 = pkgs.nodejs-slim_24;
}
