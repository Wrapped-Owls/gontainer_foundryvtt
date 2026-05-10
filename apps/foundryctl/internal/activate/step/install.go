package step

import (
	"context"
	"log/slog"

	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundryctl/internal/activate/install"
)

type installStep struct{}

func Install() Step { return installStep{} }

func (installStep) Apply(ctx context.Context, s *State, logger *slog.Logger) error {
	inst, err := install.EnsureInstall(ctx, s.App, logger)
	if err != nil {
		return err
	}
	s.Install = inst
	return nil
}
