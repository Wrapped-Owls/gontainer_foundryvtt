package profloader

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestFromFile_notExist(t *testing.T) {
	profiles, active, err := FromFile("/nonexistent/profiles.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if profiles != nil {
		t.Fatalf("expected nil, got %v", profiles)
	}
	if active != "" {
		t.Fatalf("expected empty active, got %q", active)
	}
}

func TestFromFile_valid(t *testing.T) {
	data, _ := json.Marshal(map[string]any{
		"active": "alice",
		"profiles": []map[string]any{
			{"name": "alice", "label": "Alice", "dataPath": "/data/alice"},
		},
	})
	path := filepath.Join(t.TempDir(), "profiles.json")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}

	profiles, active, err := FromFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(profiles) != 1 {
		t.Fatalf("expected 1 profile, got %d", len(profiles))
	}
	if profiles[0].Name != "alice" {
		t.Errorf("expected alice, got %q", profiles[0].Name)
	}
	if active != "alice" {
		t.Errorf("expected active alice, got %q", active)
	}
}

func TestFromFile_malformed(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bad.json")
	if err := os.WriteFile(path, []byte("{not json}"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := FromFile(path); err == nil {
		t.Error("expected error for malformed JSON")
	}
}

func TestWriteActive_createAndUpdate(t *testing.T) {
	path := filepath.Join(t.TempDir(), "profiles.json")

	if err := WriteActive(path, "bob"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, active, err := FromFile(path)
	if err != nil {
		t.Fatalf("unexpected error reading back: %v", err)
	}
	if active != "bob" {
		t.Errorf("expected bob, got %q", active)
	}

	if err := WriteActive(path, "alice"); err != nil {
		t.Fatalf("unexpected error on update: %v", err)
	}
	_, active, err = FromFile(path)
	if err != nil {
		t.Fatalf("unexpected error reading back: %v", err)
	}
	if active != "alice" {
		t.Errorf("expected alice after update, got %q", active)
	}
}

func TestWriteActive_preservesProfiles(t *testing.T) {
	data, _ := json.Marshal(map[string]any{
		"profiles": []map[string]any{
			{"name": "alice", "dataPath": "/d/alice"},
		},
	})
	path := filepath.Join(t.TempDir(), "profiles.json")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}

	if err := WriteActive(path, "alice"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	profiles, active, err := FromFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if active != "alice" {
		t.Errorf("expected alice, got %q", active)
	}
	if len(profiles) != 1 || profiles[0].Name != "alice" {
		t.Errorf("profiles not preserved: %+v", profiles)
	}
}
