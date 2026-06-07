package cmd

import (
	"context"
	"log/slog"
	"syscall"

	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundryctl/internal/activate"
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
	logger.Info("js runtime selected", "kind", state.JSRuntime.Kind, "path", state.JSRuntime.Path)

	initialState, initialActive := resolveInitialProfile(ctx, logger, state)
	mgr := procloop.New(
		initialState,
		initialActive,
		&appActivator{base: state, logger: logger},
		state.App.Manager,
		state.App.Backoff,
		logger,
	)
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
		Profiles:    s.Profiles,
	}
}

// appActivator implements procloop.Activator using the app-layer activate pipeline.
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

// resolveInitialProfile returns the initial procloop state and active profile
// name. If a last-active profile is recorded and found in the profile list, it
// activates that profile so the process starts in the correct session.
func resolveInitialProfile(
	ctx context.Context,
	logger *slog.Logger,
	state activate.State,
) (procloop.State, string) {
	if state.ActiveProfile == "" {
		return toProcloopState(state), ""
	}
	var target profile.Profile
	found := false
	for _, p := range state.Profiles {
		if p.Name == state.ActiveProfile {
			target = p
			found = true
			break
		}
	}
	if !found {
		logger.Warn(
			"last active profile not found, starting with base config",
			"profile",
			state.ActiveProfile,
		)
		return toProcloopState(state), ""
	}
	activated, err := activate.PrepareProfile(ctx, logger, state, target)
	if err != nil {
		logger.Warn(
			"failed to activate last profile, starting with base config",
			"profile",
			state.ActiveProfile,
			"err",
			err,
		)
		return toProcloopState(state), ""
	}
	logger.Info("resuming last active profile", "profile", state.ActiveProfile)
	return toProcloopState(activated), state.ActiveProfile
}
