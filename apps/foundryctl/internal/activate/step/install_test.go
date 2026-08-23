package step

import (
	"archive/zip"
	"context"
	"os"
	"path/filepath"
	"testing"

	appconfig "github.com/wrapped-owls/gontainer_foundryvtt/apps/foundryctl/config"
)

func writeMainScript(t *testing.T, installDir, version string) {
	t.Helper()
	appDir := filepath.Join(installDir, "resources", "app")
	if err := os.MkdirAll(appDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(appDir, "main.mjs"), []byte("//"), 0o644); err != nil {
		t.Fatal(err)
	}
	pkg := `{"version":"` + version + `"}`
	if err := os.WriteFile(filepath.Join(appDir, "package.json"), []byte(pkg), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestInstallStepReusesAnExistingCandidate(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	candidate := filepath.Join(root, "foundryvtt_v14.361.2")
	writeMainScript(t, candidate, "14.361.2")

	s := &State{App: appconfig.Config{
		Paths:   appconfig.PathsConfig{InstallRoot: root},
		Install: appconfig.InstallConfig{Version: "14.361.2"},
	}}

	err := installStep{}.Apply(context.Background(), s, discardLogger())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Install.Root != candidate {
		t.Fatalf("Install.Root = %q, want %q", s.Install.Root, candidate)
	}
	if s.Install.Version.String() != "14.361.2" {
		t.Fatalf("Install.Version = %q, want %q", s.Install.Version.String(), "14.361.2")
	}
}

func TestInstallStepInstallsFromALocalZipSource(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	sourcesDir := filepath.Join(root, "sources")
	if err := os.MkdirAll(sourcesDir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeTestZip(t, filepath.Join(sourcesDir, "foundryvtt_v14.361.2.zip"), map[string]string{
		"resources/app/main.mjs": "//",
	})

	s := &State{App: appconfig.Config{
		Paths:   appconfig.PathsConfig{InstallRoot: root, SourcesDir: sourcesDir},
		Install: appconfig.InstallConfig{Version: "14.361.2"},
	}}

	err := installStep{}.Apply(context.Background(), s, discardLogger())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Install.Version.String() != "14.361.2" {
		t.Fatalf("Install.Version = %q, want %q", s.Install.Version.String(), "14.361.2")
	}
	if _, statErr := os.Stat(filepath.Join(s.Install.Root, "resources", "app", "main.mjs")); statErr != nil {
		t.Fatalf("expected main.mjs at the installed root: %v", statErr)
	}
}

func TestInstallStepFailsWhenInstallRootIsEmpty(t *testing.T) {
	t.Parallel()
	s := &State{App: appconfig.Config{
		Install: appconfig.InstallConfig{Version: "14.361.2"},
	}}

	err := installStep{}.Apply(context.Background(), s, discardLogger())
	if err == nil {
		t.Fatal("expected an error when the install root is empty")
	}
}

func TestInstallStepFailsWhenNothingMatchesTheRequestedVersion(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	s := &State{App: appconfig.Config{
		Paths:   appconfig.PathsConfig{InstallRoot: root},
		Install: appconfig.InstallConfig{Version: "14.361.2"},
	}}

	err := installStep{}.Apply(context.Background(), s, discardLogger())
	if err == nil {
		t.Fatal("expected an error when no candidate or source satisfies the requested version")
	}
}

func writeTestZip(t *testing.T, path string, entries map[string]string) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()

	zw := zip.NewWriter(f)
	for name, body := range entries {
		w, createErr := zw.Create(name)
		if createErr != nil {
			t.Fatal(createErr)
		}
		if _, writeErr := w.Write([]byte(body)); writeErr != nil {
			t.Fatal(writeErr)
		}
	}
	if closeErr := zw.Close(); closeErr != nil {
		t.Fatal(closeErr)
	}
}
