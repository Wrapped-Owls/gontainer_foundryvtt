# Imports

Three import groups, separated by a single blank line, in this order:

1. **Standard library**
2. **External modules** (third-party dependencies)
3. **This module** (`github.com/wrapped-owls/gontainer_foundryvtt/...`)

`gofumpt` plus `golangci-lint`'s `goimports`/`gci` formatter (run via `task tools:fmt`) maintain
this layout. Do not reorder by hand.

## Aliasing

- Don't alias unless there is a name collision or the package name reads poorly at the call site.
- When a lib config package collides with the app config package, alias the lib's:
  `runtimecfg "github.com/wrapped-owls/gontainer_foundryvtt/libs/foundryruntime/config"`.
- Never alias the standard library (`time`, `context`, `errors`).

## Forbidden

- Dot imports (`import . "x"`) outside `*_test.go` files. They obscure the call site and break
  IDE navigation.
- Blank imports (`import _ "x"`) outside `main.go`. If a blank import is needed in `main.go`,
  document it with a one-line comment explaining the side effect.
- Importing from `apps/...` inside `libs/...`. Direction is one‑way: apps depend on libs, never
  the reverse. See [`code-placement.md`](code-placement.md).
- Importing across apps (`apps/foo` → `apps/bar`). Share via `libs/`.
- Importing `internal/` from outside the parent module. The Go compiler enforces this; don't try
  to work around it with replace directives.

## Cyclic imports

A cycle means the package boundary is wrong. Resolutions, in order of preference:

1. Move the shared symbol down into a package both sides depend on.
2. Invert the direction by introducing a small interface in the consumer package — see
   [`types.md`](types.md).
3. Merge the two packages if the split was artificial.

Never break a cycle by importing inside a function.

## See also

- [`code-placement.md`](code-placement.md) — apps depend on libs, never the reverse
- [`types.md`](types.md) — introducing interfaces to break cycles