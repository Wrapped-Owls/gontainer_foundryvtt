package jsruntime

import (
	"errors"
	"slices"
	"testing"
)

func lookupIn(available []string, asked *[]string) func(string) (string, error) {
	return func(name string) (string, error) {
		*asked = append(*asked, name)
		if slices.Contains(available, name) {
			return "/usr/local/bin/" + name, nil
		}
		return "", errors.New("not found")
	}
}

type resolveCase struct {
	name         string
	config       Config
	foundryMajor FoundryMajor
	available    []string
	wantKind     Kind
	wantPath     string
	wantAsked    []string
	wantErr      error
}

func runResolveCases(t *testing.T, testCases []resolveCase) {
	t.Helper()

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			asked := make([]string, 0, 2)
			rt, err := Resolve(
				testCase.config,
				testCase.foundryMajor,
				lookupIn(testCase.available, &asked),
			)

			if !errors.Is(err, testCase.wantErr) {
				t.Fatalf("err = %v, want %v", err, testCase.wantErr)
			}
			if testCase.wantErr == nil &&
				(rt.Kind != testCase.wantKind || rt.Path != testCase.wantPath) {
				t.Fatalf("got %+v, want kind=%v path=%q",
					rt, testCase.wantKind, testCase.wantPath)
			}
			if !slices.Equal(asked, testCase.wantAsked) {
				t.Fatalf("probed %v, want %v", asked, testCase.wantAsked)
			}
		})
	}
}

func TestResolveDefaultsToBun(t *testing.T) {
	t.Parallel()

	testCases := []resolveCase{
		{
			name:      "the default kind is bun",
			config:    DefaultConfig(),
			available: []string{string(Bun)},
			wantKind:  Bun,
			wantPath:  "/usr/local/bin/" + string(Bun),
			wantAsked: []string{string(Bun)},
		},
		{
			name:      "an empty kind falls back to the default",
			config:    Config{},
			available: []string{string(Bun)},
			wantKind:  Bun,
			wantPath:  "/usr/local/bin/" + string(Bun),
			wantAsked: []string{string(Bun)},
		},
		{
			name:         "bun ignores the foundry major entirely",
			config:       Config{Kind: Bun},
			foundryMajor: 14,
			available:    []string{string(Bun)},
			wantKind:     Bun,
			wantPath:     "/usr/local/bin/" + string(Bun),
			wantAsked:    []string{string(Bun)},
		},
	}

	runResolveCases(t, testCases)
}

func TestResolvePicksTheNodeMajor(t *testing.T) {
	t.Parallel()

	testCases := []resolveCase{
		{
			name:         "foundry v13 takes node 22",
			config:       Config{Kind: Node},
			foundryMajor: 13,
			available:    []string{binNode22, binNode24},
			wantKind:     Node,
			wantPath:     "/usr/local/bin/" + binNode22,
			wantAsked:    []string{binNode22},
		},
		{
			name:         "foundry v14 takes node 24",
			config:       Config{Kind: Node},
			foundryMajor: 14,
			available:    []string{binNode22, binNode24},
			wantKind:     Node,
			wantPath:     "/usr/local/bin/" + binNode24,
			wantAsked:    []string{binNode24},
		},
		{
			name:         "a newer foundry stays on the newest known node",
			config:       Config{Kind: Node},
			foundryMajor: 15,
			available:    []string{binNode22, binNode24},
			wantKind:     Node,
			wantPath:     "/usr/local/bin/" + binNode24,
			wantAsked:    []string{binNode24},
		},
		{
			name:         "a foundry older than v13 still runs on node 22",
			config:       Config{Kind: Node},
			foundryMajor: 12,
			available:    []string{binNode22, binNode24},
			wantKind:     Node,
			wantPath:     "/usr/local/bin/" + binNode22,
			wantAsked:    []string{binNode22},
		},
	}

	runResolveCases(t, testCases)
}

func TestResolveRejectsWhatItCannotRun(t *testing.T) {
	t.Parallel()

	testCases := []resolveCase{
		{
			name:         "an explicit path skips the lookup entirely",
			config:       Config{Kind: Node, Path: "/opt/node/bin/node"},
			foundryMajor: 14,
			wantKind:     Node,
			wantPath:     "/opt/node/bin/node",
			wantAsked:    []string{},
		},
		{
			name:         "no node on PATH is an error naming what was tried",
			config:       Config{Kind: Node},
			foundryMajor: 14,
			available:    []string{string(Bun)},
			wantAsked:    []string{binNode24},
			wantErr:      ErrNotFound,
		},
		{
			name:         "an unparseable version is reported, never guessed",
			config:       Config{Kind: Node},
			foundryMajor: 0,
			available:    []string{binNode22, binNode24},
			wantAsked:    []string{},
			wantErr:      ErrUnknownFoundry,
		},
		{
			name:      "an unsupported kind is rejected before any lookup",
			config:    Config{Kind: "deno"},
			wantAsked: []string{},
			wantErr:   ErrUnsupported,
		},
	}

	runResolveCases(t, testCases)
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

func TestLoadFromEnvRejectsAnUnknownKind(t *testing.T) {
	t.Setenv("FOUNDRY_JS_RUNTIME", "deno")

	cfg := DefaultConfig()
	if err := LoadFromEnv(&cfg); !errors.Is(err, ErrUnsupported) {
		t.Fatalf("err = %v, want %v", err, ErrUnsupported)
	}
}

func TestLoadFromEnvDefaultsToBunWhenUnset(t *testing.T) {
	cfg := DefaultConfig()
	if err := LoadFromEnv(&cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Kind != Bun {
		t.Fatalf("Kind = %v, want %v: node must never be reached by accident", cfg.Kind, Bun)
	}
	if cfg.Path != "" {
		t.Fatalf("Path = %q, want empty", cfg.Path)
	}
}
