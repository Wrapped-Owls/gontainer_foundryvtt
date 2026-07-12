// Package profloader loads Profile lists from JSON files and environment variables.
package profloader

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/profile"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrykit/fsperm"
)

type profileFile struct {
	Active   string            `json:"active,omitempty"`
	Profiles []profile.Profile `json:"profiles"`
}

// FromFile reads profiles and the last-active name from a JSON file at path.
// Returns nil profiles and empty active without error when the file does not exist.
func FromFile(path string) (profiles []profile.Profile, active string, err error) {
	data, err := os.ReadFile(path) //nolint:gosec // path is sourced from operator-controlled config
	if errors.Is(err, os.ErrNotExist) {
		return nil, "", nil
	}
	if err != nil {
		return nil, "", err
	}
	var f profileFile
	if err = json.Unmarshal(data, &f); err != nil {
		return nil, "", err
	}
	return f.Profiles, f.Active, nil
}

// WriteProfiles persists the given profiles into the JSON file at path,
// preserving the recorded active profile name. Creates the file if absent.
func WriteProfiles(path string, profiles []profile.Profile) error {
	data, _ := os.ReadFile(path) //nolint:gosec
	var f profileFile
	if len(data) > 0 {
		_ = json.Unmarshal(data, &f)
	}
	f.Profiles = profiles
	out, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(out, '\n'), fsperm.Secret)
}

// WriteActive persists the active profile name into the JSON file at path,
// preserving any existing profiles array. Creates the file if absent.
func WriteActive(path, name string) error {
	data, _ := os.ReadFile(path) //nolint:gosec
	var f profileFile
	if len(data) > 0 {
		_ = json.Unmarshal(data, &f)
	}
	f.Active = name
	out, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(out, '\n'), 0o600) //nolint:gosec
}
