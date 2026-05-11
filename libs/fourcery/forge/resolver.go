package forge

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/fourcery/source"
)

type Resolver struct {
	installRoot string
}

func NewResolver(installRoot string) *Resolver {
	return &Resolver{installRoot: installRoot}
}

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
		if s := firstMatchingSource(ctx, sources, desired); s != nil {
			return r.planInstall(s, desired), nil
		}
		if s := firstUnknownVersionSource(ctx, sources); s != nil {
			return r.planInstall(s, desired), nil
		}
		return Plan{}, fmt.Errorf(
			"%w: no source matches version %q",
			source.ErrNoMatch, desired,
		)
	}

	if s := firstSourceOfKind(sources, source.KindURL); s != nil {
		return r.planInstall(s, ""), nil
	}
	if s := highestVersionLocalSource(ctx, sources); s != nil {
		v, _ := s.Probe(ctx)
		return r.planInstall(s, v), nil
	}
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
