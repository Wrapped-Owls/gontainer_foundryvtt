package copytree

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestCopy_FilesAndDirs(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	if err := os.MkdirAll(filepath.Join(src, "resources", "app"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		filepath.Join(src, "resources", "app", "main.mjs"),
		[]byte("// main\n"),
		0o644,
	); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		filepath.Join(src, "resources", "app", "package.json"),
		[]byte(`{"version":"14.361.2"}`),
		0o600,
	); err != nil {
		t.Fatal(err)
	}

	if err := Copy(src, dst); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(filepath.Join(dst, "resources", "app", "main.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "// main\n" {
		t.Errorf("main.mjs body mismatch: %q", string(got))
	}
	info, err := os.Stat(filepath.Join(dst, "resources", "app", "package.json"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Errorf("package.json mode = %v, want 0o600", info.Mode().Perm())
	}
}

func TestCopy_RejectsEscapingSymlink(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()
	outside := t.TempDir()

	if err := os.WriteFile(filepath.Join(outside, "secret"), []byte("nope"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(outside, "secret"), filepath.Join(src, "escape")); err != nil {
		t.Fatal(err)
	}

	err := Copy(src, dst)
	if !errors.Is(err, ErrUnsafeLink) {
		t.Fatalf("want ErrUnsafeLink, got %v", err)
	}
}

func TestCopy_AllowsInternalSymlink(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	if err := os.WriteFile(filepath.Join(src, "real"), []byte("data"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("real", filepath.Join(src, "link")); err != nil {
		t.Fatal(err)
	}

	if err := Copy(src, dst); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(filepath.Join(dst, "link"))
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != "data" {
		t.Errorf("link content = %q, want %q", string(body), "data")
	}
}
