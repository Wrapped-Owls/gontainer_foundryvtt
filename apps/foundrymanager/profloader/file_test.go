package profloader

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestFromFile_notExist(t *testing.T) {
	profiles, err := FromFile("/nonexistent/profiles.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if profiles != nil {
		t.Fatalf("expected nil, got %v", profiles)
	}
}

func TestFromFile_valid(t *testing.T) {
	data, _ := json.Marshal(map[string]any{
		"profiles": []map[string]any{
			{"name": "alice", "label": "Alice", "dataPath": "/data/alice"},
		},
	})
	path := filepath.Join(t.TempDir(), "profiles.json")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}

	profiles, err := FromFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(profiles) != 1 {
		t.Fatalf("expected 1 profile, got %d", len(profiles))
	}
	if profiles[0].Name != "alice" {
		t.Errorf("expected alice, got %q", profiles[0].Name)
	}
}

func TestFromFile_malformed(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bad.json")
	if err := os.WriteFile(path, []byte("{not json}"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := FromFile(path); err == nil {
		t.Error("expected error for malformed JSON")
	}
}
