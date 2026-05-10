package lifecycle

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectInstalledMissing(t *testing.T) {
	info, err := DetectInstalled(t.TempDir())
	if err != nil || info.Present {
		t.Fatalf("got %+v err=%v", info, err)
	}
}

func TestDetectInstalledPresent(t *testing.T) {
	root := t.TempDir()
	app := filepath.Join(root, "resources", "app")
	if err := os.MkdirAll(app, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(app, "main.mjs"), []byte("//"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(app, "package.json"), []byte(`{"version":"12.331"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	info, err := DetectInstalled(root)
	if err != nil {
		t.Fatal(err)
	}
	if !info.Present || info.Version != "12.331" {
		t.Errorf("got %+v", info)
	}
}
