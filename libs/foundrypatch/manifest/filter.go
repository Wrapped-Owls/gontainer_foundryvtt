package manifest

import (
	"fmt"

	"github.com/Masterminds/semver/v3"
)

// Applicable returns the subset of patches whose Versions constraint is
// satisfied by version. version is parsed leniently — a leading "v" or
// pre-release tag is tolerated.
func (f *File) Applicable(version string) ([]Patch, error) {
	v, err := semver.NewVersion(version)
	if err != nil {
		return nil, fmt.Errorf("manifest: parse version %q: %w", version, err)
	}
	var out []Patch
	for _, p := range f.Patches {
		c, err := semver.NewConstraint(p.Versions)
		if err != nil {
			return nil, fmt.Errorf("manifest: patch %q: %w", p.ID, err)
		}
		if c.Check(v) {
			out = append(out, p)
		}
	}
	return out, nil
}
