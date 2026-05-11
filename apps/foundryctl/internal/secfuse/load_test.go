package secfuse

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func writeJSON(t *testing.T, path string, v any) {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(path, b, 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestLoad_MissingFile_ReturnsEmpty(t *testing.T) {
	res, err := Load(filepath.Join(t.TempDir(), "absent.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Applied) != 0 || len(res.Unknown) != 0 {
		t.Errorf("expected empty result, got %+v", res)
	}
}

func TestLoad_SetsEnvVars(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "secrets.json")
	writeJSON(t, p, map[string]string{
		"foundry_admin_key": "secret123",
		"foundry_username":  "user1",
	})

	// Ensure clean env before test.
	t.Setenv("FOUNDRY_ADMIN_KEY", "")
	t.Setenv("FOUNDRY_USERNAME", "")

	res, err := Load(p)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(res.Applied, "FOUNDRY_ADMIN_KEY") {
		t.Errorf("FOUNDRY_ADMIN_KEY not in Applied: %v", res.Applied)
	}
	if !slices.Contains(res.Applied, "FOUNDRY_USERNAME") {
		t.Errorf("FOUNDRY_USERNAME not in Applied: %v", res.Applied)
	}
	if os.Getenv("FOUNDRY_ADMIN_KEY") != "secret123" {
		t.Errorf("env FOUNDRY_ADMIN_KEY = %q, want %q", os.Getenv("FOUNDRY_ADMIN_KEY"), "secret123")
	}
	if len(res.Unknown) != 0 {
		t.Errorf("expected no unknowns, got %v", res.Unknown)
	}
}

func TestLoad_UnknownKeysReported(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "secrets.json")
	writeJSON(t, p, map[string]string{
		"foundry_admin_key": "x",
		"unknown_key":       "y",
	})
	t.Setenv("FOUNDRY_ADMIN_KEY", "")

	res, err := Load(p)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(res.Unknown, "unknown_key") {
		t.Errorf("expected unknown_key in Unknown, got %v", res.Unknown)
	}
}

func TestLoad_EmptyStringSkipped(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "secrets.json")
	writeJSON(t, p, map[string]string{
		"foundry_admin_key": "",
	})
	t.Setenv("FOUNDRY_ADMIN_KEY", "original")

	res, err := Load(p)
	if err != nil {
		t.Fatal(err)
	}
	if slices.Contains(res.Applied, "FOUNDRY_ADMIN_KEY") {
		t.Error("empty value should not be applied")
	}
	// env should be untouched
	if os.Getenv("FOUNDRY_ADMIN_KEY") != "original" {
		t.Errorf("env should be untouched, got %q", os.Getenv("FOUNDRY_ADMIN_KEY"))
	}
}

func TestLoad_AppliedIsSorted(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "secrets.json")
	writeJSON(t, p, map[string]string{
		"foundry_username":  "u",
		"foundry_admin_key": "k",
		"foundry_password":  "p",
	})
	for _, k := range []string{"FOUNDRY_USERNAME", "FOUNDRY_ADMIN_KEY", "FOUNDRY_PASSWORD"} {
		t.Setenv(k, "")
	}

	res, err := Load(p)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.IsSorted(res.Applied) {
		t.Errorf("Applied not sorted: %v", res.Applied)
	}
}
