// Package profile defines the Profile type and its JSON persistence.
package profile

// Profile holds per-GM configuration overrides applied on top of the base config.
// Each field, if non-empty, replaces the corresponding base value for the session.
type Profile struct {
	Name              string `json:"name"`
	Label             string `json:"label"`
	DataPath          string `json:"dataPath"`
	AdminKey          string `json:"adminKey"`
	AdminPasswordSalt string `json:"adminPasswordSalt"`
	Version           string `json:"version"`
	ManifestPath      string `json:"manifestPath"`
}
