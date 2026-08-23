package backoff

import (
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"testing/synctest"
	"time"
)

func TestKubernetesBypass(t *testing.T) {
	t.Parallel()

	m := &Manager{KubernetesBypass: true, CacheDir: t.TempDir()}
	d := m.OnFailure(7, 0)
	if d.Mode != ModeKubernetes || d.ExitCode != 7 || d.Delay != 0 {
		t.Fatalf("unexpected decision: %+v", d)
	}
	if _, err := os.Stat(filepath.Join(m.CacheDir, stateFile)); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("state file should not exist: %v", err)
	}
}

func TestNoCacheMode(t *testing.T) {
	t.Parallel()

	m := &Manager{CacheDir: ""}
	d := m.OnFailure(2, 0)
	if d.Mode != ModeNoCache {
		t.Fatalf("expected no-cache mode, got %v", d.Mode)
	}
}

func TestPersistentSchedule(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		dir := t.TempDir()
		m := &Manager{CacheDir: dir}

		expectedDelays := []time.Duration{0, 10 * time.Second, 20 * time.Second, 40 * time.Second}
		for i, want := range expectedDelays {
			d := m.OnFailure(1, 0)
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
	})
}

func TestSleepCancellation(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := Sleep(ctx, time.Hour); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
	if err := Sleep(context.Background(), 0); err != nil {
		t.Fatalf("zero delay should return nil: %v", err)
	}
}

func TestOnFailureForgetsHistoryAfterAHealthyRun(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		uptime        time.Duration
		wantForgotten bool
	}{
		{"a crash during startup keeps the history", HealthyUptime - time.Second, false},
		{"exactly the threshold counts as recovered", HealthyUptime, true},
		{"a long healthy run forgets the history", 24 * time.Hour, true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			m := &Manager{CacheDir: t.TempDir()}
			for range 4 {
				m.OnFailure(1, 0)
			}

			d := m.OnFailure(1, testCase.uptime)
			wantFailures := 5
			if testCase.wantForgotten {
				wantFailures = 1
			}
			if d.State.ConsecutiveFailures != wantFailures {
				t.Fatalf("counter = %d, want %d", d.State.ConsecutiveFailures, wantFailures)
			}
		})
	}
}

func TestOnFailureWithoutCacheStillForgetsAHealthyRun(t *testing.T) {
	t.Parallel()

	m := &Manager{CacheDir: ""}
	m.OnFailure(1, 0)

	d := m.OnFailure(1, HealthyUptime)
	if d.State.ConsecutiveFailures != 1 {
		t.Fatalf("in-memory counter = %d, want 1", d.State.ConsecutiveFailures)
	}
}

func TestOnFailureKeepsSchedulingWhenTheHistoryCannotBeCleared(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	blocked := filepath.Join(dir, stateFile)
	if err := os.MkdirAll(filepath.Join(blocked, "occupied"), 0o755); err != nil {
		t.Fatal(err)
	}

	m := &Manager{CacheDir: dir}
	d := m.OnFailure(1, HealthyUptime)
	if d.Mode == ModeKubernetes {
		t.Fatalf("mode = %v, want a schedule the caller can wait on", d.Mode)
	}
	if d.State.ConsecutiveFailures == 0 {
		t.Fatal("the failure should still have been counted")
	}
}

func TestOnFailureCountsInMemoryWhenTheCacheIsUnwritable(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	blocked := filepath.Join(dir, "cache")
	if err := os.WriteFile(blocked, []byte("not a dir"), 0o600); err != nil {
		t.Fatal(err)
	}

	m := &Manager{CacheDir: blocked}
	wantDelays := []time.Duration{0, BaseDelay, 2 * BaseDelay}
	for i, want := range wantDelays {
		d := m.OnFailure(1, 0)
		if d.Mode != ModeNoCache {
			t.Fatalf("iter %d: mode = %v, want %v", i, d.Mode, ModeNoCache)
		}
		if d.Delay != want {
			t.Fatalf("iter %d: delay = %v, want %v", i, d.Delay, want)
		}
		if d.State.ConsecutiveFailures != i+1 {
			t.Fatalf("iter %d: counter = %d, want %d", i, d.State.ConsecutiveFailures, i+1)
		}
	}
}
