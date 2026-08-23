package confloader

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
)

// Load reads a JSON config file at path (honouring the CONF_FILE env var
// override), then calls updater to overlay environment variables.
// A missing file is not an error: defaults remain.
func Load[C any](filename string, defaults C, updater func(*C) error) (C, error) {
	path := filename
	if v := os.Getenv("CONF_FILE"); v != "" {
		path = v
	}
	cfg := defaults
	b, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return cfg, fmt.Errorf("confloader: read %s: %w", path, err)
	}
	if err == nil {
		// Only JSON is supported; a library needing TOML brings its own dep.
		if jsonErr := json.Unmarshal(b, &cfg); jsonErr != nil {
			return cfg, fmt.Errorf("confloader: parse %s: %w", path, jsonErr)
		}
	}
	if updateErr := updater(&cfg); updateErr != nil {
		return cfg, updateErr
	}
	return cfg, nil
}
