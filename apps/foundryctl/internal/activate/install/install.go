package install

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"

	appconfig "github.com/wrapped-owls/gontainer_foundryvtt/apps/foundryctl/config"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundryruntime/lifecycle"
)

// Install describes a resolved Foundry installation.
type Install struct {
	Root string
	Info lifecycle.InstalledInfo
}

// EnsureInstall resolves or acquires a Foundry installation from cfg.
func EnsureInstall(
	ctx context.Context,
	cfg appconfig.Config,
	logger *slog.Logger,
) (Install, error) {
	candidates, err := scanInstallCandidates(cfg.Paths.InstallRoot)
	if err != nil {
		return Install{}, err
	}

	if cfg.Install.Version != "" {
		return resolveVersioned(ctx, cfg, candidates, logger)
	}

	if cfg.Install.ReleaseURL != "" {
		root, info, err := acquireLatestFromURL(
			ctx,
			cfg.Paths.InstallRoot,
			cfg.Install.ReleaseURL,
			logger,
		)
		if err != nil {
			return Install{}, err
		}
		return Install{Root: root, Info: info}, nil
	}

	if latest := latestCandidate(candidates); latest != nil {
		logger.Info("install selected", "install_root", latest.Path, "installed", latest.Version)
		return Install{Root: latest.Path, Info: latest.Info}, nil
	}

	return Install{}, fmt.Errorf("no Foundry install found under %s", cfg.Paths.InstallRoot)
}

func resolveVersioned(
	ctx context.Context,
	cfg appconfig.Config,
	candidates []installCandidate,
	logger *slog.Logger,
) (Install, error) {
	if match := matchCandidate(candidates, cfg.Install.Version); match != nil {
		logger.Info("install selected",
			"install_root", match.Path,
			"installed", match.Version,
			"desired", cfg.Install.Version,
		)
		return Install{Root: match.Path, Info: match.Info}, nil
	}

	targetRoot := filepath.Join(cfg.Paths.InstallRoot, normalizeVersionDir(cfg.Install.Version))
	logger.Info("install decision",
		"install_root", targetRoot,
		"desired", cfg.Install.Version,
		"action", lifecycle.ActionInstall,
	)

	switch {
	case cfg.Install.ReleaseURL != "":
		if err := acquireFromURL(ctx, targetRoot, cfg.Install.ReleaseURL, logger); err != nil {
			return Install{}, err
		}
	case canAuthenticate(cfg):
		if err := acquireFromVersion(ctx, cfg, targetRoot, logger); err != nil {
			return Install{}, err
		}
	default:
		return Install{}, fmt.Errorf(
			"no install found for version %q and no acquisition source configured",
			cfg.Install.Version,
		)
	}

	info, err := lifecycle.DetectInstalled(targetRoot)
	if err != nil {
		return Install{}, fmt.Errorf("detect install after acquisition: %w", err)
	}
	if !versionMatches(info.Version, cfg.Install.Version) {
		return Install{}, fmt.Errorf(
			"installed Foundry version %q does not satisfy desired version %q",
			info.Version, cfg.Install.Version,
		)
	}
	return Install{Root: targetRoot, Info: info}, nil
}
