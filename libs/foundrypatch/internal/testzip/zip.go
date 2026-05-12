package testzip

import (
	"archive/zip"
	"bytes"
	"io"
	"testing"
)

// MakeZip builds an in-memory zip containing entries (path → content)
// and returns the raw bytes. Used by tests that serve the zip over HTTP
// or pass it directly to an applier.
func MakeZip(t *testing.T, entries map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
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
	return buf.Bytes()
}
