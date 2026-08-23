package step

import (
	"cmp"
	"context"
	"errors"
	"log/slog"
	"slices"
	"testing"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundryruntime/jsruntime"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/fourcery/forge"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/fourcery/version"
)

const (
	binBun    = "bun"
	binNode22 = "node22"
	binNode24 = "node24"

	verFoundry13 = "13.351.0"
	verFoundry14 = "14.361.2"
)

func discardLogger() *slog.Logger { return slog.New(slog.DiscardHandler) }

func lookupIn(available ...string) func(string) (string, error) {
	return func(name string) (string, error) {
		if slices.Contains(available, name) {
			return "/bin/" + name, nil
		}
		return "", errors.New("not found")
	}
}

func resolveRuntime(t *testing.T, foundry string, available ...string) jsruntime.Runtime {
	t.Helper()

	state := &State{Install: forge.Install{Version: version.Parse(foundry)}}
	step := jsRuntimeStep{lookPath: lookupIn(available...)}
	if err := step.Apply(context.Background(), state, discardLogger()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return state.JSRuntime
}

func TestJSRuntimeStepDefaultsToBun(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		foundry   string
		available []string
	}{
		{
			name:      "no runtime configured means bun, whatever the version resolved",
			foundry:   verFoundry14,
			available: []string{binBun, binNode22, binNode24},
		},
		{
			name:      "bun stays bun on a v13 install too",
			foundry:   verFoundry13,
			available: []string{binBun, binNode22},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			got := resolveRuntime(t, testCase.foundry, testCase.available...)
			if got.Kind != jsruntime.Bun || got.Path != "/bin/"+binBun {
				t.Fatalf("got %+v, want bun at /bin/%s", got, binBun)
			}
		})
	}
}

func TestJSRuntimeStepPicksTheNodeMajorTheVersionNeeds(t *testing.T) {
	testCases := []struct {
		name      string
		foundry   string
		available []string
		wantBin   string
	}{
		{
			name:      "a v14 install takes node 24",
			foundry:   verFoundry14,
			available: []string{binBun, binNode22, binNode24},
			wantBin:   binNode24,
		},
		{
			name:      "a v13 install takes node 22",
			foundry:   verFoundry13,
			available: []string{binBun, binNode22, binNode24},
			wantBin:   binNode22,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Setenv("FOUNDRY_JS_RUNTIME", string(jsruntime.Node))

			got := resolveRuntime(t, testCase.foundry, testCase.available...)
			if got.Kind != jsruntime.Node || got.Path != "/bin/"+testCase.wantBin {
				t.Fatalf("got %+v, want node at /bin/%s", got, testCase.wantBin)
			}
		})
	}
}

func TestJSRuntimeStepReportsAnUnusableRuntime(t *testing.T) {
	testCases := []struct {
		name      string
		runtime   string
		foundry   string
		available []string
		wantErr   error
	}{
		{
			name:    "an unsupported runtime is rejected before any lookup",
			runtime: "deno",
			wantErr: jsruntime.ErrUnsupported,
		},
		{
			name:      "a runtime missing from the image is reported, not silently swapped",
			runtime:   string(jsruntime.Node),
			available: []string{binBun},
			wantErr:   jsruntime.ErrNotFound,
		},
		{
			name:    "an unparseable install version is reported, never guessed",
			runtime: string(jsruntime.Node),
			foundry: "latest",
			wantErr: jsruntime.ErrUnknownFoundry,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Setenv("FOUNDRY_JS_RUNTIME", testCase.runtime)

			state := &State{Install: forge.Install{
				Version: version.Parse(cmp.Or(testCase.foundry, verFoundry14)),
			}}
			step := jsRuntimeStep{lookPath: lookupIn(testCase.available...)}

			err := step.Apply(context.Background(), state, discardLogger())
			if !errors.Is(err, testCase.wantErr) {
				t.Fatalf("err = %v, want %v", err, testCase.wantErr)
			}
		})
	}
}
