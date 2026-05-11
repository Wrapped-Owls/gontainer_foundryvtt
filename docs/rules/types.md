# Types

Prefer **named structs** over `map[string]any` for anything that crosses a function or package
boundary. Use **small interfaces declared at the consumer**, not large interfaces exported by
the implementor.

## Named structs over maps

```go
// ❌ Bad — opaque payload
func parseInbound(raw []byte) (map[string]any, error) { ... }

// ✅ Good — typed result
type Inbound struct {
    Channel  string
    From     string
    Body     string
    SentAt   time.Time
}

func parseInbound(raw []byte) (Inbound, error) { ... }
```

`map[string]any` is acceptable only at the raw I/O boundary (e.g. unmarshalling an unknown JSON
shape into a buffer). Translate to a typed value as soon as possible.

## Consumer‑defined interfaces

Define the interface where it is used, with the smallest method set the consumer needs.

```go
// internal/activate/step/install.go
type downloader interface {
    Fetch(ctx context.Context, sess *auth.Session, version string, opts release.FetchOptions) (string, error)
}

type installStep struct {
    dl downloader
}
```

`libs/fourcery/release.Fetch` satisfies `downloader` structurally. The consumer doesn't
need to import the implementor; the implementor doesn't need to know about the consumer.

The `interfacebloat` linter flags interfaces with too many methods — split them by responsibility.

## Value vs pointer

- Use value receivers for small, immutable types (`entities.HexID`, value objects).
- Use pointer receivers when the type holds state (a service with a logger and a repo) or is
  large enough that copying matters.
- Pick one and apply it to **every** method on that type. Mixing receivers is a smell.

## `any`

- Allowed in generic type parameters where the concrete type is known at the call site.
- Forbidden as a public function parameter without a documented reason. Use a typed struct.
- Use `any`, not the legacy `interface{}` spelling. The `modernize` linter rewrites it.

## Generics

Generics are **encouraged**. When a function would otherwise differ only by element type, reach
for a type parameter — it gives the compiler one more piece of information and removes a class
of copy‑paste bugs.

The canonical example in this repo is `libs/foundrykit/confloader.BindField[T]`, which collapses
what would otherwise be one binder per primitive type into a single parametric helper:

```go
func BindField[T any](ptr *T, envKey string, parser func(string) (T, error)) Binder { ... }
```

Other shapes that benefit from generics:

- Container helpers (`slices.Map`, `slices.Filter`, `set.New[T]`).
- Helpers that process slices of a typed element (e.g. filtering patch entries by version).
- Result/option utilities, especially when paired with the typed `apperr` flow.
- Test helpers that compare or render arbitrary value types.

The single anti‑pattern worth calling out: **don't reach for generics when one concrete typed
function is already clearer**. A `func ParseUserID(s string) (entities.HexID, error)` does not
need to become `func Parse[T any](s string) (T, error)` just to be reusable. Generics are a tool
for removing duplication, not for postponing type decisions.

## Enumerations

Use a typed string with package‑level constants:

```go
type Status string

const (
    StatusDraft     Status = "draft"
    StatusActive    Status = "active"
    StatusFinalized Status = "finalized"
)
```

Validate at the boundary; assume valid afterwards.

## Zero values are valid

Design types so that the zero value is safe to use. If construction needs work, hide the zero
value behind an unexported struct and require `New(...)`.
