package cmd

import (
	"log/slog"

	appconfig "github.com/wrapped-owls/gontainer_foundryvtt/apps/foundryctl/config"
	runtimecfg "github.com/wrapped-owls/gontainer_foundryvtt/libs/foundryruntime/config"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundryruntime/lifecycle"
)

func Options(_ []string, logger *slog.Logger) int {
	cfg, err := appconfig.Load()
	if err != nil {
		logger.Error("config load failed", "err", err)
		return 1
	}
	rt := runtimecfg.Default()
	if err = runtimecfg.LoadFromEnv(&rt); err != nil {
		logger.Error("build options failed", "err", err)
		return 1
	}
	rt.DataPath = cfg.Paths.DataPath
	rt.Port = cfg.Runtime.Port
	if _, err = lifecycle.WriteOptions(cfg.Paths.DataPath, rt); err != nil {
		logger.Error("write options failed", "err", err)
		return 1
	}
	return 0
}
