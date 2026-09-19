package profloader

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/profile"
)

const (
	profAlice = "alice"
	profBob   = "bob"
)

func profilesPath(t testing.TB) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "profiles.json")
}

func seedFile(t testing.TB, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("seed: %v", err)
	}
}

func profileNames(profiles []profile.Profile) []string {
	names := make([]string, 0, len(profiles))
	for _, p := range profiles {
		names = append(names, p.Name)
	}
	return names
}

func assertStored(t testing.TB, path string, wantNames []string, wantActive string) {
	t.Helper()
	profiles, active, err := FromFile(path)
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	if names := profileNames(profiles); !slices.Equal(names, wantNames) {
		t.Errorf("profiles = %v, want %v", names, wantNames)
	}
	if active != wantActive {
		t.Errorf("active = %q, want %q", active, wantActive)
	}
}
