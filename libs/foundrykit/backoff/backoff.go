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

func New(cacheDir string) *Tracker {
	return &Tracker{CacheDir: cacheDir}
}

func NewFromEnv() *Tracker {
	cfg := Default()
	_ = LoadFromEnv(&cfg)
	return NewFromConfig(cfg)
}

func (m *Tracker) Reset() error {
	if m.CacheDir == "" {
		return nil
	}
	err := os.Remove(filepath.Join(m.CacheDir, stateFile))
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	return nil
}

func (m *Tracker) OnFailure(exitCode int) (Decision, error) {
	if m.KubernetesBypass {
		return Decision{Mode: ModeKubernetes, ExitCode: exitCode}, nil
	}

	if m.CacheDir != "" {
		if err := os.MkdirAll(m.CacheDir, fsperm.Dir); err != nil {
			return Decision{Mode: ModeNoCache, ExitCode: exitCode}, nil
		}
	}
	if m.CacheDir == "" {
		return Decision{Mode: ModeNoCache, ExitCode: exitCode}, nil
	}

	statePath := filepath.Join(m.CacheDir, stateFile)
	prev, _ := readState(statePath)

	n := prev.ConsecutiveFailures + 1
	delay := computeDelay(n)

	now := time.Now()
	next := State{
		ConsecutiveFailures: n,
		LastFailureTS:       now.UTC().Format(time.RFC3339),
	}
	if err := writeStateAtomic(statePath, next); err != nil {
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
