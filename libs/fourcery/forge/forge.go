package forge

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrykit/fsperm"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/fourcery/internal/probe"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/fourcery/source"
)

type Forge struct {
	installRoot string
	sources     []source.Source
	observer    Observer
	resolver    *Resolver
}

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
	if err := os.MkdirAll(f.installRoot, fsperm.Dir); err != nil {
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

func swapInto(staging, target string) error {
	if err := os.RemoveAll(target); err != nil &&
		!errors.Is(err, fs.ErrNotExist) { // clear target first so the rename lands atomically
		return fmt.Errorf("forge: remove existing %s: %w", target, err)
	}
	if err := os.MkdirAll(filepath.Dir(target), fsperm.Dir); err != nil {
		return fmt.Errorf("forge: mkdir parent of %s: %w", target, err)
	}
	if err := os.Rename(staging, target); err != nil {
		return fmt.Errorf("forge: rename %s -> %s: %w", staging, target, err)
	}
	return nil
}
