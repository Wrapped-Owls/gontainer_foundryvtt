package forge

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/fourcery/source"
)

// Resolver decides what to do with a list of installed candidates and
// enumerated sources, given an optional desired version. The decision
// is pure with respect to the inputs (Probe calls aside).
type Resolver struct {
	installRoot string
}

// NewResolver constructs a Resolver rooted at installRoot.
func NewResolver(installRoot string) *Resolver {
	return &Resolver{installRoot: installRoot}
}

// Resolve picks the plan according to the rules documented in
// docs/rules/sources.md. desired may be empty (no version constraint).
func (r *Resolver) Resolve(
	ctx context.Context,
	desired string,
	candidates []Candidate,
	sources []source.Source,
) (Plan, error) {
	if desired != "" {
		if match := matchCandidate(candidates, desired); match != nil {
			return Plan{
				Action:          ActionUseExisting,
				Candidate:       match,
				ResolvedVersion: match.Version,
			}, nil
		}
		// 2: prefer a source whose Probe matches desired.
		if s := firstMatchingSource(ctx, sources, desired); s != nil {
			return r.planInstall(s, desired), nil
		}
		// 3: fall back to a source whose version is unknowable
		// without materialising (URL with no label).
		if s := firstUnknownVersionSource(ctx, sources); s != nil {
			return r.planInstall(s, desired), nil
		}
		return Plan{}, fmt.Errorf(
			"%w: no source matches version %q",
			source.ErrNoMatch, desired,
		)
	}

	// desired unset: URL is the operator's explicit intent.
	if s := firstSourceOfKind(sources, source.KindURL); s != nil {
		return r.planInstall(s, ""), nil
	}
	// 5: highest-version local source.
	if s := highestVersionLocalSource(ctx, sources); s != nil {
		v, _ := s.Probe(ctx)
		return r.planInstall(s, v), nil
	}
	// 6: latest installed candidate.
	if len(candidates) > 0 {
		c := &candidates[0]
		return Plan{
			Action:          ActionUseExisting,
			Candidate:       c,
			ResolvedVersion: c.Version,
		}, nil
	}
	return Plan{}, fmt.Errorf(
		"%w: no installed candidate, no source, and no version requested",
		source.ErrNoMatch,
	)
}

func (r *Resolver) planInstall(s source.Source, desired string) Plan {
	target := r.installRoot
	if desired != "" {
		target = filepath.Join(r.installRoot, normalizeVersionDir(desired))
	}
	return Plan{
		Action:          ActionInstallFromSource,
		Source:          s,
		TargetRoot:      target,
		ResolvedVersion: desired,
	}
}

// firstMatchingSource returns the first source whose Probe equals
// desired (via versionsEqual). Sources whose Probe errors with
// ErrVersionUnknown are skipped.
func firstMatchingSource(
	ctx context.Context,
	sources []source.Source,
	desired string,
) source.Source {
	for _, s := range sources {
		v, err := s.Probe(ctx)
		if err != nil {
			continue
		}
		if versionsEqual(v, desired) {
			return s
		}
	}
	return nil
}

func firstUnknownVersionSource(ctx context.Context, sources []source.Source) source.Source {
	for _, s := range sources {
		if _, err := s.Probe(ctx); errors.Is(err, source.ErrVersionUnknown) {
			return s
		}
	}
	return nil
}

func firstSourceOfKind(sources []source.Source, k source.Kind) source.Source {
	for _, s := range sources {
		if s.Kind() == k {
			return s
		}
	}
	return nil
}

func highestVersionLocalSource(ctx context.Context, sources []source.Source) source.Source {
	var best source.Source
	var bestVer string
	for _, s := range sources {
		if s.Kind() != source.KindZip && s.Kind() != source.KindFolder {
			continue
		}
		v, err := s.Probe(ctx)
		if err != nil {
			continue
		}
		if best == nil || compareSemver(v, bestVer) > 0 {
			best = s
			bestVer = v
		}
	}
	return best
}
