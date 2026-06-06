package procloop

import (
	"log/slog"
	"testing"

	fmconfig "github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/config"
	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/profile"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrykit/backoff"
)

func makeRunner(profiles []profile.Profile) *Runner {
	return New(
		State{Profiles: profiles, Version: "14.0.0"},
		nil,
		fmconfig.Config{},
		backoff.Config{},
		slog.Default(),
	)
}

func TestRequestSwitch_unknownProfile(t *testing.T) {
	r := makeRunner([]profile.Profile{{Name: "alice"}})
	if err := r.RequestSwitch("nobody"); err == nil {
		t.Error("expected error for unknown profile")
	}
}

func TestRequestSwitch_knownProfile(t *testing.T) {
	r := makeRunner([]profile.Profile{{Name: "alice"}})
	if err := r.RequestSwitch("alice"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestActive_initial(t *testing.T) {
	r := makeRunner(nil)
	if got := r.Active(); got != "" {
		t.Errorf("expected empty initial active, got %q", got)
	}
}

func TestVersion(t *testing.T) {
	r := makeRunner(nil)
	if got := r.Version(); got != "14.0.0" {
		t.Errorf("expected 14.0.0, got %q", got)
	}
}
