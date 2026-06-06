// Package step provides the strategy pattern for the activate sequence.
// Each Step implementation handles one phase of Foundry startup preparation.
package step

import (
	"context"
	"log/slog"

	appconfig "github.com/wrapped-owls/gontainer_foundryvtt/apps/foundryctl/config"
	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/profile"
	runtimecfg "github.com/wrapped-owls/gontainer_foundryvtt/libs/foundryruntime/config"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundryruntime/jsruntime"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/fourcery/forge"
)

// State accumulates the result of each preparation step.
type State struct {
	App       appconfig.Config
	Runtime   runtimecfg.Config
	JSRuntime jsruntime.Runtime
	Install   forge.Install
	Profiles  []profile.Profile
}

// Step is the strategy interface for one phase of the activation sequence.
type Step interface {
	Apply(ctx context.Context, s *State, logger *slog.Logger) error
}

// Run executes steps in order, returning the final State or the first error.
func Run(ctx context.Context, logger *slog.Logger, steps ...Step) (State, error) {
	return RunFrom(ctx, logger, State{}, steps...)
}

// RunFrom executes steps starting from an existing state, returning the final
// State or the first error.
func RunFrom(
	ctx context.Context, logger *slog.Logger, initial State, steps ...Step,
) (State, error) {
	s := initial
	for _, step := range steps {
		if err := step.Apply(ctx, &s, logger); err != nil {
			return State{}, err
		}
	}
	return s, nil
}
