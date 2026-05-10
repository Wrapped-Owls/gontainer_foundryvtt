package install

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundryruntime/lifecycle"
)

type installCandidate struct {
	Path    string
	Info    lifecycle.InstalledInfo
	Parsed  *semver.Version
	Version string
}

func scanInstallCandidates(installRoot string) ([]installCandidate, error) {
	var out []installCandidate
	rootInfo, err := lifecycle.DetectInstalled(installRoot)
	if err != nil {
		return nil, fmt.Errorf("detect install: %w", err)
	}
	if rootInfo.Present {
		out = append(out, newCandidate(installRoot, rootInfo))
	}
	entries, err := os.ReadDir(installRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return out, nil
		}
		return nil, fmt.Errorf("read install root: %w", err)
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		child := filepath.Join(installRoot, entry.Name())
		info, derr := lifecycle.DetectInstalled(child)
		if derr != nil {
			return nil, fmt.Errorf("detect install %s: %w", child, derr)
		}
		if info.Present {
			out = append(out, newCandidate(child, info))
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Parsed == nil {
			return false
		}
		if out[j].Parsed == nil {
			return true
		}
		return out[i].Parsed.GreaterThan(out[j].Parsed)
	})
	return out, nil
}

func newCandidate(path string, info lifecycle.InstalledInfo) installCandidate {
	c := installCandidate{Path: path, Info: info, Version: info.Version}
	if v, err := semver.NewVersion(info.Version); err == nil {
		c.Parsed = v
		c.Version = v.String()
	}
	return c
}

func matchCandidate(candidates []installCandidate, desired string) *installCandidate {
	parsed, err := semver.NewVersion(desired)
	if err != nil {
		for i := range candidates {
			if candidates[i].Info.Version == desired {
				return &candidates[i]
			}
		}
		return nil
	}
	requirePatch := versionHasPatch(desired)
	for i := range candidates {
		c := &candidates[i]
		if c.Parsed == nil {
			if c.Info.Version == desired {
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

func latestCandidate(candidates []installCandidate) *installCandidate {
	if len(candidates) == 0 {
		return nil
	}
	return &candidates[0]
}

func versionMatches(actual, desired string) bool {
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
