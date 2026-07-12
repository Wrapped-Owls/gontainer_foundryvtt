package step

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundryruntime/lifecycle"
)

type licenseStep struct{}

// License returns a Step that syncs the per-version license cache so a license
// already accepted for the target Foundry version is restored instead of being
// prompted for again after a version switch.
func License() Step { return licenseStep{} }

func (licenseStep) Apply(_ context.Context, s *State, _ *slog.Logger) error {
	if err := lifecycle.SyncLicense(
		s.App.Paths.DataPath,
		s.Install.Version.String(),
		s.App.Paths.LicenseCache,
	); err != nil {
		return fmt.Errorf("sync license: %w", err)
	}
	return nil
}
