package forge

import (
	"context"
	"errors"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/fourcery/source"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/fourcery/version"
)

type rule func(ctx context.Context, candidates []Candidate, sources []source.Source) (Plan, bool)

func runRules(
	ctx context.Context,
	candidates []Candidate,
	sources []source.Source,
	rules []rule,
) (Plan, bool) {
	for _, r := range rules {
		if plan, ok := r(ctx, candidates, sources); ok {
			return plan, true
		}
	}
	return Plan{}, false
}

func ruleUseMatchingCandidate(desired version.Version) rule {
	return func(_ context.Context, candidates []Candidate, _ []source.Source) (Plan, bool) {
		match := matchCandidate(candidates, desired)
		if match == nil {
			return Plan{}, false
		}
		return Plan{
			Action:          ActionUseExisting,
			Candidate:       match,
			ResolvedVersion: match.Version,
		}, true
	}
}

func ruleMatchingSource(r *Resolver, desired version.Version) rule {
	return func(ctx context.Context, _ []Candidate, sources []source.Source) (Plan, bool) {
		for _, s := range sources {
			v, err := s.Probe(ctx)
			if err != nil {
				continue
			}
			if v.Matches(desired) {
				return r.planInstall(s, desired), true
			}
		}
		return Plan{}, false
	}
}

func ruleUnknownVersionSource(r *Resolver, desired version.Version) rule {
	return func(ctx context.Context, _ []Candidate, sources []source.Source) (Plan, bool) {
		for _, s := range sources {
			if _, err := s.Probe(ctx); errors.Is(err, source.ErrVersionUnknown) {
				return r.planInstall(s, desired), true
			}
		}
		return Plan{}, false
	}
}

func ruleHighestLocalSource(r *Resolver) rule {
	return func(ctx context.Context, _ []Candidate, sources []source.Source) (Plan, bool) {
		var best source.Source
		var bestVer version.Version
		for _, s := range sources {
			if s.Kind() != source.KindZip && s.Kind() != source.KindFolder {
				continue
			}
			v, err := s.Probe(ctx)
			if err != nil {
				continue
			}
			if best == nil || v.Compare(bestVer) > 0 {
				best = s
				bestVer = v
			}
		}
		if best == nil {
			return Plan{}, false
		}
		return r.planInstall(best, bestVer), true
	}
}

func ruleLatestCandidate() rule {
	return func(_ context.Context, candidates []Candidate, _ []source.Source) (Plan, bool) {
		if len(candidates) == 0 {
			return Plan{}, false
		}
		c := &candidates[0] // candidates arrive pre-sorted newest-first
		return Plan{
			Action:          ActionUseExisting,
			Candidate:       c,
			ResolvedVersion: c.Version,
		}, true
	}
}

func ruleFirstSourceOfKind(r *Resolver, k source.Kind) rule {
	return func(_ context.Context, _ []Candidate, sources []source.Source) (Plan, bool) {
		for _, s := range sources {
			if s.Kind() == k {
				return r.planInstall(s, version.Version{}), true
			}
		}
		return Plan{}, false
	}
}
