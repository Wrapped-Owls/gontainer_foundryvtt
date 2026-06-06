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

func FromFile(path string) ([]profile.Profile, error) {
	stored, err := os.ReadFile(
		path,
	) //nolint:gosec // path is sourced from operator-controlled config
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var f profileFile
	if err = json.Unmarshal(stored, &f); err != nil {
		return nil, err
	}
	return f.Profiles, nil
}
