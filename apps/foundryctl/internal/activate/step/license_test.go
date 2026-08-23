package step

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	appconfig "github.com/wrapped-owls/gontainer_foundryvtt/apps/foundryctl/config"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundryruntime/lifecycle"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/fourcery/forge"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/fourcery/version"
)

func licenseState(dataPath, cacheDir, ver string) *State {
	return &State{
		App:     appconfig.Config{Paths: appconfig.PathsConfig{DataPath: dataPath, LicenseCache: cacheDir}},
		Install: forge.Install{Version: version.Parse(ver)},
	}
}

func TestLicenseStepNoopsWhenCacheDirUnset(t *testing.T) {
	t.Parallel()
	dataPath := t.TempDir()

	s := licenseState(dataPath, "", "14.361.2")
	err := licenseStep{}.Apply(context.Background(), s, discardLogger())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLicenseStepHarvestsAndSeedsRoundTrip(t *testing.T) {
	t.Parallel()
	dataPath := t.TempDir()
	cacheDir := t.TempDir()

	existing := filepath.Join(lifecycle.ConfigDir(dataPath), "license.json")
	if err := os.MkdirAll(filepath.Dir(existing), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(existing, []byte(`{"version":"13.351.0"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	s := licenseState(dataPath, cacheDir, "13.351.0")
	err := licenseStep{}.Apply(context.Background(), s, discardLogger())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cached := filepath.Join(cacheDir, "13.351.0", "license.json")
	if _, statErr := os.Stat(cached); statErr != nil {
		t.Fatalf("license should have been harvested into the cache: %v", statErr)
	}
}

func TestLicenseStepFailsOnMalformedLicenseFile(t *testing.T) {
	t.Parallel()
	dataPath := t.TempDir()
	cacheDir := t.TempDir()

	existing := filepath.Join(lifecycle.ConfigDir(dataPath), "license.json")
	if err := os.MkdirAll(filepath.Dir(existing), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(existing, []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}

	s := licenseState(dataPath, cacheDir, "13.351.0")
	err := licenseStep{}.Apply(context.Background(), s, discardLogger())
	if err == nil {
		t.Fatal("expected an error for a malformed license.json")
	}
}

func TestLicenseStepIgnoresAVersionThatEscapesTheCacheDir(t *testing.T) {
	t.Parallel()
	dataPath := t.TempDir()
	cacheDir := t.TempDir()

	s := licenseState(dataPath, cacheDir, "../escape")
	err := licenseStep{}.Apply(context.Background(), s, discardLogger())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	entries, readErr := os.ReadDir(cacheDir)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if len(entries) != 0 {
		t.Fatalf("cache dir should stay empty for a traversal-attempting version, got %v", entries)
	}
}
