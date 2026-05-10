package activate

import (
	"context"
	"log/slog"

	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundryctl/internal/activate/step"
)

type State = step.State

func Prepare(ctx context.Context, logger *slog.Logger) (State, error) {
	return step.Run(ctx, logger,
		step.AppConfig(),
		step.Secrets(),
		step.Install(),
		step.Options(),
		step.Patches(),
		step.JSRuntime(),
	)
}
