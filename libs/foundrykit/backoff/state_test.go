package backoff

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestReadWriteStateRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state.json")

	// Missing file returns zero State.
	s, err := readState(path)
	if err == nil {
		t.Fatal("expected error for missing state file")
	}
	_ = s

	// Write and read back.
	want := State{
		ConsecutiveFailures: 3,
		LastFailureTS: time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC).
			Format("2006-01-02T15:04:05Z"),
	}
	if err := writeStateAtomic(path, want); err != nil {
		t.Fatalf("writeStateAtomic: %v", err)
	}
	got, err := readState(path)
	if err != nil {
		t.Fatalf("readState: %v", err)
	}
	if got.ConsecutiveFailures != want.ConsecutiveFailures ||
		got.LastFailureTS != want.LastFailureTS {
		t.Errorf("round-trip mismatch: got %+v, want %+v", got, want)
	}
}

func TestReadStateInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(path, []byte("not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := readState(path)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestReadStateNegativeFailuresClampedToZero(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "neg.json")
	if err := os.WriteFile(path, []byte(`{"consecutive_failures":-5}`), 0o644); err != nil {
		t.Fatal(err)
	}
	s, err := readState(path)
	if err != nil {
		t.Fatalf("readState: %v", err)
	}
	if s.ConsecutiveFailures != 0 {
		t.Errorf("negative failures should clamp to 0, got %d", s.ConsecutiveFailures)
	}
}

func TestWriteStateAtomicIsAtomic(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state.json")

	// First write.
	s1 := State{ConsecutiveFailures: 1}
	if err := writeStateAtomic(path, s1); err != nil {
		t.Fatal(err)
	}
	// Overwrite.
	s2 := State{ConsecutiveFailures: 2}
	if err := writeStateAtomic(path, s2); err != nil {
		t.Fatal(err)
	}
	got, err := readState(path)
	if err != nil {
		t.Fatal(err)
	}
	if got.ConsecutiveFailures != 2 {
		t.Errorf("got %d, want 2", got.ConsecutiveFailures)
	}
	// No .tmp file left behind.
	if _, err := os.Stat(path + ".tmp"); !os.IsNotExist(err) {
		t.Error("tmp file should not exist after atomic write")
	}
}
