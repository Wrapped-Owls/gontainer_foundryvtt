package profile

import "errors"

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
