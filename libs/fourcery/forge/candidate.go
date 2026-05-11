package forge

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/Masterminds/semver/v3"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/fourcery/internal/probe"
)

// Candidate describes an existing FoundryVTT install discovered under
// the install root.
type Candidate struct {
	// Path is the absolute path to the install directory (the one
	// whose resources/app/package.json was read).
	Path string
	// Version is the canonical semver string when parseable, or the
	// raw package.json value otherwise.
	Version string
	// Parsed is non-nil when Version parsed as semver.
	Parsed *semver.Version
}

func newCandidate(path, version string) Candidate {
	c := Candidate{Path: path, Version: version}
	if v, err := semver.NewVersion(version); err == nil {
		c.Parsed = v
		c.Version = v.String()
	}
	return c
}

// scanCandidates walks installRoot looking for installs (a directory
// whose resources/app/main.mjs exists). It returns candidates sorted
// newest first by semver; non-semver versions sort last.
func scanCandidates(installRoot string) ([]Candidate, error) {
	out := make([]Candidate, 0)
	if c, ok, err := readCandidate(installRoot); err != nil {
		return nil, err
	} else if ok {
		out = append(out, c)
	}
	entries, err := os.ReadDir(installRoot)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return out, nil
		}
		return nil, fmt.Errorf("forge: read install root: %w", err)
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if strings.HasPrefix(e.Name(), ".") {
			continue
		}
		child := filepath.Join(installRoot, e.Name())
		c, ok, derr := readCandidate(child)
		if derr != nil {
			return nil, derr
		}
		if ok {
			out = append(out, c)
		}
	}
	slices.SortStableFunc(out, func(a, b Candidate) int {
		if a.Parsed == nil && b.Parsed == nil {
			return 0
		}
		if a.Parsed == nil {
			return 1 // non-semver sorts last
		}
		if b.Parsed == nil {
			return -1 // non-semver sorts last
		}
		return b.Parsed.Compare(a.Parsed) // newest first (descending)
	})
	return out, nil
}

// readCandidate inspects path; (Candidate, true, nil) iff path looks
// like a Foundry install (resources/app/main.mjs exists).
func readCandidate(path string) (Candidate, bool, error) {
	mainPath := filepath.Join(path, "resources", "app", "main.mjs")
	if _, err := os.Stat(mainPath); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return Candidate{}, false, nil
		}
		return Candidate{}, false, fmt.Errorf("forge: stat %s: %w", mainPath, err)
	}
	version, err := probe.Folder(path)
	if err != nil && !errors.Is(err, probe.ErrNoVersion) {
		return Candidate{}, false, fmt.Errorf("forge: probe %s: %w", path, err)
	}
	return newCandidate(path, version), true, nil
}

// matchCandidate returns the candidate satisfying desired, or nil.
// When desired has a patch component, an exact semver match is
// required. Otherwise the major.minor must match.
func matchCandidate(candidates []Candidate, desired string) *Candidate {
	parsed, err := semver.NewVersion(desired)
	if err != nil {
		for i := range candidates {
			if candidates[i].Version == strings.TrimSpace(desired) {
				return &candidates[i]
			}
		}
		return nil
	}
	requirePatch := versionHasPatch(desired)
	for i := range candidates {
		c := &candidates[i]
		if c.Parsed == nil {
			if c.Version == desired {
				return c
			}
			continue
		}
		if requirePatch {
			if c.Parsed.Equal(parsed) {
				return c
			}
			continue
		}
		if c.Parsed.Major() == parsed.Major() && c.Parsed.Minor() == parsed.Minor() {
			return c
		}
	}
	return nil
}

// versionsEqual returns true when actual and desired refer to the same
// install — equal semver when both parse, equal patch when desired
// requires it, or equal trimmed strings otherwise.
func versionsEqual(actual, desired string) bool {
	if actual == "" || desired == "" {
		return actual == desired
	}
	a, errA := semver.NewVersion(actual)
	d, errD := semver.NewVersion(desired)
	if errA != nil || errD != nil {
		return strings.TrimSpace(actual) == strings.TrimSpace(desired)
	}
	if versionHasPatch(desired) {
		return a.Equal(d)
	}
	return a.Major() == d.Major() && a.Minor() == d.Minor()
}

func versionHasPatch(v string) bool {
	return strings.Count(strings.TrimSpace(v), ".") >= 2
}

// normalizeVersionDir returns the canonical subdirectory name for an
// install of the given version: "foundryvtt_v<semver>" when parseable,
// "foundryvtt_v<raw>" otherwise.
func normalizeVersionDir(version string) string {
	if parsed, err := semver.NewVersion(version); err == nil {
		return "foundryvtt_v" + parsed.String()
	}
	return "foundryvtt_v" + strings.TrimSpace(version)
}
