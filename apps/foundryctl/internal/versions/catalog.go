package versions

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundryctl/config"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/fourcery/forge"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/fourcery/source"
)

type Catalog struct {
	paths   config.PathsConfig
	install config.InstallConfig
	logger  *slog.Logger
}

func New(paths config.PathsConfig, install config.InstallConfig, logger *slog.Logger) *Catalog {
	return &Catalog{paths: paths, install: install, logger: logger}
}

func (m *Catalog) Installed(_ context.Context) ([]string, error) {
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

func (m *Catalog) Download(ctx context.Context, version, url string) error {
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
