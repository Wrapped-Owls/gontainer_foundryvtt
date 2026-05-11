package forge

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/fourcery/internal/probe"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/fourcery/source"
)

const dirPerm fs.FileMode = 0o755

// Forge is the orchestrator: scan candidates, resolve a Plan, then
// materialise the chosen source into a versioned subdirectory.
// Construct via New(...).Build(); never zero-value.
type Forge struct {
	installRoot string
	sources     []source.Source
	observer    Observer
	resolver    *Resolver
}

// Sources returns the configured source list (read-only view).
func (f *Forge) Sources() []source.Source { return f.sources }

// Resolve scans the install root, asks the resolver to pick a plan,
// and emits EventResolved. desired may be empty.
func (f *Forge) Resolve(ctx context.Context, desired string) (Plan, error) {
	candidates, err := scanCandidates(f.installRoot)
	if err != nil {
		return Plan{}, err
	}
	plan, err := f.resolver.Resolve(ctx, desired, candidates, f.sources)
	if err != nil {
		return Plan{}, err
	}
	f.observer.Notify(EventResolved{Plan: plan})
	return plan, nil
}

// Acquire executes the plan: reuses an existing candidate or stages
// the chosen source into a versioned subdirectory.
func (f *Forge) Acquire(ctx context.Context, p Plan) (Install, error) {
	switch p.Action {
	case ActionUseExisting:
		if p.Candidate == nil {
			return Install{}, errors.New("forge: use-existing plan has no candidate")
		}
		inst := Install{Root: p.Candidate.Path, Version: p.Candidate.Version}
		f.observer.Notify(EventSkipped{Reason: "candidate already installed", Install: inst})
		return inst, nil
	case ActionInstallFromSource:
		return f.materialise(ctx, p)
	default:
		return Install{}, fmt.Errorf("forge: unknown action %d", p.Action)
	}
}

func (f *Forge) materialise(ctx context.Context, p Plan) (Install, error) {
	if p.Source == nil {
		return Install{}, errors.New("forge: install-from-source plan has no source")
	}
	if err := os.MkdirAll(f.installRoot, dirPerm); err != nil {
		return Install{}, fmt.Errorf("forge: mkdir install root: %w", err)
	}
	staging, err := os.MkdirTemp(f.installRoot, ".fourcery-staging-*")
	if err != nil {
		return Install{}, fmt.Errorf("forge: stage temp dir: %w", err)
	}
	cleanedStaging := false
	defer func() {
		if !cleanedStaging {
			_ = os.RemoveAll(staging)
		}
	}()
	f.observer.Notify(EventInstalling{Source: p.Source, Target: p.TargetRoot})

	res, err := p.Source.Materialise(ctx, staging)
	if err != nil {
		return Install{}, fmt.Errorf("forge: materialise: %w", err)
	}
	version := res.Version
	if version == "" {
		// Source didn't report; introspect the staged tree.
		if v, perr := probe.Folder(staging); perr == nil {
			version = v
		}
	}
	target := p.TargetRoot
	if target == "" {
		if version == "" {
			return Install{}, errors.New(
				"forge: install completed but version unknown and no target specified",
			)
		}
		target = filepath.Join(f.installRoot, normalizeVersionDir(version))
	}
	if err = swapInto(staging, target); err != nil {
		return Install{}, err
	}
	cleanedStaging = true
	inst := Install{Root: target, Version: version}
	f.observer.Notify(EventInstalled{Install: inst})
	return inst, nil
}

// swapInto atomically replaces target with staging via rename. If
// target exists it is removed first.
func swapInto(staging, target string) error {
	if err := os.RemoveAll(target); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("forge: remove existing %s: %w", target, err)
	}
	if err := os.MkdirAll(filepath.Dir(target), dirPerm); err != nil {
		return fmt.Errorf("forge: mkdir parent of %s: %w", target, err)
	}
	if err := os.Rename(staging, target); err != nil {
		return fmt.Errorf("forge: rename %s -> %s: %w", staging, target, err)
	}
	return nil
}
