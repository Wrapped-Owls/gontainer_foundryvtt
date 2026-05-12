package backoff

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrykit/fsperm"
)

// New constructs a Manager with the supplied cache directory.
func New(cacheDir string) *Manager {
	return &Manager{CacheDir: cacheDir, Now: time.Now}
}

// NewFromEnv constructs a Manager from environment variables:
//
//   - CONTAINER_CACHE: default "/data/container_cache" when unset; an
//     explicit empty string disables persistence.
//   - KUBERNETES_SERVICE_HOST: any non-empty value enables bypass.
func NewFromEnv() *Manager {
	cfg := Default()
	_ = LoadFromEnv(&cfg)
	return NewFromConfig(cfg)
}

// Reset deletes the persisted state file. Safe to call when the file or
// the cache directory does not exist.
func (m *Manager) Reset() error {
	if m.CacheDir == "" {
		return nil
	}
	err := os.Remove(filepath.Join(m.CacheDir, stateFile))
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	return nil
}

// OnFailure records a failure and computes the next action.
func (m *Manager) OnFailure(exitCode int) (Decision, error) {
	if m.KubernetesBypass {
		return Decision{Mode: ModeKubernetes, ExitCode: exitCode}, nil
	}

	if m.CacheDir != "" {
		if err := os.MkdirAll(m.CacheDir, fsperm.Dir); err != nil {
			// Fall through to no-cache mode on permission failures.
			return Decision{Mode: ModeNoCache, ExitCode: exitCode}, nil
		}
	}
	if m.CacheDir == "" {
		return Decision{Mode: ModeNoCache, ExitCode: exitCode}, nil
	}

	statePath := filepath.Join(m.CacheDir, stateFile)
	prev, _ := readState(statePath) // missing/corrupt → zero value, treated as no prior failures

	n := prev.ConsecutiveFailures + 1
	delay := computeDelay(n)

	now := m.now()
	next := State{
		ConsecutiveFailures: n,
		LastFailureTS:       now.UTC().Format("2006-01-02T15:04:05Z"),
	}
	if err := writeStateAtomic(statePath, next); err != nil {
		// Mirror the bash fallback: if we can't persist, sleep indefinitely.
		return Decision{Mode: ModeNoCache, ExitCode: exitCode}, nil
	}

	return Decision{
		Mode:      ModeBackoff,
		Delay:     delay,
		ExitCode:  exitCode,
		State:     next,
		StateFile: statePath,
	}, nil
}

// Sleep blocks for d or until ctx is done, whichever comes first.
// Returns ctx.Err() if interrupted, nil otherwise.
func Sleep(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return nil
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(d):
		return nil
	}
}

func (m *Manager) now() time.Time {
	if m.Now != nil {
		return m.Now()
	}
	return time.Now()
}
