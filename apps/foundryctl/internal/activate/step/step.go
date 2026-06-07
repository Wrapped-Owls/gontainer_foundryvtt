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

type State struct {
	App           appconfig.Config
	Runtime       runtimecfg.Config
	JSRuntime     jsruntime.Runtime
	Install       forge.Install
	Profiles      []profile.Profile
	ActiveProfile string
}

type Step interface {
	Apply(ctx context.Context, s *State, logger *slog.Logger) error
}

func Run(ctx context.Context, logger *slog.Logger, steps ...Step) (State, error) {
	return RunFrom(ctx, logger, State{}, steps...)
}

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
