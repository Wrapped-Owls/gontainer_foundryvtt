package step

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	appconfig "github.com/wrapped-owls/gontainer_foundryvtt/apps/foundryctl/config"
	runtimecfg "github.com/wrapped-owls/gontainer_foundryvtt/libs/foundryruntime/config"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundryruntime/lifecycle"
)

func optionsState(dataPath string) *State {
	return &State{
		App: appconfig.Config{
			Paths: appconfig.PathsConfig{DataPath: dataPath},
			Admin: appconfig.AdminConfig{Key: "hunter2", PasswordSalt: "salt"},
		},
		Runtime: runtimecfg.Default(),
	}
}

func TestOptionsStepWritesBothFilesOnSuccess(t *testing.T) {
	t.Parallel()
	dataPath := t.TempDir()

	err := optionsStep{}.Apply(context.Background(), optionsState(dataPath), discardLogger())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, statErr := os.Stat(filepath.Join(lifecycle.ConfigDir(dataPath), "options.json")); statErr != nil {
		t.Fatalf("options.json was not written: %v", statErr)
	}
	if _, statErr := os.Stat(filepath.Join(lifecycle.ConfigDir(dataPath), "admin.txt")); statErr != nil {
		t.Fatalf("admin.txt was not written: %v", statErr)
	}
}

func TestOptionsStepFailsWhenOptionsWriteFails(t *testing.T) {
	t.Parallel()
	dataPath := t.TempDir()
	regularFile := filepath.Join(dataPath, "blocks-config-dir")
	if err := os.WriteFile(regularFile, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	err := optionsStep{}.Apply(context.Background(), optionsState(regularFile), discardLogger())
	if err == nil {
		t.Fatal("expected an error when the Config directory cannot be created")
	}
}

func TestOptionsStepFailsWhenAdminWriteFailsAfterOptionsSucceed(t *testing.T) {
	t.Parallel()
	dataPath := t.TempDir()
	configDir := lifecycle.ConfigDir(dataPath)
	if err := os.MkdirAll(filepath.Join(configDir, "admin.txt"), 0o755); err != nil {
		t.Fatal(err)
	}

	err := optionsStep{}.Apply(context.Background(), optionsState(dataPath), discardLogger())
	if err == nil {
		t.Fatal("expected an error when admin.txt cannot be written")
	}

	if _, statErr := os.Stat(filepath.Join(configDir, "options.json")); statErr != nil {
		t.Fatalf("options.json should already be on disk from the first write: %v", statErr)
	}
}
