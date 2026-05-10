package manifest

import (
	"fmt"

	"github.com/Masterminds/semver/v3"
)

func (f *File) Applicable(version string) ([]Patch, error) {
	v, err := semver.NewVersion(version)
	if err != nil {
		return nil, fmt.Errorf("manifest: parse version %q: %w", version, err)
	}
	var out []Patch
	for _, p := range f.Patches {
		c, constraintErr := semver.NewConstraint(p.Versions)
		if constraintErr != nil {
			return nil, fmt.Errorf("manifest: patch %q: %w", p.ID, constraintErr)
		}
		if c.Check(v) {
			out = append(out, p)
		}
	}
	return out, nil
}
