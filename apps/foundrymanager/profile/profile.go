package profile

import "errors"

// Sentinel errors returned by profile management operations. Callers map these
// to transport-specific responses via errors.Is.
var (
	ErrNotFound = errors.New("profile: not found")
	ErrExists   = errors.New("profile: already exists")
	ErrInvalid  = errors.New("profile: invalid")
)

// Profile holds per-GM configuration overrides applied on top of the base config.
// Each field, if non-empty, replaces the corresponding base value for the session.
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
