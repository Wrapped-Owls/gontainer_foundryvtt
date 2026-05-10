// Package activate prepares Foundry VTT for launch by running the activation sequence.
package activate

import (
	"context"
	"log/slog"

	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundryctl/internal/activate/step"
)

// State is the resolved runtime state produced by Prepare.
type State = step.State

// Prepare runs the full activation sequence and returns the resolved State.
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
