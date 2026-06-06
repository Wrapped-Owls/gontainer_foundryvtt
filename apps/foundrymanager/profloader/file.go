// Package profloader loads Profile lists from JSON files and environment variables.
package profloader

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/profile"
)

type profileFile struct {
	Profiles []profile.Profile `json:"profiles"`
}

// FromFile reads profiles from a JSON file at path.
// Returns nil without error when the file does not exist.
func FromFile(path string) ([]profile.Profile, error) {
	data, err := os.ReadFile(path) //nolint:gosec // path is sourced from operator-controlled config
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var f profileFile
	if err = json.Unmarshal(data, &f); err != nil {
		return nil, err
	}
	return f.Profiles, nil
}
