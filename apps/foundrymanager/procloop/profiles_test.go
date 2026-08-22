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
	t.Parallel()

	r := makeStore(t, "", nil)
	if err := r.CreateProfile(profile.Profile{Name: profAlice, DataPath: "/d/alice"}); err != nil {
		t.Fatalf("create: %v", err)
	}
	got, ok := r.GetProfile(profAlice)
	if !ok || got.DataPath != "/d/alice" {
		t.Errorf("profile not stored: %+v ok=%v", got, ok)
	}
}

func TestCreateProfile_validation(t *testing.T) {
	t.Parallel()

	r := makeStore(t, "", []profile.Profile{{Name: profAlice, DataPath: "/d"}})
	if err := r.CreateProfile(profile.Profile{Name: profBob}); !errors.Is(err, profile.ErrInvalid) {
		t.Errorf("expected ErrInvalid, got %v", err)
	}
	if err := r.CreateProfile(
		profile.Profile{Name: profAlice, DataPath: "/x"},
	); !errors.Is(
		err,
		profile.ErrExists,
	) {
		t.Errorf("expected ErrExists, got %v", err)
	}
}

func TestUpdateProfile_appliesOnlyEditableFields(t *testing.T) {
	t.Parallel()

	r := makeStore(t, "", []profile.Profile{{Name: profAlice, DataPath: "/d", AdminKey: "keep"}})
	err := r.UpdateProfile(profAlice, profile.Profile{
		Label:             "Alice",
		Version:           verProfile,
		World:             "w",
		DataPath:          "/evil",
		ManifestPath:      "/evil/manifest",
		AdminKey:          "hijack",
		AdminPasswordSalt: "hijack",
	})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	got, _ := r.GetProfile(profAlice)
	if got.Label != "Alice" || got.Version != verProfile || got.World != "w" {
		t.Errorf("editable fields not applied: %+v", got)
	}
	if got.DataPath != "/d" || got.ManifestPath != "" ||
		got.AdminKey != "keep" || got.AdminPasswordSalt != "" {
		t.Errorf("immutable fields changed: %+v", got)
	}
}

func TestUpdateProfile_activeProfile_queuesSwitchOnVersionOrWorldChange(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		update     profile.Profile
		wantSwitch bool
	}{
		{"version changed", profile.Profile{Version: verOlder}, true},
		{"world changed", profile.Profile{World: "avalon"}, true},
		{"label only", profile.Profile{Label: "Alice"}, false},
		{"same version", profile.Profile{Version: verProfile}, false},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			r := makeStore(
				t, profAlice,
				[]profile.Profile{{Name: profAlice, DataPath: "/d", Version: verProfile}},
			)
			if err := r.UpdateProfile(profAlice, testCase.update); err != nil {
				t.Fatalf("update: %v", err)
			}
			select {
			case got := <-r.ctrl.SwitchCh:
				if !testCase.wantSwitch {
					t.Errorf("unexpected switch queued for %q", got)
				} else if got != profAlice {
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
	t.Parallel()

	r := makeStore(t, profAlice, []profile.Profile{
		{Name: profAlice, DataPath: "/d"},
		{Name: profBob, DataPath: "/d2", Version: verProfile},
	})
	if err := r.UpdateProfile(profBob, profile.Profile{Version: verOlder}); err != nil {
		t.Fatalf("update: %v", err)
	}
	select {
	case got := <-r.ctrl.SwitchCh:
		t.Errorf("unexpected switch queued for %q; bob is not the active profile", got)
	default:
	}
}

func TestUpdateProfile_notFound(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

	r := makeStore(t, "", []profile.Profile{{Name: profAlice, DataPath: "/d"}})
	if err := r.DeleteProfile(profAlice); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, ok := r.GetProfile(profAlice); ok {
		t.Error("profile still present after delete")
	}
}

func TestDeleteProfile_activeRefused(t *testing.T) {
	t.Parallel()

	r := makeStore(t, profAlice, []profile.Profile{{Name: profAlice, DataPath: "/d"}})
	if err := r.DeleteProfile(profAlice); !errors.Is(err, profile.ErrInvalid) {
		t.Errorf("expected ErrInvalid deleting active, got %v", err)
	}
}

func TestDeleteProfile_notFound(t *testing.T) {
	t.Parallel()

	r := makeStore(t, "", nil)
	if err := r.DeleteProfile("ghost"); !errors.Is(err, profile.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}
