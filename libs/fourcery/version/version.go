package version

import (
	"strings"

	"github.com/Masterminds/semver/v3"
)

type Version struct {
	raw    string
	parsed *semver.Version
}

func Parse(s string) Version {
	v := Version{raw: strings.TrimSpace(s)}
	if p, err := semver.NewVersion(v.raw); err == nil {
		v.parsed = p
	}
	return v
}

func (v Version) String() string {
	if v.parsed != nil {
		return v.parsed.String()
	}
	return v.raw
}

func (v Version) IsZero() bool { return v.raw == "" }

func (v Version) HasPatch() bool {
	return strings.Count(v.raw, ".") >= 2 // "14.361" is major.minor even though semver adds ".0"
}

func (v Version) Compare(other Version) int { // non-semver sorts first; two non-semver are equal
	switch {
	case v.parsed == nil && other.parsed == nil:
		return 0
	case v.parsed == nil:
		return -1
	case other.parsed == nil:
		return 1
	}
	return v.parsed.Compare(other.parsed)
}

func (v Version) Matches(desired Version) bool {
	if v.IsZero() || desired.IsZero() {
		return v.raw == desired.raw
	}
	if v.parsed == nil || desired.parsed == nil {
		return v.raw == desired.raw
	}
	if desired.HasPatch() {
		return v.parsed.Equal(desired.parsed)
	}
	return v.parsed.Major() == desired.parsed.Major() &&
		v.parsed.Minor() == desired.parsed.Minor()
}

func (v Version) DirName() string {
	if v.parsed != nil {
		return "foundryvtt_v" + v.parsed.String()
	}
	return "foundryvtt_v" + v.raw
}
