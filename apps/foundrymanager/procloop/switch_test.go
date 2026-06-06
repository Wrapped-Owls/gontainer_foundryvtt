package procloop

import (
	"context"
	"log/slog"
	"testing"

	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/internal/controller"
	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/profile"
)

// stubActivator implements Activator for testing.
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

func makeRunnerWithProfiles(profiles []profile.Profile) *Runner {
	return &Runner{
		logger: slog.Default(),
		ctrl:   controller.New(),
		state:  State{Profiles: profiles},
	}
}

func TestFindProfile_found(t *testing.T) {
	r := makeRunnerWithProfiles([]profile.Profile{
		{Name: "alice", DataPath: "/data/alice"},
		{Name: "bob", DataPath: "/data/bob"},
	})
	p, ok := r.findProfile("bob")
	if !ok {
		t.Fatal("expected bob to be found")
	}
	if p.DataPath != "/data/bob" {
		t.Errorf("unexpected DataPath: %q", p.DataPath)
	}
}

func TestFindProfile_notFound(t *testing.T) {
	r := makeRunnerWithProfiles([]profile.Profile{{Name: "alice"}})
	_, ok := r.findProfile("charlie")
	if ok {
		t.Error("expected not found")
	}
}

func TestApplySwitch_success(t *testing.T) {
	activator := &stubActivator{result: State{Version: "14.1.0"}}
	r := makeRunnerWithProfiles([]profile.Profile{{Name: "alice"}})
	r.activator = activator

	r.ctrl.SwitchCh <- "alice"

	if err := r.applySwitch(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !activator.called {
		t.Error("Activator.Switch was not called")
	}
	if r.state.Version != "14.1.0" {
		t.Errorf("state not updated: got version %q", r.state.Version)
	}
	if r.ctrl.Active() != "alice" {
		t.Errorf("active profile not set: got %q", r.ctrl.Active())
	}
}

func TestApplySwitch_unknownProfile(t *testing.T) {
	r := makeRunnerWithProfiles([]profile.Profile{{Name: "alice"}})
	r.ctrl.SwitchCh <- "unknown"
	if err := r.applySwitch(context.Background()); err == nil {
		t.Error("expected error for unknown profile")
	}
}

func TestApplySwitch_noPending(t *testing.T) {
	r := makeRunnerWithProfiles(nil)
	if err := r.applySwitch(context.Background()); err != nil {
		t.Errorf("expected nil error when no switch pending: %v", err)
	}
}
