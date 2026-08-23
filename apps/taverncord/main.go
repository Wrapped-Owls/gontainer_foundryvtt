package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/wrapped-owls/gontainer_foundryvtt/apps/taverncord/config"
	"github.com/wrapped-owls/gontainer_foundryvtt/apps/taverncord/internal/alerts"
	"github.com/wrapped-owls/gontainer_foundryvtt/apps/taverncord/internal/command"
	"github.com/wrapped-owls/gontainer_foundryvtt/apps/taverncord/internal/discordadapter"
	"github.com/wrapped-owls/gontainer_foundryvtt/apps/taverncord/internal/foundryclient"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrykit/colorlog"
)

const (
	exitUsage    = 1
	alertPollGap = 30 * time.Second
)

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

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	fc := foundryclient.New(cfg.Foundry.DashboardURL)
	cmds := command.New(fc, logger)

	router := discordadapter.NewRouter(ctx, "foundry", "Manage Foundry VTT profiles", logger).
		Use(cfg.Discord.GMRoleID).
		Add(discordadapter.ReadCommands(cmds)...).
		Add(discordadapter.ControlCommands(cmds)...)

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

	logger.Info("taverncord bot running - press Ctrl+C to stop")

	var wg sync.WaitGroup
	if cfg.Foundry.AlertChannelID != "" {
		poller := alerts.New(fc, adapter, cfg.Foundry.AlertChannelID, alertPollGap, logger)
		wg.Go(func() { poller.Run(ctx) })
		logger.Info("crash alert poller enabled", "channel", cfg.Foundry.AlertChannelID)
	}

	<-ctx.Done()
	wg.Wait()

	logger.Info("shutting down")
}
