package version

import (
	"strings"

	"github.com/Masterminds/semver/v3"
)

// Version represents a FoundryVTT release version. It may carry a
// canonical semver (parseable) or a raw opaque string (non-semver).
type Version struct {
	// raw is the original trimmed input. Used for HasPatch detection
	// and string-equality fallback on non-semver values.
	raw    string
	parsed *semver.Version
}

// Parse creates a Version from a raw string. The value is trimmed; if
// it parses as semver, String() returns the canonical semver form.
func Parse(s string) Version {
	v := Version{raw: strings.TrimSpace(s)}
	if p, err := semver.NewVersion(v.raw); err == nil {
		v.parsed = p
	}
	return v
}

// String returns the canonical string representation: semver canonical
// form when parseable, the original trimmed input otherwise.
func (v Version) String() string {
	if v.parsed != nil {
		return v.parsed.String()
	}
	return v.raw
}

// IsZero reports whether v is the zero value (empty / unset).
func (v Version) IsZero() bool { return v.raw == "" }

// HasPatch reports whether the original input contains a patch component
// (two or more dot separators, e.g. "14.361.2"). This reflects user
// intent: "14.361" is treated as a major.minor constraint even though
// semver would expand it to "14.361.0".
func (v Version) HasPatch() bool {
	return strings.Count(v.raw, ".") >= 2
}

// Compare returns -1, 0, or 1, like strings.Compare. Non-semver values
// sort before any parseable semver; two non-semver values are considered
// equal.
func (v Version) Compare(other Version) int {
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

// Matches reports whether v satisfies a desired version constraint.
// When desired has a patch component an exact semver match is required;
// otherwise only major.minor must agree. Falls back to trimmed string
// equality when either value is not parseable as semver.
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

// DirName returns the canonical install-directory name for this version:
// "foundryvtt_v<semver>" when parseable, "foundryvtt_v<raw>" otherwise.
func (v Version) DirName() string {
	if v.parsed != nil {
		return "foundryvtt_v" + v.parsed.String()
	}
	return "foundryvtt_v" + v.raw
}
