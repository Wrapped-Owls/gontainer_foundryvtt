package cmd

import (
	"context"
	"errors"
	"log/slog"
	"slices"
	"testing"

	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundryctl/internal/activate"
	fmconfig "github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/config"
	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/profile"
)

const (
	profAlice = "alice"
	profBob   = "bob"
)

func stateWith(active, fallback string, names ...string) activate.State {
	profiles := make([]profile.Profile, 0, len(names))
	for _, name := range names {
		profiles = append(profiles, profile.Profile{Name: name, DataPath: "/data/" + name})
	}
	var state activate.State
	state.ActiveProfile = active
	state.Profiles = profiles
	state.App.Manager = fmconfig.Config{DefaultProfile: fallback}
	return state
}

func TestProfileCandidates(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		state activate.State
		want  []string
	}{
		{
			name:  "no profiles at all leaves nothing to try",
			state: stateWith("", ""),
			want:  []string{},
		},
		{
			name:  "the first profile stands in when nothing is configured",
			state: stateWith("", "", profAlice, profBob),
			want:  []string{profAlice},
		},
		{
			name:  "the recorded active profile comes first",
			state: stateWith(profBob, "carol", profAlice, profBob, "carol"),
			want:  []string{profBob, "carol", profAlice},
		},
		{
			name:  "the configured default follows the recorded one",
			state: stateWith("", "carol", profAlice, profBob, "carol"),
			want:  []string{"carol", profAlice},
		},
		{
			name:  "a name repeated across sources is tried once",
			state: stateWith(profAlice, profAlice, profAlice, profBob),
			want:  []string{profAlice},
		},
		{
			name:  "a recorded name absent from the list is still tried first",
			state: stateWith("ghost", "", profAlice),
			want:  []string{"ghost", profAlice},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			if got := profileCandidates(testCase.state); !slices.Equal(got, testCase.want) {
				t.Fatalf("got %v, want %v", got, testCase.want)
			}
		})
	}
}

var errActivation = errors.New("data path is unwritable")

func runResolve(
	t *testing.T,
	state activate.State,
	failing []string,
) (active string, tried []string) {
	t.Helper()

	tried = make([]string, 0, len(state.Profiles))
	prepare := func(
		_ context.Context,
		_ *slog.Logger,
		s activate.State,
		p profile.Profile,
	) (activate.State, error) {
		tried = append(tried, p.Name)
		if slices.Contains(failing, p.Name) {
			return activate.State{}, errActivation
		}
		return s, nil
	}
	_, active = resolveInitialProfile(context.Background(), slog.Default(), state, prepare)
	return active, tried
}

func TestResolveInitialProfilePicksTheMostSpecific(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		state      activate.State
		wantActive string
		wantTried  []string
	}{
		{
			name:       "the recorded profile is resumed",
			state:      stateWith(profBob, "", profAlice, profBob),
			wantActive: profBob,
			wantTried:  []string{profBob},
		},
		{
			name:       "the configured default is used when nothing was recorded",
			state:      stateWith("", profBob, profAlice, profBob),
			wantActive: profBob,
			wantTried:  []string{profBob},
		},
		{
			name:       "a recorded profile that no longer exists falls through",
			state:      stateWith("ghost", "", profAlice),
			wantActive: profAlice,
			wantTried:  []string{profAlice},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			active, tried := runResolve(t, testCase.state, nil)
			if active != testCase.wantActive || !slices.Equal(tried, testCase.wantTried) {
				t.Fatalf("active=%q tried=%v, want active=%q tried=%v",
					active, tried, testCase.wantActive, testCase.wantTried)
			}
		})
	}
}

func TestResolveInitialProfileFallsThroughFailures(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		state      activate.State
		failing    []string
		wantActive string
		wantTried  []string
	}{
		{
			name:       "a profile that fails to activate falls through instead of dropping out",
			state:      stateWith(profBob, "", profAlice, profBob),
			failing:    []string{profBob},
			wantActive: profAlice,
			wantTried:  []string{profBob, profAlice},
		},
		{
			name:      "every profile failing leaves the base config",
			state:     stateWith(profBob, "", profAlice, profBob),
			failing:   []string{profAlice, profBob},
			wantTried: []string{profBob, profAlice},
		},
		{
			name:      "no profiles configured leaves the base config",
			state:     stateWith("", ""),
			wantTried: []string{},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			active, tried := runResolve(t, testCase.state, testCase.failing)
			if active != testCase.wantActive || !slices.Equal(tried, testCase.wantTried) {
				t.Fatalf("active=%q tried=%v, want active=%q tried=%v",
					active, tried, testCase.wantActive, testCase.wantTried)
			}
		})
	}
}
