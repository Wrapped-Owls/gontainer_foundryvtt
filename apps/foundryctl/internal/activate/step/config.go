package step

import (
	"context"
	"fmt"
	"log/slog"

	appconfig "github.com/wrapped-owls/gontainer_foundryvtt/apps/foundryctl/config"
	runtimecfg "github.com/wrapped-owls/gontainer_foundryvtt/libs/foundryruntime/config"
)

type appConfigStep struct{}

// AppConfig returns a Step that loads the app config and overlays runtime config defaults.
func AppConfig() Step { return appConfigStep{} }

func (appConfigStep) Apply(_ context.Context, s *State, _ *slog.Logger) error {
	cfg, err := appconfig.Load()
	if err != nil {
		return fmt.Errorf("config load: %w", err)
	}
	s.App = cfg

	rt := runtimecfg.Default()
	if err = runtimecfg.LoadFromEnv(&rt); err != nil {
		return fmt.Errorf("build options: %w", err)
	}
	rt.DataPath = cfg.Paths.DataPath
	rt.Port = cfg.Runtime.Port
	s.Runtime = rt
	return nil
}
