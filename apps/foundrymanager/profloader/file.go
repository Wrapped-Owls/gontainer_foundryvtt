package profloader

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/profile"
)

type profileFile struct {
	Active   string            `json:"active,omitempty"`
	Profiles []profile.Profile `json:"profiles"`
}

func FromFile(path string) (profiles []profile.Profile, active string, err error) {
	var stored []byte
	stored, err = os.ReadFile(path) //nolint:gosec // operator-configured path
	if errors.Is(err, os.ErrNotExist) {
		return nil, "", nil
	}
	if err != nil {
		return nil, "", err
	}
	var f profileFile
	if err = json.Unmarshal(stored, &f); err != nil {
		return nil, "", err
	}
	return f.Profiles, f.Active, nil
}

func WriteActive(path, name string) error {
	stored, _ := os.ReadFile(path) //nolint:gosec
	var f profileFile
	if len(stored) > 0 {
		_ = json.Unmarshal(stored, &f)
	}
	f.Active = name
	out, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(out, '\n'), 0o600) //nolint:gosec
}
