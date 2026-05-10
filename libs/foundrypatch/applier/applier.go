package applier

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrypatch/applier/action"
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

	runners map[manifest.ActionType]action.Runner
}

// Apply runs every action in every applicable patch in order. The first
// error aborts the run. logf is optional.
func (a *Applier) Apply(
	ctx context.Context,
	patches []manifest.Patch,
	logf func(string, ...any),
) error {
	if logf == nil {
		logf = func(string, ...any) {}
	}
	a.initRunners()
	for _, p := range patches {
		logf("applying patch %s: %s", p.ID, p.Description)
		for i, act := range p.Actions {
			if err := a.runAction(ctx, act); err != nil {
				return fmt.Errorf("patch %s action[%d] %s: %w", p.ID, i, act.Type, err)
			}
		}
	}
	return nil
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
