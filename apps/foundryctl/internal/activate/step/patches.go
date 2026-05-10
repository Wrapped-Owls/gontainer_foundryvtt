package step

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrypatch/applier"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrypatch/manifest"
)

type patchesStep struct{}

// Patches returns a Step that applies any manifest-declared patches for the installed version.
func Patches() Step { return patchesStep{} }

func (patchesStep) Apply(ctx context.Context, s *State, logger *slog.Logger) error {
	m, err := manifest.Load(s.App.Paths.ManifestPath)
	if err != nil {
		logger.Warn(
			"patch manifest load failed; skipping",
			"path",
			s.App.Paths.ManifestPath,
			"err",
			err,
		)
		return nil
	}
	version := s.Install.Info.Version
	if version == "" {
		version = s.App.Install.Version
	}
	patches, err := m.Applicable(version)
	if err != nil {
		logger.Warn("patch filtering failed; skipping", "err", err)
		return nil
	}
	if len(patches) == 0 {
		return nil
	}
	a := &applier.Applier{Root: s.Install.Root}
	return a.Apply(ctx, patches, func(f string, args ...any) {
		logger.Info(fmt.Sprintf(f, args...))
	})
}
