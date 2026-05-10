// Package confloader provides a generic typed configuration loader that
// reads from a JSON file and overlays environment variable values.
//
// Usage pattern for a library package:
//
//	// In yourpkg/config.go:
//	const envFoo = "YOUR_FOO"
//
//	type Config struct { Foo string }
//
//	func Default() Config { return Config{Foo: "default"} }
//
//	func LoadFromEnv(c *Config) error {
//	    return confloader.BindEnv(
//	        confloader.BindField(&c.Foo, envFoo, nil),
//	    )
//	}
//
// Usage pattern for an app's confwire package:
//
//	func Load() (Config, error) {
//	    return confloader.Load("conf.toml", Default(), LoadFromEnv)
//	}
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
// A missing file is not an error — defaults remain.
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
		// Try JSON first (simple; avoids a toml dep), then treat as plain
		// key=value if JSON fails. For now, only JSON is supported.
		// Libraries that need TOML can add the dependency themselves.
		if jsonErr := json.Unmarshal(b, &cfg); jsonErr != nil {
			return cfg, fmt.Errorf("confloader: parse %s: %w", path, jsonErr)
		}
	}
	if err := updater(&cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}
