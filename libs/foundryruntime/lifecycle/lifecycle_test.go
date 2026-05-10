package lifecycle

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectInstalledMissing(t *testing.T) {
	installed, err := DetectInstalled(t.TempDir())
	if err != nil || installed.Present {
		t.Fatalf("got %+v err=%v", installed, err)
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
	if err := os.WriteFile(
		filepath.Join(app, "package.json"),
		[]byte(`{"version":"12.331"}`),
		0o644,
	); err != nil {
		t.Fatal(err)
	}
	installed, err := DetectInstalled(root)
	if err != nil {
		t.Fatal(err)
	}
	if !installed.Present || installed.Version != "12.331" {
		t.Errorf("got %+v", installed)
	}
}
