package procloop

import (
	"errors"
	"log/slog"
	"path/filepath"
	"testing"

	fmconfig "github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/config"
	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/profile"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrykit/backoff"
)

func makeStore(t *testing.T, active string, profiles []profile.Profile) *Runner {
	t.Helper()
	file := filepath.Join(t.TempDir(), "profiles.json")
	return New(
		State{Profiles: profiles},
		active,
		nil,
		nil,
		fmconfig.Config{ProfilesFile: file},
		backoff.Config{},
		slog.Default(),
	)
}

func TestCreateProfile_persists(t *testing.T) {
	r := makeStore(t, "", nil)
	if err := r.CreateProfile(profile.Profile{Name: "alice", DataPath: "/d/alice"}); err != nil {
		t.Fatalf("create: %v", err)
	}
	got, ok := r.GetProfile("alice")
	if !ok || got.DataPath != "/d/alice" {
		t.Errorf("profile not stored: %+v ok=%v", got, ok)
	}
}

func TestCreateProfile_validation(t *testing.T) {
	r := makeStore(t, "", []profile.Profile{{Name: "alice", DataPath: "/d"}})
	if err := r.CreateProfile(profile.Profile{Name: "bob"}); !errors.Is(err, profile.ErrInvalid) {
		t.Errorf("expected ErrInvalid, got %v", err)
	}
	if err := r.CreateProfile(
		profile.Profile{Name: "alice", DataPath: "/x"},
	); !errors.Is(
		err,
		profile.ErrExists,
	) {
		t.Errorf("expected ErrExists, got %v", err)
	}
}

func TestUpdateProfile_appliesOnlyEditableFields(t *testing.T) {
	r := makeStore(t, "", []profile.Profile{{Name: "alice", DataPath: "/d", AdminKey: "keep"}})
	// A crafted update carries editable and immutable fields; only label, version
	// and world may take effect — data location, manifest and secrets must not.
	err := r.UpdateProfile("alice", profile.Profile{
		Label:             "Alice",
		Version:           "14.0.0",
		World:             "w",
		DataPath:          "/evil",
		ManifestPath:      "/evil/manifest",
		AdminKey:          "hijack",
		AdminPasswordSalt: "hijack",
	})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	got, _ := r.GetProfile("alice")
	if got.Label != "Alice" || got.Version != "14.0.0" || got.World != "w" {
		t.Errorf("editable fields not applied: %+v", got)
	}
	if got.DataPath != "/d" || got.ManifestPath != "" ||
		got.AdminKey != "keep" || got.AdminPasswordSalt != "" {
		t.Errorf("immutable fields changed: %+v", got)
	}
}

func TestUpdateProfile_activeProfile_queuesSwitchOnVersionOrWorldChange(t *testing.T) {
	testCases := []struct {
		name       string
		update     profile.Profile
		wantSwitch bool
	}{
		{"version changed", profile.Profile{Version: "13.0.0"}, true},
		{"world changed", profile.Profile{World: "avalon"}, true},
		{"label only", profile.Profile{Label: "Alice"}, false},
		{"same version", profile.Profile{Version: "14.0.0"}, false},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			r := makeStore(
				t, "alice",
				[]profile.Profile{{Name: "alice", DataPath: "/d", Version: "14.0.0"}},
			)
			if err := r.UpdateProfile("alice", testCase.update); err != nil {
				t.Fatalf("update: %v", err)
			}
			select {
			case got := <-r.ctrl.SwitchCh:
				if !testCase.wantSwitch {
					t.Errorf("unexpected switch queued for %q", got)
				} else if got != "alice" {
					t.Errorf("expected switch queued for alice, got %q", got)
				}
			default:
				if testCase.wantSwitch {
					t.Error("expected a switch to be queued")
				}
			}
		})
	}
}

func TestUpdateProfile_inactiveProfile_doesNotQueueSwitch(t *testing.T) {
	r := makeStore(t, "alice", []profile.Profile{
		{Name: "alice", DataPath: "/d"},
		{Name: "bob", DataPath: "/d2", Version: "14.0.0"},
	})
	if err := r.UpdateProfile("bob", profile.Profile{Version: "13.0.0"}); err != nil {
		t.Fatalf("update: %v", err)
	}
	select {
	case got := <-r.ctrl.SwitchCh:
		t.Errorf("unexpected switch queued for %q; bob is not the active profile", got)
	default:
	}
}

func TestUpdateProfile_notFound(t *testing.T) {
	r := makeStore(t, "", nil)
	if err := r.UpdateProfile(
		"ghost",
		profile.Profile{World: "w"},
	); !errors.Is(
		err,
		profile.ErrNotFound,
	) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestDeleteProfile_removes(t *testing.T) {
	r := makeStore(t, "", []profile.Profile{{Name: "alice", DataPath: "/d"}})
	if err := r.DeleteProfile("alice"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, ok := r.GetProfile("alice"); ok {
		t.Error("profile still present after delete")
	}
}

func TestDeleteProfile_activeRefused(t *testing.T) {
	r := makeStore(t, "alice", []profile.Profile{{Name: "alice", DataPath: "/d"}})
	if err := r.DeleteProfile("alice"); !errors.Is(err, profile.ErrInvalid) {
		t.Errorf("expected ErrInvalid deleting active, got %v", err)
	}
}

func TestDeleteProfile_notFound(t *testing.T) {
	r := makeStore(t, "", nil)
	if err := r.DeleteProfile("ghost"); !errors.Is(err, profile.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}
