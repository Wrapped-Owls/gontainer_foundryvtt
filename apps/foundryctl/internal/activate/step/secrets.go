package step

import (
	"context"
	"log/slog"

	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundryctl/internal/secfuse"
)

type secretsStep struct{}

// Secrets returns a Step that loads the secret-fuse file and injects known keys as env vars.
// Errors are logged but do not abort the sequence.
func Secrets() Step { return secretsStep{} }

func (secretsStep) Apply(_ context.Context, s *State, logger *slog.Logger) error {
	res, err := secfuse.Load(s.App.Secrets.Path)
	if err != nil {
		logger.Error("secret load failed", "err", err)
		return nil
	}
	if len(res.Applied) > 0 {
		logger.Info("secrets applied", "vars", res.Applied, "source", res.SourcePath)
	}
	return nil
}
