package forge

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/fourcery/internal/probe"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/fourcery/version"
)

type Candidate struct {
	Path string

	Version version.Version
}

func newCandidate(path, ver string) Candidate {
	return Candidate{Path: path, Version: version.Parse(ver)}
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
		return b.Version.Compare(a.Version) // newest first; non-semver sorts last
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
	ver, err := probe.Folder(path)
	if err != nil && !errors.Is(err, probe.ErrNoVersion) {
		return Candidate{}, false, fmt.Errorf("forge: probe %s: %w", path, err)
	}
	return newCandidate(path, ver), true, nil
}

func matchCandidate(candidates []Candidate, desired version.Version) *Candidate {
	for i := range candidates {
		if candidates[i].Version.Matches(desired) {
			return &candidates[i]
		}
	}
	return nil
}
