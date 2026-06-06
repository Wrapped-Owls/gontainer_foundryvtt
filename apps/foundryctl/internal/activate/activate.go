// Package activate prepares Foundry VTT for launch by running the activation sequence.
package activate

import (
	"context"
	"log/slog"

	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/profile"

	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundryctl/internal/activate/step"
)

// State is the resolved runtime state produced by Prepare.
type State = step.State

// Prepare runs the full activation sequence and returns the resolved State.
func Prepare(ctx context.Context, logger *slog.Logger) (State, error) {
	return step.Run(
		ctx, logger,
		step.AppConfig(),
		step.EnsureDirs(),
		step.Secrets(),
		step.Install(),
		step.Options(),
		step.Patches(),
		step.JSRuntime(),
		step.Profiles(),
	)
}

// PrepareProfile re-activates using profile overrides applied on top of the
// base state. Only the steps relevant to the changed fields are re-run.
func PrepareProfile(
	ctx context.Context,
	logger *slog.Logger,
	base State,
	p profile.Profile,
) (State, error) {
	s := base
	if p.DataPath != "" {
		s.App.Paths.DataPath = p.DataPath
	}
	if p.AdminKey != "" {
		s.App.Admin.Key = p.AdminKey
	}
	if p.AdminPasswordSalt != "" {
		s.App.Admin.PasswordSalt = p.AdminPasswordSalt
	}
	if p.ManifestPath != "" {
		s.App.Paths.ManifestPath = p.ManifestPath
	}

	versionChanged := p.Version != "" && p.Version != base.Install.Version.String()
	if versionChanged {
		s.App.Install.Version = p.Version
		return step.RunFrom(
			ctx, logger, s,
			step.EnsureDirs(),
			step.Install(),
			step.Options(),
			step.Patches(),
		)
	}
	return step.RunFrom(
		ctx, logger, s,
		step.EnsureDirs(),
		step.Options(),
	)
}
