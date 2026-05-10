package install

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	appconfig "github.com/wrapped-owls/gontainer_foundryvtt/apps/foundryctl/config"
)

func TestEnsureInstallDownloadsFromReleaseURL(t *testing.T) {
	root := t.TempDir()
	zipBytes := nodeReleaseZip(t, "14.361.0")
	zipPath := filepath.Join(t.TempDir(), "release.zip")
	if err := os.WriteFile(zipPath, zipBytes, 0o644); err != nil {
		t.Fatalf("write release zip: %v", err)
	}
	oldDownload := downloadReleaseFunc
	downloadReleaseFunc = func(_ context.Context, _ string) (string, string, error) {
		sum := sha256.Sum256(zipBytes)
		return zipPath, hex.EncodeToString(sum[:]), nil
	}
	defer func() { downloadReleaseFunc = oldDownload }()

	cfg := appconfig.Default()
	cfg.Paths.InstallRoot = root
	cfg.Install.Version = "14.361"
	cfg.Install.ReleaseURL = "https://example.invalid/release.zip"

	got, err := EnsureInstall(context.Background(), cfg, slog.Default())
	if err != nil {
		t.Fatalf("EnsureInstall error: %v", err)
	}
	if !got.Info.Present {
		t.Fatal("expected installed info to be present")
	}
	if got.Info.Version != "14.361.0" {
		t.Fatalf("installed version = %q", got.Info.Version)
	}
	mainPath := filepath.Join(root, "foundryvtt_v14.361.0", "resources", "app", "main.mjs")
	if _, err := os.Stat(mainPath); err != nil {
		t.Fatalf("main.mjs missing: %v", err)
	}
}

func TestEnsureInstallSelectsLatestWhenVersionUnset(t *testing.T) {
	root := t.TempDir()
	makeInstall(t, filepath.Join(root, "foundryvtt_v14.360.0"), "14.360.0")
	makeInstall(t, filepath.Join(root, "foundryvtt_v14.361.0"), "14.361.0")

	cfg := appconfig.Default()
	cfg.Paths.InstallRoot = root

	got, err := EnsureInstall(context.Background(), cfg, slog.Default())
	if err != nil {
		t.Fatalf("EnsureInstall error: %v", err)
	}
	if got.Root != filepath.Join(root, "foundryvtt_v14.361.0") {
		t.Fatalf("root = %q", got.Root)
	}
}

func nodeReleaseZip(t *testing.T, version string) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	files := map[string]string{
		"main.mjs":     "console.log('foundry');\n",
		"package.json": "{\"version\":\"" + version + "\"}\n",
		"public/test":  "ok\n",
	}
	for name, body := range files {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatalf("create zip entry %s: %v", name, err)
		}
		if _, err := w.Write([]byte(body)); err != nil {
			t.Fatalf("write zip entry %s: %v", name, err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("close zip: %v", err)
	}
	return buf.Bytes()
}

func makeInstall(t *testing.T, root, version string) {
	t.Helper()
	app := filepath.Join(root, "resources", "app")
	if err := os.MkdirAll(app, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(app, "main.mjs"), []byte("//"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(app, "package.json"), []byte(`{"version":"`+version+`"}`), 0o644); err != nil {
		t.Fatal(err)
	}
}
