package forge

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/fourcery/source"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/fourcery/version"
)

// VersionLatest is a sentinel value for the desired version that means
// "prefer the highest-version local artefact; fall back to remote sources
// only if nothing local is available."
const VersionLatest = "latest"

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
// docs/rules/sources.md. desired may be empty (no version constraint)
// or VersionLatest (local-first, remote as last resort).
func (r *Resolver) Resolve(
	ctx context.Context,
	desired string,
	candidates []Candidate,
	sources []source.Source,
) (Plan, error) {
	var rules []rule
	switch desired {
	case VersionLatest:
		rules = []rule{
			ruleHighestLocalSource(r),
			ruleLatestCandidate(),
			ruleFirstSourceOfKind(r, source.KindURL),
			ruleFirstSourceOfKind(r, source.KindSession),
		}
	case "":
		rules = []rule{
			ruleFirstSourceOfKind(r, source.KindURL),
			ruleHighestLocalSource(r),
			ruleLatestCandidate(),
		}
	default:
		desiredVer := version.Parse(desired)
		rules = []rule{
			ruleUseMatchingCandidate(desiredVer),
			ruleMatchingSource(r, desiredVer),
			ruleUnknownVersionSource(r, desiredVer),
		}
	}

	if plan, ok := runRules(ctx, candidates, sources, rules); ok {
		return plan, nil
	}
	return Plan{}, r.errNoMatch(desired)
}

func (r *Resolver) planInstall(s source.Source, desired version.Version) Plan {
	target := r.installRoot
	if !desired.IsZero() {
		target = filepath.Join(r.installRoot, desired.DirName())
	}
	return Plan{
		Action:          ActionInstallFromSource,
		Source:          s,
		TargetRoot:      target,
		ResolvedVersion: desired,
	}
}

func (r *Resolver) errNoMatch(desired string) error {
	switch desired {
	case "":
		return fmt.Errorf(
			"%w: no installed candidate, no source, and no version requested",
			source.ErrNoMatch,
		)
	case VersionLatest:
		return fmt.Errorf(
			"%w: no local source, no installed candidate, and no remote source for %q",
			source.ErrNoMatch, desired,
		)
	default:
		return fmt.Errorf("%w: no source matches version %q", source.ErrNoMatch, desired)
	}
}
