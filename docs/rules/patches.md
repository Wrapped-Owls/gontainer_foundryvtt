# Patches — the `foundrypatch` manifest system

FoundryVTT releases occasionally need post-install patches: file replacements, asset overlays,
or supplemental downloads. `libs/foundrypatch` manages this through a typed YAML manifest.

## Layout

```
libs/foundrypatch/
├── go.mod
├── manifest/
│   ├── doc.go            # package doc
│   ├── types.go          # Manifest, Entry, Action types
│   ├── load.go           # YAML loading + validation
│   └── filter.go         # version-based entry filtering
└── applier/
    ├── doc.go
    ├── applier.go         # Applier struct + Apply method
    ├── applier_test.go
    ├── download_test.go
    ├── extract_test.go
    ├── safe_test.go
    └── action/
        ├── action.go      # Action interface
        ├── download.go    # DownloadAction
        ├── filereplace.go # FileReplaceAction
        └── zipoverlay.go  # ZipOverlayAction
```

The manifest path defaults to `/etc/foundry/patches/manifest.yaml` and is controlled by the
`FOUNDRY_PATCH_MANIFEST` env var (see `config.PathsConfig.ManifestPath`).

## Manifest loading

`manifest.Load` reads the YAML file and returns a typed `Manifest`. `manifest.Filter` narrows
the entries to those whose version range covers the currently installed Foundry build. The
activation step (`step.Patches`) calls both in sequence.

## Actions

`applier.Applier.Apply` iterates the filtered entries and dispatches each action type:

| Action | Effect |
|---|---|
| `download` | Downloads a file to a target path |
| `filereplace` | Replaces a file inside the Foundry installation tree |
| `zipoverlay` | Extracts a zip archive over a target directory |

Each action is a named struct implementing the `Action` interface — no `map[string]any`. Actions
carry only the fields they need and are independently testable.

## Testing patches

- Use fixture manifests (checked in as YAML under `test/` or `testdata/`).
- Test each `Action` implementation independently against a `t.TempDir()` tree.
- Integration tests call `applier.Applier.Apply` with a real filtered manifest.

## Forbidden

- Inline patch logic inside activation steps — delegate to `libs/foundrypatch`.
- Hard-coding patch entries in Go; they belong in the YAML manifest file.
- Actions that mutate paths outside their declared `target`.
- Returning `map[string]any` from any manifest or action function.

## See also

- [`../patterns/integration-tests.md`](../patterns/integration-tests.md) — testing with fixture
  manifests.
- [`code-placement.md`](code-placement.md) — `libs/` placement rules.
