package testzip

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"testing"
)

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
		var w io.Writer
		w, err = zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = io.WriteString(w, body); err != nil {
			t.Fatal(err)
		}
	}
	if err = zw.Close(); err != nil {
		t.Fatal(err)
	}
	return zp
}
