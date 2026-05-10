package cmd

import (
	"context"
	"log/slog"
	"time"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundryruntime/health"
)

func Healthcheck(_ []string, logger *slog.Logger) int {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := health.Check(ctx, health.Default()); err != nil {
		logger.Error("healthcheck failed", "err", err)
		return 1
	}
	return 0
}
