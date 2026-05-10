# foundryvtt-docker

Container runtime for Foundry Virtual Tabletop.

## Layout

```text
apps/foundryctl/      PID 1 controller and startup logic
libs/foundryacquire/  Foundry auth and release URL fetch
libs/foundrypatch/    version-gated install patching
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

`FOUNDRY_INSTALL_ROOT` is treated as an install store. The controller:

1. checks `FOUNDRY_VERSION` and tries to select an installed version from the store
2. downloads that version when it is missing and an acquisition source is configured
3. otherwise selects the newest installed version in the store

Acquisition sources:

- `FOUNDRY_RELEASE_URL`
- `FOUNDRY_SESSION`
- `FOUNDRY_USERNAME` + `FOUNDRY_PASSWORD`

Example:

```sh
docker run --name foundry \
  -p 30000:30000 \
  -v "$PWD/resources/foundry/data:/data" \
  -v "$PWD/resources/foundry/installation:/foundry" \
  -v "$PWD/patches/manifest.yaml:/etc/foundry/patches/manifest.yaml:ro" \
  -e FOUNDRY_ADMIN_KEY=changeme \
  -e FOUNDRY_VERSION=14.361 \
  -e FOUNDRY_JS_RUNTIME=bun \
  -e FOUNDRY_RELEASE_URL='https://r2.foundryvtt.com/...signed...' \
  foundryvtt-docker:dev
```

## Development

```sh
make vet test
make fmt
make tidy
```

## License

MIT for this repo. FoundryVTT itself is proprietary and not distributed by this source tree.
