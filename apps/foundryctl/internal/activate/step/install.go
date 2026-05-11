package step

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/fourcery/forge"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/fourcery/source"
)

type installStep struct{}

// Install returns a Step that resolves or acquires the Foundry installation
// via the fourcery library.
func Install() Step { return installStep{} }

func (installStep) Apply(ctx context.Context, s *State, logger *slog.Logger) error {
	reg := source.NewRegistry(source.Config{
		SourcesDir: s.App.Paths.SourcesDir,
		ReleaseURL: s.App.Install.ReleaseURL,
		Version:    s.App.Install.Version,
		Session:    s.App.Install.Session,
		Username:   s.App.Install.Username,
		Password:   s.App.Install.Password,
	})
	sources, err := reg.Enumerate(ctx)
	if err != nil {
		return fmt.Errorf("step install: enumerate sources: %w", err)
	}

	f, err := forge.New(s.App.Paths.InstallRoot).
		WithSources(sources...).
		WithLogger(logger).
		Build()
	if err != nil {
		return fmt.Errorf("step install: build forge: %w", err)
	}

	plan, err := f.Resolve(ctx, s.App.Install.Version)
	if err != nil {
		return fmt.Errorf("step install: resolve: %w", err)
	}
	inst, err := f.Acquire(ctx, plan)
	if err != nil {
		return fmt.Errorf("step install: acquire: %w", err)
	}
	s.Install = inst
	return nil
}
