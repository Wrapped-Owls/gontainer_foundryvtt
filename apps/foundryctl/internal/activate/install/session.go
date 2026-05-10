package install

import (
	"context"
	"fmt"
	"log/slog"

	appconfig "github.com/wrapped-owls/gontainer_foundryvtt/apps/foundryctl/config"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundryacquire/auth"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundryacquire/release"
)

func canAuthenticate(cfg appconfig.Config) bool {
	return cfg.Install.Session != "" || (cfg.Install.Username != "" && authPassword(cfg) != "")
}

func acquireFromVersion(
	ctx context.Context,
	cfg appconfig.Config,
	targetRoot string,
	logger *slog.Logger,
) error {
	sess, err := loadSession(ctx, cfg)
	if err != nil {
		return fmt.Errorf("authenticate: %w", err)
	}
	url, err := release.Fetch(ctx, sess, cfg.Install.Version, release.FetchOptions{})
	if err != nil {
		return fmt.Errorf("fetch release URL: %w", err)
	}
	return acquireFromURL(ctx, targetRoot, url, logger)
}

func loadSession(ctx context.Context, cfg appconfig.Config) (*auth.Session, error) {
	if cfg.Install.Session != "" {
		sess, err := auth.LoadSession(
			cfg.Install.Session,
			auth.Options{UserAgent: auth.DefaultUserAgent},
		)
		if err == nil {
			return sess, nil
		}
		if cfg.Install.Username == "" || authPassword(cfg) == "" {
			return nil, err
		}
	}
	return auth.Login(
		ctx,
		cfg.Install.Username,
		authPassword(cfg),
		auth.Options{UserAgent: auth.DefaultUserAgent},
	)
}

func authPassword(cfg appconfig.Config) string {
	if cfg.Install.Password != "" {
		return cfg.Install.Password
	}
	return cfg.Admin.Key
}
