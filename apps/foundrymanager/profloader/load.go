package profloader

import (
	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/profile"
)

// Load combines FromFile and FromEnv. File profiles are loaded first; env
// profiles with matching Name update those fields; unmatched names are appended.
func Load(filePath, envPrefix string) ([]profile.Profile, error) {
	base, err := FromFile(filePath)
	if err != nil {
		return nil, err
	}
	overrides, err := FromEnv(envPrefix)
	if err != nil {
		return nil, err
	}
	return Merge(base, overrides), nil
}

// Merge applies overrides on top of base: a matching Name updates non-empty
// fields; an unmatched name is appended.
func Merge(base, overrides []profile.Profile) []profile.Profile {
	if len(overrides) == 0 {
		return base
	}
	result := make([]profile.Profile, len(base))
	copy(result, base)

	for _, ov := range overrides {
		idx := indexByName(result, ov.Name)
		if idx < 0 {
			result = append(result, ov)
			continue
		}
		p := &result[idx]
		applyNonEmpty(p, ov)
	}
	return result
}

func indexByName(ps []profile.Profile, name string) int {
	for i := range ps {
		if ps[i].Name == name {
			return i
		}
	}
	return -1
}

func applyNonEmpty(dst *profile.Profile, src profile.Profile) {
	if src.Label != "" {
		dst.Label = src.Label
	}
	if src.DataPath != "" {
		dst.DataPath = src.DataPath
	}
	if src.AdminKey != "" {
		dst.AdminKey = src.AdminKey
	}
	if src.AdminPasswordSalt != "" {
		dst.AdminPasswordSalt = src.AdminPasswordSalt
	}
	if src.Version != "" {
		dst.Version = src.Version
	}
	if src.ManifestPath != "" {
		dst.ManifestPath = src.ManifestPath
	}
}
