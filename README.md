# foundryvtt-docker

Container runtime for Foundry Virtual Tabletop.

## Layout

```text
apps/foundryctl/      PID 1 controller and startup logic
libs/fourcery/        unified install acquisition + sourcing (strategy/builder/factory/observer)
libs/foundrypatch/    version-gated install patching with idempotent ledger
libs/foundryruntime/  runtime config, health, js runtime, lifecycle
examples/             runnable examples
nix/                  image and package module definitions
flake.nix             Nix flake (packages: foundryctl, bun, image)
patches/manifest.yaml install patches applied before launch
tests/                integration and e2e modules
```

## Build

```sh
docker build -t foundryvtt-docker:dev .
```

Or with Nix:

```sh
nix build .#image
```

## Run

`FOUNDRY_INSTALL_ROOT` (default `/foundry`) is the install store. The controller:

1. checks `FOUNDRY_VERSION` and tries to select an already-installed version
2. checks `FOUNDRY_SOURCES_DIR` for a matching zip or pre-extracted folder
3. falls back to a network acquisition source if configured
4. otherwise selects the newest installed version in the store

Acquisition sources (checked in this order after local artefacts):

- `FOUNDRY_SESSION` — authenticated session cookie
- `FOUNDRY_USERNAME` + `FOUNDRY_PASSWORD` — login credentials
- `FOUNDRY_RELEASE_URL` — presigned download URL (used only when no local artefact matches)

### Local zip / folder install

Drop a zip or pre-extracted folder into the sources volume. The filename must encode the version:
`foundryvtt_v14.361.2.zip` or `foundryvtt_v14.361.2/`. No credentials required.

```sh
mkdir -p resources/foundry/sources
cp ~/Downloads/foundryvtt_v14.361.2.zip resources/foundry/sources/
```

### Example run command

```sh
docker run --name foundry \
  -p 30000:30000 \
  -v "$PWD/resources/foundry/data:/data" \
  -v "$PWD/resources/foundry/installation:/foundry" \
  -v "$PWD/resources/foundry/sources:/foundry/sources" \
  -v "$PWD/patches/manifest.yaml:/etc/foundry/patches/manifest.yaml:ro" \
  -e FOUNDRY_ADMIN_KEY=changeme \
  -e FOUNDRY_VERSION=14.361.2 \
  -e FOUNDRY_JS_RUNTIME=bun \
  foundryvtt-docker:dev
```

Restarting the container with the same `/foundry` volume reuses the installed version and skips
already-applied patches — no re-download, no re-patch.

## Development

```sh
make vet test
make fmt
make tidy
```

## License

MIT for this repo. FoundryVTT itself is proprietary and not distributed by this source tree.
