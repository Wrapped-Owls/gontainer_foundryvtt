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

type Candidate struct {
	Path    string
	Version string
	Parsed  *semver.Version
}

func newCandidate(path, version string) Candidate {
	c := Candidate{Path: path, Version: version}
	if v, err := semver.NewVersion(version); err == nil {
		c.Parsed = v
		c.Version = v.String()
	}
	return c
}

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
			return 1
		}
		if b.Parsed == nil {
			return -1
		}
		return b.Parsed.Compare(a.Parsed)
	})
	return out, nil
}

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

func normalizeVersionDir(version string) string {
	if parsed, err := semver.NewVersion(version); err == nil {
		return "foundryvtt_v" + parsed.String()
	}
	return "foundryvtt_v" + strings.TrimSpace(version)
}
