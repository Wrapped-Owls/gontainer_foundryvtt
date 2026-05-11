package source

import (
	"archive/zip"
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// --- folderSource.Materialise ---

func TestFolderSource_Materialise(t *testing.T) {
	src := t.TempDir()
	// create a minimal Foundry-like tree
	if err := os.MkdirAll(filepath.Join(src, "resources", "app"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "resources", "app", "main.mjs"), []byte("//"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		filepath.Join(src, "resources", "app", "package.json"),
		[]byte(`{"version":"14.361.5"}`),
		0o644,
	); err != nil {
		t.Fatal(err)
	}

	s := NewFolder(src)
	dst := t.TempDir()
	res, err := s.Materialise(context.Background(), dst)
	if err != nil {
		t.Fatal(err)
	}
	if res.Kind != KindFolder {
		t.Errorf("Kind = %q, want %q", res.Kind, KindFolder)
	}
	// main.mjs must be present in dst
	if _, err = os.Stat(filepath.Join(dst, "resources", "app", "main.mjs")); err != nil {
		t.Errorf("main.mjs not copied: %v", err)
	}
}

func TestFolderSource_Materialise_EmptyPath(t *testing.T) {
	s := NewFolder("")
	_, err := s.Materialise(context.Background(), t.TempDir())
	if err == nil {
		t.Error("expected error for empty path")
	}
}

// --- zipSource.Materialise ---

func buildZip(t *testing.T, entries map[string]string) []byte {
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

func TestZipSource_Materialise(t *testing.T) {
	zdata := buildZip(t, map[string]string{
		"resources/app/main.mjs":     "//",
		"resources/app/package.json": `{"version":"14.361.3"}`,
	})
	zp := filepath.Join(t.TempDir(), "foundryvtt_v14.361.3.zip")
	if err := os.WriteFile(zp, zdata, 0o644); err != nil {
		t.Fatal(err)
	}

	s := NewZip(zp)
	dst := t.TempDir()
	res, err := s.Materialise(context.Background(), dst)
	if err != nil {
		t.Fatal(err)
	}
	if res.Kind != KindZip {
		t.Errorf("Kind = %q, want %q", res.Kind, KindZip)
	}
	if _, err = os.Stat(filepath.Join(dst, "resources", "app", "main.mjs")); err != nil {
		t.Errorf("main.mjs not extracted: %v", err)
	}
}

func TestZipSource_Materialise_EmptyPath(t *testing.T) {
	s := NewZip("")
	_, err := s.Materialise(context.Background(), t.TempDir())
	if err == nil {
		t.Error("expected error for empty path")
	}
}

// --- urlSource.Materialise ---

func TestURLSource_Materialise(t *testing.T) {
	zdata := buildZip(t, map[string]string{
		"resources/app/main.mjs": "//",
	})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/zip")
		_, _ = w.Write(zdata)
	}))
	defer srv.Close()

	s := NewURL(srv.URL, http.DefaultClient, "14.361.9")
	dst := t.TempDir()
	res, err := s.Materialise(context.Background(), dst)
	if err != nil {
		t.Fatal(err)
	}
	if res.Kind != KindURL {
		t.Errorf("Kind = %q, want %q", res.Kind, KindURL)
	}
	if _, err = os.Stat(filepath.Join(dst, "resources", "app", "main.mjs")); err != nil {
		t.Errorf("main.mjs not extracted: %v", err)
	}
}

func TestURLSource_Materialise_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "gone", http.StatusGone)
	}))
	defer srv.Close()

	s := NewURL(srv.URL, http.DefaultClient, "")
	_, err := s.Materialise(context.Background(), t.TempDir())
	if err == nil {
		t.Error("expected error on HTTP 410")
	}
}

func TestURLSource_Materialise_EmptyURL(t *testing.T) {
	s := NewURL("", http.DefaultClient, "")
	_, err := s.Materialise(context.Background(), t.TempDir())
	if err == nil {
		t.Error("expected error for empty URL")
	}
}
