package step

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrykit/fsperm"
)

type ensureDirsStep struct{}

// EnsureDirs returns a Step that creates all runtime directories that may be
// shadowed when operators mount volumes over /foundry or /data. It must run
// after AppConfig so paths are resolved from config.
func EnsureDirs() Step { return ensureDirsStep{} }

func (ensureDirsStep) Apply(_ context.Context, s *State, logger *slog.Logger) error {
	dirs := []string{
		s.App.Paths.DataPath,
		s.App.Paths.InstallRoot,
		s.App.Paths.SourcesDir,
	}
	for _, d := range dirs {
		if d == "" {
			continue
		}
		if err := os.MkdirAll(d, fsperm.Dir); err != nil {
			return fmt.Errorf("ensure dir %s: %w", d, err)
		}
		logger.Debug("dir ready", "path", d)
	}
	return nil
}
