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

func (m *Tracker) OnFailure(exitCode int, uptime time.Duration) Decision {
	if m.KubernetesBypass {
		return Decision{Mode: ModeKubernetes, ExitCode: exitCode}
	}
	if uptime >= HealthyUptime { // reset here, else the delay saturates at MaxDelay for the volume's life
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
		LastFailureTS:       time.Now().UTC().Format(time.RFC3339),
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

func (m *Tracker) Reset() error {
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

func (m *Tracker) degraded(exitCode int) Decision {
	m.memFailures++
	return Decision{
		Mode:     ModeNoCache,
		Delay:    computeDelay(m.memFailures),
		ExitCode: exitCode,
		State: State{
			ConsecutiveFailures: m.memFailures,
			LastFailureTS:       time.Now().UTC().Format(time.RFC3339),
		},
	}
}

func (m *Tracker) statePath() (path string, hasCache bool) {
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
