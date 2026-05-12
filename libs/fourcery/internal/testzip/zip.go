package testzip

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"testing"
)

// MakeZip creates a zip file under t.TempDir() containing entries (path → content)
// and returns its path on disk.
func MakeZip(t *testing.T, entries map[string]string) string {
	t.Helper()
	zp := filepath.Join(t.TempDir(), "release.zip")
	f, err := os.Create(zp)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	zw := zip.NewWriter(f)
	for name, body := range entries {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = io.WriteString(w, body); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return zp
}
