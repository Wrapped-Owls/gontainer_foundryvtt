package forge

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/fourcery/source"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/fourcery/version"
)

const VersionLatest = "latest" // prefer highest local artefact, else fall back to remote sources

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
