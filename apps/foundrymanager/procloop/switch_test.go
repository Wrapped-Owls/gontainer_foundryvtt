package procloop

import (
	"context"
	"log/slog"
	"path/filepath"
	"testing"

	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/internal/controller"
	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/profile"
	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/profloader"
)

type stubActivator struct {
	result State
	err    error
	called bool
}

func (s *stubActivator) Switch(
	_ context.Context,
	_ *slog.Logger,
	_ profile.Profile,
) (State, error) {
	s.called = true
	return s.result, s.err
}

func makeRunnerWithProfiles(t testing.TB, profiles []profile.Profile) *Runner {
	t.Helper()
	return &Runner{
		logger:       slog.Default(),
		ctrl:         controller.New(),
		state:        State{Profiles: profiles},
		profilesFile: profloader.NewWriter(filepath.Join(t.TempDir(), "profiles.json")),
	}
}

func TestFindProfile_found(t *testing.T) {
	t.Parallel()

	r := makeRunnerWithProfiles(t, []profile.Profile{
		{Name: profAlice, DataPath: "/data/alice"},
		{Name: profBob, DataPath: "/data/bob"},
	})
	p, ok := r.findProfile(profBob)
	if !ok {
		t.Fatal("expected bob to be found")
	}
	if p.DataPath != "/data/bob" {
		t.Errorf("unexpected DataPath: %q", p.DataPath)
	}
}

func TestFindProfile_notFound(t *testing.T) {
	t.Parallel()

	r := makeRunnerWithProfiles(t, []profile.Profile{{Name: profAlice}})
	_, ok := r.findProfile("charlie")
	if ok {
		t.Error("expected not found")
	}
}

func TestApplySwitch_success(t *testing.T) {
	t.Parallel()

	activator := &stubActivator{result: State{Version: "14.1.0"}}
	r := makeRunnerWithProfiles(t, []profile.Profile{{Name: profAlice}})
	r.activator = activator

	r.ctrl.SwitchCh <- profAlice

	if err := r.applySwitch(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !activator.called {
		t.Error("Activator.Switch was not called")
	}
	if r.state.Version != "14.1.0" {
		t.Errorf("state not updated: got version %q", r.state.Version)
	}
	if r.ctrl.Active() != profAlice {
		t.Errorf("active profile not set: got %q", r.ctrl.Active())
	}
}

func TestApplySwitch_preservesLiveProfileList(t *testing.T) {
	t.Parallel()

	activator := &stubActivator{result: State{Version: verOlder}}
	r := makeRunnerWithProfiles(t, []profile.Profile{
		{Name: profAlice, Version: verOlder},
		{Name: profBob, Version: verProfile},
	})
	r.activator = activator

	r.ctrl.SwitchCh <- profAlice
	if err := r.applySwitch(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(r.state.Profiles) != 2 {
		t.Fatalf("live profile list lost after switch: %+v", r.state.Profiles)
	}
	if bob, ok := r.GetProfile(profBob); !ok || bob.Version != verProfile {
		t.Errorf("bob's data not preserved: %+v ok=%v", bob, ok)
	}
}

func TestApplySwitch_unknownProfile(t *testing.T) {
	t.Parallel()

	r := makeRunnerWithProfiles(t, []profile.Profile{{Name: profAlice}})
	r.ctrl.SwitchCh <- "unknown"
	if err := r.applySwitch(context.Background()); err == nil {
		t.Error("expected error for unknown profile")
	}
}

func TestApplySwitch_noPending(t *testing.T) {
	t.Parallel()

	r := makeRunnerWithProfiles(t, nil)
	if err := r.applySwitch(context.Background()); err != nil {
		t.Errorf("expected nil error when no switch pending: %v", err)
	}
}
