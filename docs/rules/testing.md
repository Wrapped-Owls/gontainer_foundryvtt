# Testing

## Layout

- **All tests live in the same package as the code under test.** No black-box `<pkg>_test`
  packages. If a test needs an unexported helper, the test sits next to the helper and uses it
  directly. Forcing exports just to satisfy a separate package bloats the public API.
- Tests are colocated: `foo_test.go` next to `foo.go` in the same directory and the same
  `package <name>` declaration.
- **Integration tests** (requiring network or a real FoundryVTT tree) live under `test/` at the
  repo root or are guarded by `//go:build integration` where colocated with the code.
  See [`../patterns/integration-tests.md`](../patterns/integration-tests.md).
- Shared fixtures and helpers go in `test/internal/fixtures/`.

## Unit tests — no I/O

Unit tests **must not** touch external systems. No network, no real disk paths, no real time
sleeps. If the code under test needs one of those, depend on it through an interface and pass a
fake.

- ❌ HTTP client hitting the network — use `httptest.NewServer` or a fake `HTTPDoer`
- ❌ Reading from `/etc`, `~`, or any path you didn't create with `t.TempDir()`
- ❌ Goroutines that outlive the test (`t.Cleanup` to stop them)

If a test inevitably needs I/O, it is an integration test — see the next section.

## Integration tests

Integration tests require a real environment (network access to FoundryVTT API, a temp
filesystem tree, or a running process):

1. **Library tests** — colocated with the library code (e.g. `libs/foundrypatch/applier/`),
   guarded by `//go:build integration` when they require network or a real install tree.
2. **Activation tests** — can call `activate.Prepare` against a test environment set up with
   `t.TempDir()` and a fixture manifest.

Both flavours are compiled and run only with the `integration` tag:

```sh
go test -tags=integration ./...
```

CI runs unit tests first (no tag), then the integration suite.

## Table‑driven

Default to table‑driven tests:

```go
func TestParseStatus(t *testing.T) {
    t.Parallel()

    tests := []struct {
        name    string
        input   string
        want    Status
        wantErr error
    }{
        {"draft", "draft", StatusDraft, nil},
        {"unknown", "xyz", "", ErrUnknownStatus},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()
            got, err := ParseStatus(tt.input)
            if !errors.Is(err, tt.wantErr) {
                t.Fatalf("err = %v, want %v", err, tt.wantErr)
            }
            if got != tt.want {
                t.Fatalf("got = %v, want %v", got, tt.want)
            }
        })
    }
}
```

- `t.Parallel()` is the default for both the outer and the inner `t.Run`.
- Capture the loop variable (`tt := tt`) when targeting Go versions older than 1.22; Go ≥ 1.22
  scopes per‑iteration variables and the line is unnecessary.

## Helpers

- Helpers that fail the test call `t.Helper()` first and accept `t testing.TB`.
- Set up via constructor injection. No package‑level `var db = ...` in tests; use `t.Cleanup`.

## Fakes and mocks

- For small interfaces (one or two methods), write a hand-rolled fake in the test file.
- For larger interfaces, generate with `mockgen` (`//go:generate` directive), output as
  `_mock_test.go` colocated with the test, same package.
- Consumer-defined interfaces (see [`types.md`](types.md)) make fakes trivial — the interface
  is already minimal.

## Coverage

- Run `go test -cover ./...` locally; `task main:test` does this for CI.
- We don't enforce a coverage floor by linter — the bar is "every business branch is exercised".

## Forbidden

- `time.Sleep` in tests. Use the synchronous primitives offered by the code under test
  (channels, callbacks). If you really need a deadline, use `t.Context()` (Go 1.24+) or
  `context.WithTimeout`.
- Sharing state between tests via package globals.
- `os.Exit` from a test (panics the test runner).

---

## Test dependency isolation

Heavy test-only dependencies must live exclusively in the `test/` module. Production modules must not import integration-only dependencies.

Examples of test-only dependencies:
- `testcontainers`
- Browser automation frameworks
- Cross-app integration harnesses

## Runtime patch testing

Runtime patching systems require:
- Isolated package structure.
- Dedicated integration coverage.
- Typed HTTP clients.
- Fixture-based regression tests.
