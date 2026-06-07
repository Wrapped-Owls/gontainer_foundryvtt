// Package cmd contains the foundrymanager subcommand implementations.
package cmd

import (
	"context"
	"log/slog"
	"os/signal"
	"syscall"

	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/config"
	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/procloop"
	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/profile"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrykit/backoff"
)

// Run loads config and starts the dashboard in standalone mode (no process
// management). Useful for testing the dashboard independently.
func Run(_ []string, logger *slog.Logger) int {
	cfg, err := config.Load()
	if err != nil {
		logger.Error("failed to load config", "err", err)
		return 1
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	mgr := procloop.New(
		procloop.State{},
		"",
		&noopActivator{},
		cfg,
		backoff.Config{},
		logger,
	)
	return mgr.Run(ctx)
}

// noopActivator is used in standalone mode where no Foundry process is managed.
type noopActivator struct{}

func (noopActivator) Switch(
	_ context.Context,
	_ *slog.Logger,
	_ profile.Profile,
) (procloop.State, error) {
	return procloop.State{}, nil
}
