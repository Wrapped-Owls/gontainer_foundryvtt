package cmd

import (
	"context"
	"errors"
	"log/slog"
	"os/signal"
	"syscall"

	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/config"
	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/procloop"
	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/profile"
)

func Run(_ []string, logger *slog.Logger) int {
	cfg, err := config.Load()
	if err != nil {
		logger.Error("failed to load config", "err", err)
		return 1
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	mgr := procloop.New(procloop.Params{
		Activator: &noopActivator{},
		Versions:  noopVersions{},
		Config:    cfg,
		Logger:    logger,
	})
	return mgr.Run(ctx)
}

type noopVersions struct{}

func (noopVersions) Installed(_ context.Context) ([]string, error) { return nil, nil }

func (noopVersions) Download(_ context.Context, _, _ string) error {
	return errors.New("version download is unavailable in standalone mode")
}

type noopActivator struct{}

func (noopActivator) Switch(
	_ context.Context,
	_ *slog.Logger,
	_ profile.Profile,
) (procloop.State, error) {
	return procloop.State{}, nil
}
