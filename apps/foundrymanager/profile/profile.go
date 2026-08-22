package profile

import (
	"errors"
	"slices"
)

var (
	ErrNotFound = errors.New("profile: not found")
	ErrExists   = errors.New("profile: already exists")
	ErrInvalid  = errors.New("profile: invalid")
)

type Profile struct {
	Name              string `json:"name"`
	Label             string `json:"label"`
	DataPath          string `json:"dataPath"`
	AdminKey          string `json:"adminKey"`
	AdminPasswordSalt string `json:"adminPasswordSalt"`
	Version           string `json:"version"`
	World             string `json:"world"`
	ManifestPath      string `json:"manifestPath"`
}

func ByName(profiles []Profile, name string) (Profile, bool) {
	i := slices.IndexFunc(profiles, func(p Profile) bool { return p.Name == name })
	if i < 0 {
		return Profile{}, false
	}
	return profiles[i], true
}
