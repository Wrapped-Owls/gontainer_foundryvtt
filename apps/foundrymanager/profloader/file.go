package profloader

import (
	"bytes"
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
	if len(bytes.TrimSpace(stored)) == 0 {
		return nil, "", nil
	}
	var f profileFile
	if err = json.Unmarshal(stored, &f); err != nil {
		return nil, "", err
	}
	return f.Profiles, f.Active, nil
}
