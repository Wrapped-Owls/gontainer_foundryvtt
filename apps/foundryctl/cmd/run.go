package cmd

import (
	"context"
	"log/slog"
	"syscall"

	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundryctl/internal/activate"
	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundryctl/internal/versions"
	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/procloop"
	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/profile"
)

func Run(_ []string, logger *slog.Logger) int {
	ctx, cancel := context.WithCancelCause(context.Background())
	defer cancel(nil)
	ctx, stop := signalNotifyContext(ctx, syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	state, err := activate.Prepare(ctx, logger)
	if err != nil {
		logger.Error("activation failed", "err", err, "install_root", state.App.Paths.InstallRoot)
		return 1
	}

	healthErr := startHealthServer(ctx, logger, state.App.Paths.HealthAddr, state.Runtime.Port)
	go func() {
		if err := <-healthErr; err != nil {
			cancel(err)
		}
	}()

	initialState, initialActive := resolveInitialProfile(
		ctx,
		logger,
		state,
		activate.PrepareProfile,
	)
	mgr := procloop.New(procloop.Params{
		Initial:       initialState,
		InitialActive: initialActive,
		Activator:     &appActivator{base: state, logger: logger},
		Versions:      versions.New(state.App.Paths, state.App.Install, logger),
		Config:        state.App.Manager,
		Backoff:       state.App.Backoff,
		Logger:        logger,
	})
	return mgr.Run(ctx)
}

func toProcloopState(s activate.State) procloop.State {
	return procloop.State{
		DataPath:    s.App.Paths.DataPath,
		InstallRoot: s.Install.Root,
		MainScript:  s.App.Paths.MainScript,
		JSRuntime:   s.JSRuntime,
		Port:        s.Runtime.Port,
		Version:     s.Install.Version.String(),
		World:       s.Runtime.World,
		Profiles:    s.Profiles,
	}
}

type appActivator struct {
	base   activate.State
	logger *slog.Logger
}

func (a *appActivator) Switch(
	ctx context.Context,
	logger *slog.Logger,
	p profile.Profile,
) (procloop.State, error) {
	newState, err := activate.PrepareProfile(ctx, logger, a.base, p)
	if err != nil {
		return procloop.State{}, err
	}
	return toProcloopState(newState), nil
}
