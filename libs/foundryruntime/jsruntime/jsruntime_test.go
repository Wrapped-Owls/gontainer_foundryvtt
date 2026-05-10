package jsruntime

import (
	"errors"
	"testing"
)

func TestResolveDefaultsToBun(t *testing.T) {
	rt, err := Resolve(
		DefaultConfig(),
		func(name string) (string, error) {
			if name == "bun" {
				return "/usr/local/bin/bun", nil
			}
			return "", errors.New("not found")
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if rt.Kind != Bun || rt.Path != "/usr/local/bin/bun" {
		t.Fatalf("got %+v", rt)
	}
}

func TestResolveExplicitNode(t *testing.T) {
	rt, err := Resolve(Config{Kind: Node}, func(name string) (string, error) {
		return "/usr/bin/" + name, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if rt.Kind != Node || rt.Path != "/usr/bin/node" {
		t.Fatalf("got %+v", rt)
	}
}

func TestResolveExplicitPathSkipsLookup(t *testing.T) {
	called := false
	rt, err := Resolve(
		Config{Kind: Bun, Path: "/opt/bun/bin/bun"},
		func(string) (string, error) { called = true; return "", errors.New("nope") },
	)
	if err != nil || called {
		t.Fatalf("lookup should have been skipped: err=%v called=%v", err, called)
	}
	if rt.Path != "/opt/bun/bin/bun" {
		t.Fatalf("path: %q", rt.Path)
	}
}

func TestResolveUnsupported(t *testing.T) {
	_, err := Resolve(Config{Kind: "deno"}, nil)
	if !errors.Is(err, ErrUnsupported) {
		t.Fatalf("expected ErrUnsupported, got %v", err)
	}
}

func TestLoadFromEnv(t *testing.T) {
	t.Setenv("FOUNDRY_JS_RUNTIME", "node")
	t.Setenv("FOUNDRY_JS_RUNTIME_PATH", "/usr/bin/node")

	cfg := DefaultConfig()
	if err := LoadFromEnv(&cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.Kind != Node || cfg.Path != "/usr/bin/node" {
		t.Fatalf("cfg = %+v", cfg)
	}
}
