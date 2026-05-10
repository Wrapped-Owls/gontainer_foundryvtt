package step

import (
	"context"
	"log/slog"

	appconfig "github.com/wrapped-owls/gontainer_foundryvtt/apps/foundryctl/config"
	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundryctl/internal/activate/install"
	runtimecfg "github.com/wrapped-owls/gontainer_foundryvtt/libs/foundryruntime/config"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundryruntime/jsruntime"
)

type State struct {
	App       appconfig.Config
	Runtime   runtimecfg.Config
	JSRuntime jsruntime.Runtime
	Install   install.Install
}

type Step interface {
	Apply(ctx context.Context, s *State, logger *slog.Logger) error
}

func Run(ctx context.Context, logger *slog.Logger, steps ...Step) (State, error) {
	var s State
	for _, step := range steps {
		if err := step.Apply(ctx, &s, logger); err != nil {
			return State{}, err
		}
	}
	return s, nil
}
