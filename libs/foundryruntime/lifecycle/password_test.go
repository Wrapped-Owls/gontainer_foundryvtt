package lifecycle

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteAdminPassword(t *testing.T) {
	dir := t.TempDir()
	wrote, err := WriteAdminPassword(dir, "secret", "")
	if err != nil || !wrote {
		t.Fatalf("create: wrote=%v err=%v", wrote, err)
	}
	dest := filepath.Join(dir, "Config", "admin.txt")
	first, _ := os.ReadFile(dest)
	if len(first) == 0 {
		t.Fatal("admin.txt empty")
	}
	wrote, err = WriteAdminPassword(dir, "secret", "")
	if err != nil || wrote {
		t.Errorf("re-write same key should be no-op: wrote=%v err=%v", wrote, err)
	}
	wrote, err = WriteAdminPassword(dir, "", "")
	if err != nil || !wrote {
		t.Errorf("removal: wrote=%v err=%v", wrote, err)
	}
	if _, err := os.Stat(dest); !os.IsNotExist(err) {
		t.Errorf("expected file removed, got err=%v", err)
	}
	wrote, err = WriteAdminPassword(dir, "", "")
	if err != nil || wrote {
		t.Errorf("removing missing file should be no-op: wrote=%v err=%v", wrote, err)
	}
}
