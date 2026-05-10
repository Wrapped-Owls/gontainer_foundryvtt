package release

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestBackoffIncreases(t *testing.T) {
	// Each attempt should produce a delay >= InitialRetryDelay.
	for attempt := 2; attempt <= 5; attempt++ {
		d := backoff(attempt, nil)
		if d < InitialRetryDelay {
			t.Errorf("backoff(attempt=%d) = %v, want >= %v", attempt, d, InitialRetryDelay)
		}
	}
}

func TestBackoffGrowsWithAttempt(t *testing.T) {
	// Higher attempts should produce larger or equal delays.
	prev := backoff(1, nil)
	for attempt := 2; attempt <= 6; attempt++ {
		cur := backoff(attempt, nil)
		if cur < prev {
			t.Errorf("backoff not monotone: attempt %d gave %v < attempt %d gave %v",
				attempt, cur, attempt-1, prev)
		}
		prev = cur
	}
}

func TestSleepCtxZeroDuration(t *testing.T) {
	if err := sleepCtx(context.Background(), 0); err != nil {
		t.Fatalf("zero duration should return nil, got %v", err)
	}
	if err := sleepCtx(context.Background(), -1); err != nil {
		t.Fatalf("negative duration should return nil, got %v", err)
	}
}

func TestSleepCtxCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := sleepCtx(ctx, time.Hour)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}
