package lifecycle

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundryruntime/config"
)

func TestWriteOptionsIdempotent(t *testing.T) {
	dir := t.TempDir()
	c := config.Default()
	wrote, err := WriteOptions(dir, c)
	if err != nil || !wrote {
		t.Fatalf("first write: wrote=%v err=%v", wrote, err)
	}
	wrote, err = WriteOptions(dir, c)
	if err != nil || wrote {
		t.Fatalf("second write should be no-op: wrote=%v err=%v", wrote, err)
	}
	if _, err := os.Stat(filepath.Join(dir, "Config", "options.json")); err != nil {
		t.Errorf("file missing: %v", err)
	}
}
