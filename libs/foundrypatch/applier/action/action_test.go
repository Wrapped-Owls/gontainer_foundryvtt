package action

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrypatch/internal/testzip"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrypatch/manifest"
)

func bodySum(b []byte) string { s := sha256.Sum256(b); return hex.EncodeToString(s[:]) }

// --- Download ---

func TestDownloadRunner_WritesFile(t *testing.T) {
	body := []byte("hello download")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(body)
	}))
	defer srv.Close()

	dest := filepath.Join(t.TempDir(), "sub", "file.txt")
	act := manifest.Action{
		Type:   manifest.ActionDownload,
		URL:    srv.URL,
		SHA256: bodySum(body),
		Dest:   "sub/file.txt",
	}
	if err := Download(http.DefaultClient).Run(context.Background(), act, dest); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(dest)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, body) {
		t.Errorf("content mismatch: got %q, want %q", got, body)
	}
}

func TestDownloadRunner_HashMismatch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("payload"))
	}))
	defer srv.Close()

	dest := filepath.Join(t.TempDir(), "out")
	act := manifest.Action{
		Type:   manifest.ActionDownload,
		URL:    srv.URL,
		SHA256: "deadbeef",
		Dest:   "out",
	}
	if err := Download(http.DefaultClient).Run(context.Background(), act, dest); err == nil {
		t.Fatal("expected hash mismatch error")
	}
}

func TestDownloadRunner_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "nope", http.StatusForbidden)
	}))
	defer srv.Close()

	dest := filepath.Join(t.TempDir(), "out")
	act := manifest.Action{Type: manifest.ActionDownload, URL: srv.URL, SHA256: "xx", Dest: "out"}
	if err := Download(http.DefaultClient).Run(context.Background(), act, dest); err == nil {
		t.Fatal("expected error on HTTP 403")
	}
}

// --- FileReplace ---

func TestFileReplaceRunner_WritesContent(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "nested", "target.txt")
	act := manifest.Action{Type: manifest.ActionFileReplace, Content: "replaced!"}
	if err := FileReplace().Run(context.Background(), act, dest); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(dest)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "replaced!" {
		t.Errorf("got %q, want %q", got, "replaced!")
	}
}

func TestFileReplaceRunner_OverwritesExisting(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "file.txt")
	if err := os.WriteFile(dest, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	act := manifest.Action{Type: manifest.ActionFileReplace, Content: "new"}
	if err := FileReplace().Run(context.Background(), act, dest); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(dest)
	if string(got) != "new" {
		t.Errorf("got %q, want %q", got, "new")
	}
}

// --- ZipOverlay ---

func TestZipOverlayRunner_ExtractsFiles(t *testing.T) {
	overlay := testzip.MakeZip(t, map[string]string{
		"subdir/patch.txt": "patched",
	})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(overlay)
	}))
	defer srv.Close()

	dest := t.TempDir()
	act := manifest.Action{
		Type:   manifest.ActionZipOverlay,
		URL:    srv.URL,
		SHA256: bodySum(overlay),
	}
	if err := ZipOverlay(http.DefaultClient).Run(context.Background(), act, dest); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Join(dest, "subdir", "patch.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "patched" {
		t.Errorf("got %q, want %q", got, "patched")
	}
}

func TestZipOverlayRunner_HashMismatch(t *testing.T) {
	overlay := testzip.MakeZip(t, map[string]string{"f.txt": "x"})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(overlay)
	}))
	defer srv.Close()

	act := manifest.Action{Type: manifest.ActionZipOverlay, URL: srv.URL, SHA256: "bad"}
	if err := ZipOverlay(http.DefaultClient).Run(context.Background(), act, t.TempDir()); err == nil {
		t.Fatal("expected hash mismatch error")
	}
}
