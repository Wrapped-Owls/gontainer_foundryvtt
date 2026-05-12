package source

import (
	"archive/zip"
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestRegistry_Enumerate_EmptyConfig(t *testing.T) {
	reg := NewRegistry(Config{})
	got, err := reg.Enumerate(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("want 0 sources, got %d", len(got))
	}
}

func TestRegistry_Enumerate_MissingSourcesDir(t *testing.T) {
	reg := NewRegistry(Config{SourcesDir: filepath.Join(t.TempDir(), "absent")})
	got, err := reg.Enumerate(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("want 0 sources for missing dir, got %d", len(got))
	}
}

func TestRegistry_Enumerate_OrdersAndKinds(t *testing.T) {
	dir := t.TempDir()
	// folder
	folder := filepath.Join(dir, "foundryvtt_v14.361.0")
	if err := os.MkdirAll(filepath.Join(folder, "resources", "app"), 0o755); err != nil {
		t.Fatal(err)
	}
	// zip
	zipPath := filepath.Join(dir, "foundryvtt_v14.361.1.zip")
	writeZip(t, zipPath, map[string]string{"main.mjs": "//"})

	// hidden
	if err := os.Mkdir(filepath.Join(dir, ".cache"), 0o755); err != nil {
		t.Fatal(err)
	}

	reg := NewRegistry(Config{
		SourcesDir: dir,
		ReleaseURL: "https://example.invalid/x.zip",
		Version:    "14.361.1",
		Username:   "u",
		Password:   "p",
	})
	got, err := reg.Enumerate(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	wantKinds := []Kind{KindFolder, KindZip, KindSession, KindURL}
	if len(got) != len(wantKinds) {
		t.Fatalf("want %d sources, got %d", len(wantKinds), len(got))
	}
	for i, w := range wantKinds {
		if got[i].Kind() != w {
			t.Errorf("source[%d] kind = %q, want %q", i, got[i].Kind(), w)
		}
	}
}

func TestZipSource_Probe(t *testing.T) {
	dir := t.TempDir()
	zp := filepath.Join(dir, "foundryvtt_v14.361.4.zip")
	writeZip(t, zp, map[string]string{"main.mjs": "//"})

	s := NewZip(zp)
	v, err := s.Probe(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if v != "14.361.4" {
		t.Errorf("want 14.361.4, got %q", v)
	}
}

func TestFolderSource_Probe_FallbackToPackageJSON(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "unnamed")
	if err := os.MkdirAll(filepath.Join(target, "resources", "app"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		filepath.Join(target, "resources", "app", "package.json"),
		[]byte(`{"version":"14.361.7"}`),
		0o644,
	); err != nil {
		t.Fatal(err)
	}
	s := NewFolder(target)
	v, err := s.Probe(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if v != "14.361.7" {
		t.Errorf("want 14.361.7, got %q", v)
	}
}

func TestURLSource_ProbeWithoutLabel(t *testing.T) {
	s := NewURL("https://example.invalid/x.zip", nil, "", "")
	_, err := s.Probe(context.Background())
	if err != ErrVersionUnknown {
		t.Errorf("want ErrVersionUnknown, got %v", err)
	}
}

func TestURLSource_ProbeWithLabel(t *testing.T) {
	s := NewURL("https://example.invalid/x.zip", nil, "14.361.0", "")
	v, err := s.Probe(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if v != "14.361.0" {
		t.Errorf("want 14.361.0, got %q", v)
	}
}

func writeZip(t *testing.T, path string, entries map[string]string) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	zw := zip.NewWriter(f)
	for name, body := range entries {
		w, werr := zw.Create(name)
		if werr != nil {
			t.Fatal(werr)
		}
		if _, werr = io.WriteString(w, body); werr != nil {
			t.Fatal(werr)
		}
	}
	if err = zw.Close(); err != nil {
		t.Fatal(err)
	}
}
