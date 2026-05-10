package lifecycle

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrykit/fsperm"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundryruntime/config"
)

func ConfigDir(dataPath string) string { return filepath.Join(dataPath, "Config") }

func WriteOptions(dataPath string, c config.Config) (bool, error) {
	dir := ConfigDir(dataPath)
	if err := os.MkdirAll(dir, fsperm.Dir); err != nil {
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
	if err := os.WriteFile(dest, buf.Bytes(), fsperm.File); err != nil {
		return false, fmt.Errorf("lifecycle: write %s: %w", dest, err)
	}
	return true, nil
}
