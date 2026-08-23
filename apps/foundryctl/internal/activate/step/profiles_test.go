package step

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	appconfig "github.com/wrapped-owls/gontainer_foundryvtt/apps/foundryctl/config"
)

func TestProfilesStepLoadsFromFileAndMergesEnv(t *testing.T) {
	profilesPath := filepath.Join(t.TempDir(), "profiles.json")
	body := `{"active":"alice","profiles":[{"name":"alice","label":"Alice"}]}`
	if err := os.WriteFile(profilesPath, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("FOUNDRY_PROFILE_ALICE_NAME", "alice")
	t.Setenv("FOUNDRY_PROFILE_ALICE_LABEL", "Alice Overridden")

	s := &State{App: appconfig.Config{}}
	s.App.Manager.ProfilesFile = profilesPath

	err := profilesStep{}.Apply(context.Background(), s, discardLogger())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.ActiveProfile != "alice" {
		t.Fatalf("ActiveProfile = %q, want %q", s.ActiveProfile, "alice")
	}
	if len(s.Profiles) != 1 || s.Profiles[0].Label != "Alice Overridden" {
		t.Fatalf("Profiles = %+v, want the env override applied", s.Profiles)
	}
}

func TestProfilesStepTreatsAMissingFileAsEmpty(t *testing.T) {
	s := &State{App: appconfig.Config{}}
	s.App.Manager.ProfilesFile = filepath.Join(t.TempDir(), "absent.json")

	err := profilesStep{}.Apply(context.Background(), s, discardLogger())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.Profiles) != 0 || s.ActiveProfile != "" {
		t.Fatalf("expected no profiles and no active profile, got %+v / %q", s.Profiles, s.ActiveProfile)
	}
}

func TestProfilesStepFailsOnMalformedFile(t *testing.T) {
	profilesPath := filepath.Join(t.TempDir(), "profiles.json")
	if err := os.WriteFile(profilesPath, []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}

	s := &State{App: appconfig.Config{}}
	s.App.Manager.ProfilesFile = profilesPath

	err := profilesStep{}.Apply(context.Background(), s, discardLogger())
	if err == nil {
		t.Fatal("expected an error for a malformed profiles file")
	}
}
