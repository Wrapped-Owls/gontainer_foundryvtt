package step

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundryruntime/lifecycle"
)

type optionsStep struct{}

// Options returns a Step that writes options.json and admin.txt to the data path.
func Options() Step { return optionsStep{} }

func (optionsStep) Apply(_ context.Context, s *State, _ *slog.Logger) error {
	if _, err := lifecycle.WriteOptions(s.App.Paths.DataPath, s.Runtime); err != nil {
		return fmt.Errorf("write options: %w", err)
	}
	if _, err := lifecycle.WriteAdminPassword(s.App.Paths.DataPath, s.App.Admin.Key, s.App.Admin.PasswordSalt); err != nil {
		return fmt.Errorf("write admin.txt: %w", err)
	}
	return nil
}
