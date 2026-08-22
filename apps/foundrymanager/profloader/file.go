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

func FromFile(path string) (profiles []profile.Profile, active string, err error) {
	var stored []byte
	stored, err = os.ReadFile(
		path,
	) //nolint:gosec // path is sourced from operator-controlled config
	if errors.Is(err, os.ErrNotExist) {
		return nil, "", nil
	}
	if err != nil {
		return nil, "", err
	}
	if len(bytes.TrimSpace(stored)) == 0 {
		return nil, "", nil
	}
	var f profileFile
	if err = json.Unmarshal(stored, &f); err != nil {
		return nil, "", err
	}
	return f.Profiles, f.Active, nil
}

func WriteProfiles(path string, profiles []profile.Profile) error {
	stored, _ := os.ReadFile(path) //nolint:gosec // path is sourced from operator-controlled config
	var f profileFile
	if len(stored) > 0 {
		_ = json.Unmarshal(stored, &f)
	}
	f.Profiles = profiles
	out, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(out, '\n'), fsperm.Secret)
}

func WriteActive(path, name string) error {
	stored, _ := os.ReadFile(path) //nolint:gosec // path is sourced from operator-controlled config
	var f profileFile
	if len(stored) > 0 {
		_ = json.Unmarshal(stored, &f)
	}
	f.Active = name
	out, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(out, '\n'), fsperm.Secret)
}
