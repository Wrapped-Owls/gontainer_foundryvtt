package versions

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundryctl/config"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/fourcery/forge"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/fourcery/source"
)

// Manager resolves and acquires Foundry installs using the same configuration
// the activation pipeline uses at boot.
type Manager struct {
	paths   config.PathsConfig
	install config.InstallConfig
	logger  *slog.Logger
}

// New builds a Manager from the resolved path and install configuration.
func New(paths config.PathsConfig, install config.InstallConfig, logger *slog.Logger) *Manager {
	return &Manager{paths: paths, install: install, logger: logger}
}

// Installed returns the versions already present under the install root, newest
// first. Versions that cannot be parsed from disk are skipped.
func (m *Manager) Installed(_ context.Context) ([]string, error) {
	f, err := forge.New(m.paths.InstallRoot).WithLogger(m.logger).Build()
	if err != nil {
		return nil, fmt.Errorf("versions: build forge: %w", err)
	}
	candidates, err := f.Installed()
	if err != nil {
		return nil, fmt.Errorf("versions: list installed: %w", err)
	}
	out := make([]string, 0, len(candidates))
	for _, c := range candidates {
		if v := c.Version.String(); v != "" {
			out = append(out, v)
		}
	}
	return out, nil
}

// Download acquires the given version, preferring the caller-supplied presigned
// url, then the configured authenticated session, then an already-installed
// copy. It returns an error when none of these can satisfy the request.
func (m *Manager) Download(ctx context.Context, version, url string) error {
	if version == "" {
		return fmt.Errorf("versions: version is required")
	}
	releaseURL := m.install.ReleaseURL
	if url != "" {
		releaseURL = url
	}
	reg := source.NewRegistry(source.Config{
		SourcesDir: m.paths.SourcesDir,
		ReleaseURL: releaseURL,
		Version:    version,
		Session:    m.install.Session,
		Username:   m.install.Username,
		Password:   m.install.Password,
	})
	sources, err := reg.Enumerate(ctx)
	if err != nil {
		return fmt.Errorf("versions: enumerate sources: %w", err)
	}
	f, err := forge.New(m.paths.InstallRoot).WithSources(sources...).WithLogger(m.logger).Build()
	if err != nil {
		return fmt.Errorf("versions: build forge: %w", err)
	}
	plan, err := f.Resolve(ctx, version)
	if err != nil {
		return fmt.Errorf("versions: resolve %q: %w", version, err)
	}
	if _, err = f.Acquire(ctx, plan); err != nil {
		return fmt.Errorf("versions: acquire %q: %w", version, err)
	}
	return nil
}
