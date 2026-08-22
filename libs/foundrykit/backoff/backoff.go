package backoff

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrykit/fsperm"
)

// OnFailure records a failure; an uptime past HealthyUptime forgets the history
// first. A cache it cannot clear costs a longer delay, never the supervisor.
func (m *Manager) OnFailure(exitCode int, uptime time.Duration) Decision {
	if m.KubernetesBypass {
		return Decision{Mode: ModeKubernetes, ExitCode: exitCode}
	}
	if uptime >= HealthyUptime {
		m.memFailures = 0
		_ = m.Reset()
	}

	statePath, hasCache := m.statePath()
	if !hasCache {
		return m.degraded(exitCode)
	}

	prev, _ := readState(statePath) // missing/corrupt -> zero value, treated as no prior failures
	next := State{
		ConsecutiveFailures: prev.ConsecutiveFailures + 1,
		LastFailureTS:       m.now().UTC().Format(time.RFC3339),
	}
	if err := writeStateAtomic(statePath, next); err != nil {
		return m.degraded(exitCode)
	}
	m.memFailures = next.ConsecutiveFailures

	return Decision{
		Mode:      ModeBackoff,
		Delay:     computeDelay(next.ConsecutiveFailures),
		ExitCode:  exitCode,
		State:     next,
		StateFile: statePath,
	}
}

func (m *Manager) Reset() error {
	m.memFailures = 0

	statePath, hasCache := m.statePath()
	if !hasCache {
		return nil
	}
	if err := os.Remove(statePath); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("clear backoff state: %w", err)
	}
	return nil
}

// degraded keeps the schedule running in memory when the state file is unreachable:
// losing persistence must not stop the supervisor restarting.
func (m *Manager) degraded(exitCode int) Decision {
	m.memFailures++
	return Decision{
		Mode:     ModeNoCache,
		Delay:    computeDelay(m.memFailures),
		ExitCode: exitCode,
		State: State{
			ConsecutiveFailures: m.memFailures,
			LastFailureTS:       m.now().UTC().Format(time.RFC3339),
		},
	}
}

func (m *Manager) statePath() (path string, hasCache bool) {
	if m.CacheDir == "" {
		return "", false
	}
	if err := os.MkdirAll(m.CacheDir, fsperm.Dir); err != nil {
		return "", false
	}
	return filepath.Join(m.CacheDir, stateFile), true
}

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
