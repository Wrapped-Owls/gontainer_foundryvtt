// Package main is the composition root for the taverncord bot.
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/wrapped-owls/gontainer_foundryvtt/apps/taverncord/config"
	"github.com/wrapped-owls/gontainer_foundryvtt/apps/taverncord/internal/command"
	"github.com/wrapped-owls/gontainer_foundryvtt/apps/taverncord/internal/discordadapter"
	"github.com/wrapped-owls/gontainer_foundryvtt/apps/taverncord/internal/foundryclient"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrykit/colorlog"
)

const exitUsage = 1

func main() {
	logger := colorlog.New("taverncord", colorlog.LevelFromEnv())

	cfg, err := config.Load()
	if err != nil {
		logger.Error("failed to load config", "err", err)
		os.Exit(exitUsage)
	}
	if cfg.Discord.Token == "" || cfg.Discord.ApplicationID == "" {
		logger.Error("DISCORD_TOKEN and DISCORD_APPLICATION_ID are required")
		os.Exit(exitUsage)
	}

	fc := foundryclient.New(cfg.Foundry.DashboardURL)
	cmds := command.New(fc, logger)

	router := discordadapter.NewRouter("foundry", "Manage Foundry VTT profiles", logger).
		Use(cfg.Discord.GMRoleID).
		Add(discordadapter.ListCmd(cmds)).
		Add(discordadapter.SwitchCmd(cmds)).
		Add(discordadapter.StatusCmd(cmds))

	adapter, err := discordadapter.New(cfg, router, logger)
	if err != nil {
		logger.Error("failed to create Discord adapter", "err", err)
		os.Exit(exitUsage)
	}

	if err = adapter.Open(); err != nil {
		logger.Error("failed to open Discord session", "err", err)
		os.Exit(exitUsage)
	}
	defer func() { _ = adapter.Close() }()

	logger.Info("taverncord bot running — press Ctrl+C to stop")

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()
	<-ctx.Done()

	logger.Info("shutting down")
}
