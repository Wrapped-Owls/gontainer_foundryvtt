package lifecycle

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundryruntime/config"
)

// ConfigDir is the conventional Foundry "Config" subdirectory of the
// data path where options.json and admin.txt live.
func ConfigDir(dataPath string) string { return filepath.Join(dataPath, "Config") }

// WriteOptions renders c into <dataPath>/Config/options.json. If the
// existing file already has identical bytes the write is skipped (so a
// container restart with unchanged env doesn't bump mtimes).
//
// Returns true iff the file was (re)written.
func WriteOptions(dataPath string, c config.Config) (bool, error) {
	dir := ConfigDir(dataPath)
	if err := os.MkdirAll(dir, dirPerm); err != nil {
		return false, fmt.Errorf("lifecycle: mkdir %s: %w", dir, err)
	}
	dest := filepath.Join(dir, "options.json")
	var buf bytes.Buffer
	if err := config.WriteConfig(&buf, c); err != nil {
		return false, err
	}
	if existing, err := os.ReadFile(dest); err == nil && bytes.Equal(existing, buf.Bytes()) {
		return false, nil
	}
	if err := os.WriteFile(dest, buf.Bytes(), filePerm); err != nil {
		return false, fmt.Errorf("lifecycle: write %s: %w", dest, err)
	}
	return true, nil
}
