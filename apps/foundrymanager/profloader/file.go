package profloader

import (
	"bytes"
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

// FromFile treats a missing or empty file as no profiles, not an error: the image
// pre-creates it, and empty-as-malformed would fail activation before Foundry starts.
func FromFile(path string) (profiles []profile.Profile, active string, err error) {
	var data []byte
	data, err = os.ReadFile(path) //nolint:gosec // path is sourced from operator-controlled config
	if errors.Is(err, os.ErrNotExist) {
		return nil, "", nil
	}
	if err != nil {
		return nil, "", err
	}
	if len(bytes.TrimSpace(data)) == 0 {
		return nil, "", nil
	}
	var f profileFile
	if err = json.Unmarshal(data, &f); err != nil {
		return nil, "", err
	}
	return f.Profiles, f.Active, nil
}

// WriteProfiles preserves the recorded active name; it creates the file if absent.
func WriteProfiles(path string, profiles []profile.Profile) error {
	data, _ := os.ReadFile(path) //nolint:gosec // path is sourced from operator-controlled config
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

// WriteActive preserves any existing profiles array; it creates the file if absent.
func WriteActive(path, name string) error {
	data, _ := os.ReadFile(path) //nolint:gosec // path is sourced from operator-controlled config
	var f profileFile
	if len(data) > 0 {
		_ = json.Unmarshal(data, &f)
	}
	f.Active = name
	out, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(out, '\n'), fsperm.Secret)
}
