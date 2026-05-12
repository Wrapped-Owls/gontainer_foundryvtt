package forge

import (
	"context"
	"errors"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/fourcery/source"
)

// rule is a single resolution attempt. It returns (Plan, true) when it
// can satisfy the request, or (Plan{}, false) to pass to the next rule.
type rule func(ctx context.Context, candidates []Candidate, sources []source.Source) (Plan, bool)

// runRules executes rules in order and returns the first successful Plan.
func runRules(ctx context.Context, candidates []Candidate, sources []source.Source, rules []rule) (Plan, bool) {
	for _, r := range rules {
		if plan, ok := r(ctx, candidates, sources); ok {
			return plan, true
		}
	}
	return Plan{}, false
}

// ruleUseMatchingCandidate returns the first installed candidate that
// satisfies desired.
func ruleUseMatchingCandidate(desired string) rule {
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

// ruleMatchingSource returns the first source whose Probe equals desired.
func ruleMatchingSource(r *Resolver, desired string) rule {
	return func(ctx context.Context, _ []Candidate, sources []source.Source) (Plan, bool) {
		for _, s := range sources {
			v, err := s.Probe(ctx)
			if err != nil {
				continue
			}
			if versionsEqual(v, desired) {
				return r.planInstall(s, desired), true
			}
		}
		return Plan{}, false
	}
}

// ruleUnknownVersionSource returns the first source that cannot report
// its version without materialising (typically a presigned URL).
func ruleUnknownVersionSource(r *Resolver, desired string) rule {
	return func(ctx context.Context, _ []Candidate, sources []source.Source) (Plan, bool) {
		for _, s := range sources {
			if _, err := s.Probe(ctx); errors.Is(err, source.ErrVersionUnknown) {
				return r.planInstall(s, desired), true
			}
		}
		return Plan{}, false
	}
}

// ruleHighestLocalSource picks the zip or folder source with the
// highest semver version.
func ruleHighestLocalSource(r *Resolver) rule {
	return func(ctx context.Context, _ []Candidate, sources []source.Source) (Plan, bool) {
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
		if best == nil {
			return Plan{}, false
		}
		return r.planInstall(best, bestVer), true
	}
}

// ruleLatestCandidate picks the newest installed candidate. Candidates
// are pre-sorted newest-first by scanCandidates.
func ruleLatestCandidate() rule {
	return func(_ context.Context, candidates []Candidate, _ []source.Source) (Plan, bool) {
		if len(candidates) == 0 {
			return Plan{}, false
		}
		c := &candidates[0]
		return Plan{
			Action:          ActionUseExisting,
			Candidate:       c,
			ResolvedVersion: c.Version,
		}, true
	}
}

// ruleFirstSourceOfKind returns the first source of the given kind,
// installing without a version constraint (target = install root).
func ruleFirstSourceOfKind(r *Resolver, k source.Kind) rule {
	return func(_ context.Context, _ []Candidate, sources []source.Source) (Plan, bool) {
		for _, s := range sources {
			if s.Kind() == k {
				return r.planInstall(s, ""), true
			}
		}
		return Plan{}, false
	}
}
