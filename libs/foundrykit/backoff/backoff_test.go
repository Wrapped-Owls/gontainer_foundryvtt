package backoff

import (
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestKubernetesBypass(t *testing.T) {
	m := &Manager{KubernetesBypass: true, CacheDir: t.TempDir()}
	d, err := m.OnFailure(7)
	if err != nil {
		t.Fatal(err)
	}
	if d.Mode != ModeKubernetes || d.ExitCode != 7 || d.Delay != 0 {
		t.Fatalf("unexpected decision: %+v", d)
	}
	// State file must NOT have been created.
	if _, err = os.Stat(filepath.Join(m.CacheDir, stateFile)); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("state file should not exist: %v", err)
	}
}

func TestNoCacheMode(t *testing.T) {
	m := &Manager{CacheDir: ""}
	d, err := m.OnFailure(2)
	if err != nil {
		t.Fatal(err)
	}
	if d.Mode != ModeNoCache {
		t.Fatalf("expected no-cache mode, got %v", d.Mode)
	}
}

func TestPersistentSchedule(t *testing.T) {
	dir := t.TempDir()
	m := &Manager{CacheDir: dir, Now: func() time.Time {
		return time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	}}

	expectedDelays := []time.Duration{0, 10 * time.Second, 20 * time.Second, 40 * time.Second}
	for i, want := range expectedDelays {
		d, err := m.OnFailure(1)
		if err != nil {
			t.Fatal(err)
		}
		if d.Mode != ModeBackoff {
			t.Fatalf("iter %d: mode = %v", i, d.Mode)
		}
		if d.Delay != want {
			t.Fatalf("iter %d: delay = %v want %v", i, d.Delay, want)
		}
		if d.State.ConsecutiveFailures != i+1 {
			t.Fatalf("iter %d: counter = %d want %d", i, d.State.ConsecutiveFailures, i+1)
		}
	}

	// Verify the on-disk JSON shape so external tooling can parse it.
	b, err := os.ReadFile(filepath.Join(dir, stateFile))
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]any
	if err = json.Unmarshal(b, &raw); err != nil {
		t.Fatal(err)
	}
	if raw["consecutive_failures"].(float64) != 4 {
		t.Errorf("on-disk counter = %v, want 4", raw["consecutive_failures"])
	}
	if _, ok := raw["last_failure_timestamp"].(string); !ok {
		t.Errorf("missing last_failure_timestamp")
	}
}

func TestResetDeletesState(t *testing.T) {
	dir := t.TempDir()
	m := New(dir)
	if _, err := m.OnFailure(1); err != nil {
		t.Fatal(err)
	}
	if err := m.Reset(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, stateFile)); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("state file should be removed: %v", err)
	}
	// Reset is idempotent.
	if err := m.Reset(); err != nil {
		t.Fatalf("idempotent Reset: %v", err)
	}
}

func TestCorruptStateRecovers(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, stateFile), []byte("garbage"), 0o644); err != nil {
		t.Fatal(err)
	}
	m := New(dir)
	d, err := m.OnFailure(99)
	if err != nil {
		t.Fatal(err)
	}
	if d.State.ConsecutiveFailures != 1 {
		t.Fatalf("corrupt file should reset counter, got %d", d.State.ConsecutiveFailures)
	}
}

func TestSleepCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := Sleep(ctx, time.Hour); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
	if err := Sleep(context.Background(), 0); err != nil {
		t.Fatalf("zero delay should return nil: %v", err)
	}
}
