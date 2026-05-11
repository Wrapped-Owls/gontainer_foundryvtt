# Patch manifest format

How to add entries to `patches/manifest.yaml`, the runtime patch manifest consumed by the
`foundrypatch` activation step.

## File location

```
patches/
└── manifest.yaml   # top-level repo; mounted into the container at FOUNDRY_PATCH_MANIFEST
```

The default mount path is `/etc/foundry/patches/manifest.yaml`, overridable via
`FOUNDRY_PATCH_MANIFEST`.

## Manifest structure

```yaml
version: 1
patches:
  - versions: ">=12.0.0 <13.0.0"   # semver range (github.com/Masterminds/semver/v3)
    actions:
      - type: filereplace
        target: resources/app/public/css/style.css
        source: ./patches/v12/style.css
      - type: download
        url: https://example.com/asset.zip
        target: resources/app/public/asset.zip
      - type: zipoverlay
        url: https://example.com/overlay.zip
        target: resources/app/public/
```

## Action types

| `type` | Required fields | Effect |
|---|---|---|
| `filereplace` | `target`, `source` | Copies `source` (relative to manifest dir) to `target` (relative to `FOUNDRY_INSTALL_ROOT`) |
| `download` | `url`, `target` | Downloads `url` and writes to `target` (relative to `FOUNDRY_INSTALL_ROOT`) |
| `zipoverlay` | `url`, `target` | Downloads `url` (a zip), extracts contents over `target` directory |

## Version ranges

Version ranges follow the semver syntax from `github.com/Masterminds/semver/v3`. Use `>=` / `<`
for exclusive upper bounds. A missing `versions` field applies to all Foundry builds.

## Testing patches

1. Add a fixture `manifest.yaml` to a `testdata/` directory.
2. Call `manifest.Load` and `manifest.Filter` with the fixture path and a test version.
3. Call `applier.Applier.Apply` with a `t.TempDir()` as the install root.
4. Assert the target files are created/replaced.

## See also

- [`../rules/patches.md`](../rules/patches.md) — `foundrypatch` package structure.
- [`integration-tests.md`](integration-tests.md) — testing applier with fixtures.
