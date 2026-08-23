package step

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	appconfig "github.com/wrapped-owls/gontainer_foundryvtt/apps/foundryctl/config"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrypatch/ledger"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/fourcery/forge"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/fourcery/version"
)

func writeManifest(t *testing.T, dir, dest string) string {
	t.Helper()
	path := filepath.Join(dir, "manifest.yaml")
	body := "version: 1\npatches:\n  - id: patch-1\n    versions: \">=1.0.0\"\n    actions:\n      - type: file-replace\n        dest: " + dest + "\n        content: patched\n"
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func patchesState(root, manifestPath string) *State {
	return &State{
		App: appconfig.Config{
			Paths:   appconfig.PathsConfig{ManifestPath: manifestPath},
			Install: appconfig.InstallConfig{Version: "14.361.2"},
		},
		Install: forge.Install{Root: root, Version: version.Parse("14.361.2")},
	}
}

func TestPatchesStepSkipsGracefullyOnLoadOrFilterFailure(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		setup func(t *testing.T, root string) string
	}{
		{
			name: "manifest path is a directory, not a file",
			setup: func(t *testing.T, root string) string {
				dir := filepath.Join(root, "manifest-is-a-dir")
				if err := os.MkdirAll(dir, 0o755); err != nil {
					t.Fatal(err)
				}
				return dir
			},
		},
		{
			name: "manifest file is missing",
			setup: func(t *testing.T, root string) string {
				return filepath.Join(root, "absent-manifest.yaml")
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			manifestPath := testCase.setup(t, root)

			s := patchesState(root, manifestPath)
			err := patchesStep{}.Apply(context.Background(), s, discardLogger())
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestPatchesStepSkipsOnFilterFailure(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	manifestPath := writeManifest(t, root, "patched.txt")

	s := patchesState(root, manifestPath)
	s.Install.Version = version.Parse("not-a-semver")
	s.App.Install.Version = "not-a-semver"

	err := patchesStep{}.Apply(context.Background(), s, discardLogger())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, statErr := os.Stat(ledger.Path(root)); !os.IsNotExist(statErr) {
		t.Fatalf("ledger should not have been written, stat err = %v", statErr)
	}
}

func TestPatchesStepNoApplicablePatchesIsANoop(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	manifestPath := writeManifest(t, root, "patched.txt")

	s := patchesState(root, manifestPath)
	s.Install.Version = version.Parse("0.1.0")
	s.App.Install.Version = "0.1.0"

	err := patchesStep{}.Apply(context.Background(), s, discardLogger())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, statErr := os.Stat(ledger.Path(root)); !os.IsNotExist(statErr) {
		t.Fatalf("ledger should not have been written, stat err = %v", statErr)
	}
}

func TestPatchesStepRecoversFromCorruptLedger(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	manifestPath := writeManifest(t, root, "patched.txt")
	if err := os.WriteFile(ledger.Path(root), []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}

	s := patchesState(root, manifestPath)
	err := patchesStep{}.Apply(context.Background(), s, discardLogger())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(root, "patched.txt"))
	if err != nil {
		t.Fatalf("patch content was not written: %v", err)
	}
	if string(got) != "patched" {
		t.Fatalf("patch content = %q, want %q", got, "patched")
	}

	rebuilt, err := ledger.Load(root)
	if err != nil {
		t.Fatalf("ledger should be valid after rebuild, got error: %v", err)
	}
	if len(rebuilt.Entries) != 1 || rebuilt.Entries[0].ID != "patch-1" {
		t.Fatalf("rebuilt ledger entries = %+v, want one entry for patch-1", rebuilt.Entries)
	}
}

func TestPatchesStepAbortsOnUnsupportedLedgerSchema(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	manifestPath := writeManifest(t, root, "patched.txt")
	future := `{"schema_version": 99, "entries": []}`
	if err := os.WriteFile(ledger.Path(root), []byte(future), 0o644); err != nil {
		t.Fatal(err)
	}

	s := patchesState(root, manifestPath)
	err := patchesStep{}.Apply(context.Background(), s, discardLogger())
	if err == nil {
		t.Fatal("expected an error for an unsupported ledger schema version")
	}
	if errors.Is(err, ledger.ErrLedgerCorrupt) {
		t.Fatalf("unsupported schema must not be treated as corrupt: %v", err)
	}
	if !errors.Is(err, ledger.ErrSchemaUnsupported) {
		t.Fatalf("err = %v, want wrapping %v", err, ledger.ErrSchemaUnsupported)
	}
	if _, statErr := os.Stat(filepath.Join(root, "patched.txt")); !os.IsNotExist(statErr) {
		t.Fatal("patch must not be applied when the ledger fails to load")
	}
}

func TestPatchesStepAbortsWhenAPatchActionFails(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	manifestPath := writeManifest(t, root, "../escapes-root.txt")

	s := patchesState(root, manifestPath)
	err := patchesStep{}.Apply(context.Background(), s, discardLogger())
	if err == nil {
		t.Fatal("expected an error when a patch action's dest escapes the install root")
	}
	if _, statErr := os.Stat(ledger.Path(root)); !os.IsNotExist(statErr) {
		t.Fatal("ledger must not be saved when a patch action fails")
	}
}

func TestPatchesStepReportsLedgerSaveFailureAfterSuccessfulApply(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	sub := filepath.Join(root, "sub")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	manifestPath := writeManifest(t, root, "sub/patched.txt")
	existing := `{"schema_version": 1, "entries": []}`
	if err := os.WriteFile(ledger.Path(root), []byte(existing), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := os.Chmod(root, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(root, 0o755) })

	s := patchesState(root, manifestPath)
	err := patchesStep{}.Apply(context.Background(), s, discardLogger())
	if err == nil {
		t.Fatal("expected a ledger save error")
	}
	if !errors.Is(err, ledger.ErrLedgerWriteFailed) {
		t.Fatalf("err = %v, want wrapping %v", err, ledger.ErrLedgerWriteFailed)
	}

	got, readErr := os.ReadFile(filepath.Join(sub, "patched.txt"))
	if readErr != nil {
		t.Fatalf("patch content should already be on disk: %v", readErr)
	}
	if string(got) != "patched" {
		t.Fatalf("patch content = %q, want %q", got, "patched")
	}
}

func TestPatchesStepFullSuccess(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	manifestPath := writeManifest(t, root, "patched.txt")

	s := patchesState(root, manifestPath)
	err := patchesStep{}.Apply(context.Background(), s, discardLogger())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(root, "patched.txt"))
	if err != nil {
		t.Fatalf("patch content was not written: %v", err)
	}
	if string(got) != "patched" {
		t.Fatalf("patch content = %q, want %q", got, "patched")
	}

	saved, err := ledger.Load(root)
	if err != nil {
		t.Fatalf("ledger should load cleanly: %v", err)
	}
	if len(saved.Entries) != 1 || saved.Entries[0].ID != "patch-1" {
		t.Fatalf("saved ledger entries = %+v, want one entry for patch-1", saved.Entries)
	}
}
