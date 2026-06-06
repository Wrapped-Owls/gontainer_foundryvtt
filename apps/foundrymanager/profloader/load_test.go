package profloader

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/profile"
)

func writeProfilesFile(t *testing.T, profiles []map[string]any) string {
	t.Helper()
	data, _ := json.Marshal(map[string]any{"profiles": profiles})
	path := filepath.Join(t.TempDir(), "profiles.json")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestMerge_envOnly(t *testing.T) {
	result := Merge(nil, []profile.Profile{{Name: "alice", DataPath: "/d/alice"}})
	if len(result) != 1 || result[0].Name != "alice" {
		t.Errorf("unexpected: %+v", result)
	}
}

func TestMerge_fileOnly(t *testing.T) {
	base := []profile.Profile{{Name: "alice", Label: "Alice"}}
	result := Merge(base, nil)
	if len(result) != 1 || result[0].Label != "Alice" {
		t.Errorf("unexpected: %+v", result)
	}
}

func TestMerge_overlap(t *testing.T) {
	base := []profile.Profile{{Name: "alice", Label: "Old", DataPath: "/old"}}
	ov := []profile.Profile{{Name: "alice", DataPath: "/new"}}
	result := Merge(base, ov)
	if len(result) != 1 {
		t.Fatalf("expected 1, got %d", len(result))
	}
	if result[0].DataPath != "/new" {
		t.Errorf("expected /new, got %q", result[0].DataPath)
	}
	if result[0].Label != "Old" {
		t.Errorf("label should be unchanged: %q", result[0].Label)
	}
}

func TestLoad_fileAndEnv(t *testing.T) {
	path := writeProfilesFile(t, []map[string]any{
		{"name": "alice", "dataPath": "/file/alice"},
	})
	t.Setenv("TEST_LOAD_0_NAME", "bob")
	t.Setenv("TEST_LOAD_0_DATA_PATH", "/env/bob")

	profiles, err := Load(path, "TEST_LOAD")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(profiles) != 2 {
		t.Fatalf("expected 2, got %d: %+v", len(profiles), profiles)
	}
}
