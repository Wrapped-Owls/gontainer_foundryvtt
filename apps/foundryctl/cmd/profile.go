package cmd

import (
	"context"
	"log/slog"
	"slices"

	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundryctl/internal/activate"
	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/procloop"
	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/profile"
)

// profilePreparer re-runs the activation pipeline with a profile's overrides applied.
// It is a parameter so the resolution order can be tested without a real install tree.
type profilePreparer func(
	context.Context,
	*slog.Logger,
	activate.State,
	profile.Profile,
) (activate.State, error)

// resolveInitialProfile tries last-active, then the configured default, then the
// first configured one. Activating none is a deliberate last resort, not a fallback.
func resolveInitialProfile(
	ctx context.Context,
	logger *slog.Logger,
	state activate.State,
	prepare profilePreparer,
) (procloop.State, string) {
	for _, name := range profileCandidates(state) {
		target, found := profile.ByName(state.Profiles, name)
		if !found {
			logger.Warn("configured profile not found", "profile", name)
			continue
		}
		activated, err := prepare(ctx, logger, state, target)
		if err != nil {
			logger.Warn("failed to activate profile", "profile", name, "err", err)
			continue
		}
		logger.Info("active profile resolved", "profile", name)
		return toProcloopState(activated), name
	}

	if len(state.Profiles) > 0 {
		logger.Error("no configured profile could be activated; starting with base config")
	}
	return toProcloopState(state), ""
}

func profileCandidates(state activate.State) []string {
	var names []string
	for _, name := range []string{state.ActiveProfile, state.App.Manager.DefaultProfile} {
		if name != "" && !slices.Contains(names, name) {
			names = append(names, name)
		}
	}
	if len(state.Profiles) > 0 && !slices.Contains(names, state.Profiles[0].Name) {
		names = append(names, state.Profiles[0].Name)
	}
	return names
}
