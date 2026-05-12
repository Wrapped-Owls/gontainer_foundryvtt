package step

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrypatch/applier"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrypatch/ledger"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrypatch/manifest"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/fourcery/version"
)

type patchesStep struct{}

// Patches returns a Step that applies any manifest-declared patches for the
// installed version. The applier is gated by a per-install ledger so each
// patch runs at most once per (install, patch-hash).
func Patches() Step { return patchesStep{} }

func (patchesStep) Apply(ctx context.Context, s *State, logger *slog.Logger) error {
	m, err := manifest.Load(s.App.Paths.ManifestPath)
	if err != nil {
		logger.Warn(
			"patch manifest load failed; skipping",
			"path", s.App.Paths.ManifestPath,
			"err", err,
		)
		return nil
	}
	ver := s.Install.Version
	if ver.IsZero() {
		ver = version.Parse(s.App.Install.Version)
	}
	patches, err := m.Applicable(ver.String())
	if err != nil {
		logger.Warn("patch filtering failed; skipping", "err", err)
		return nil
	}
	if len(patches) == 0 {
		return nil
	}

	l, err := ledger.Load(s.Install.Root)
	if err != nil {
		if errors.Is(err, ledger.ErrLedgerCorrupt) {
			logger.Warn(
				"patch ledger corrupt; rebuilding",
				"path", ledger.Path(s.Install.Root),
				"err", err,
			)
			l = &ledger.Ledger{}
		} else {
			return fmt.Errorf("step patches: load ledger: %w", err)
		}
	}

	a := &applier.Applier{
		Root:      s.Install.Root,
		Ledger:    l,
		OnApplied: l.Upsert,
	}
	if err = a.Apply(ctx, patches, func(f string, args ...any) {
		logger.Info(fmt.Sprintf(f, args...))
	}); err != nil {
		return err
	}
	if err = ledger.Save(s.Install.Root, l); err != nil {
		return fmt.Errorf("step patches: save ledger: %w", err)
	}
	return nil
}
