package step

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/profloader"
)

type profilesStep struct{}

func Profiles() Step { return profilesStep{} }

func (profilesStep) Apply(_ context.Context, s *State, logger *slog.Logger) error {
	profiles, active, err := profloader.Load(s.App.Manager.ProfilesFile, "FOUNDRY_PROFILE")
	if err != nil {
		return fmt.Errorf("step profiles: %w", err)
	}
	s.Profiles = profiles
	s.ActiveProfile = active
	logger.Debug("profiles loaded", "count", len(profiles), "active", active)
	return nil
}
