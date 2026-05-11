package applier

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrypatch/applier/action"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrypatch/ledger"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrypatch/manifest"
)

// HTTPDoer abstracts HTTP client calls.
type HTTPDoer = action.HTTPDoer

// ErrHashMismatch is re-exported from the action package for callers that
// inspect download errors.
var ErrHashMismatch = action.ErrHashMismatch

// Applier executes patch actions. Root is the Foundry install root that
// every action's Dest is resolved relative to. HTTPClient defaults to
// http.DefaultClient.
type Applier struct {
	Root       string
	HTTPClient HTTPDoer

	// Ledger, when non-nil, gates each patch on a (id, hash)
	// already-applied check. The Applier never writes the ledger
	// itself; instead it calls OnApplied after each successful
	// patch so the caller can persist the updated state.
	Ledger *ledger.Ledger

	// OnApplied is called once per successfully applied patch, with
	// the Entry the caller should persist. Nil disables the callback.
	OnApplied func(ledger.Entry)

	// Now is the clock used to stamp ledger entries; defaults to
	// time.Now().UTC().
	Now func() time.Time

	runners map[manifest.ActionType]action.Runner
}

// Apply runs every action in every applicable patch in order. The first
// error aborts the run. logf is optional. When Ledger is non-nil,
// patches whose (id, content-hash) match an existing ledger entry are
// skipped.
func (a *Applier) Apply(
	ctx context.Context,
	patches []manifest.Patch,
	logf func(string, ...any),
) error {
	if logf == nil {
		logf = func(string, ...any) {}
	}
	a.initRunners()
	now := a.now
	for _, p := range patches {
		hash := ledger.HashPatch(p)
		if a.Ledger != nil && a.Ledger.Has(p.ID, hash) {
			logf("patch %s already applied (hash %s), skipping", p.ID, shortHash(hash))
			continue
		}
		logf("applying patch %s: %s", p.ID, p.Description)
		for i, act := range p.Actions {
			if err := a.runAction(ctx, act); err != nil {
				return fmt.Errorf("patch %s action[%d] %s: %w", p.ID, i, act.Type, err)
			}
		}
		if a.OnApplied != nil {
			a.OnApplied(ledger.Entry{
				ID:        p.ID,
				Versions:  p.Versions,
				PatchHash: hash,
				AppliedAt: now(),
			})
		}
	}
	return nil
}

func (a *Applier) now() time.Time {
	if a.Now != nil {
		return a.Now()
	}
	return time.Now().UTC()
}

func shortHash(h string) string {
	if len(h) > 8 {
		return h[:8]
	}
	return h
}

func (a *Applier) initRunners() {
	if a.HTTPClient == nil {
		a.HTTPClient = http.DefaultClient
	}
	a.runners = map[manifest.ActionType]action.Runner{
		manifest.ActionDownload:    action.Download(a.HTTPClient),
		manifest.ActionZipOverlay:  action.ZipOverlay(a.HTTPClient),
		manifest.ActionFileReplace: action.FileReplace(),
	}
}

func (a *Applier) runAction(ctx context.Context, act manifest.Action) error {
	dest, err := a.safeJoin(act.Dest)
	if err != nil {
		return err
	}
	runner, ok := a.runners[act.Type]
	if !ok {
		return fmt.Errorf("%w: %q", manifest.ErrUnknownAction, act.Type)
	}
	return runner.Run(ctx, act, dest)
}

func (a *Applier) safeJoin(p string) (string, error) {
	if filepath.IsAbs(p) {
		return "", fmt.Errorf("applier: dest must be relative: %q", p)
	}
	if strings.Contains(p, "..") {
		return "", fmt.Errorf("applier: dest may not contain '..': %q", p)
	}
	return filepath.Join(a.Root, filepath.Clean(p)), nil
}
