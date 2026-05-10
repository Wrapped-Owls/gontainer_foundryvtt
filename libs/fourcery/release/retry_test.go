package release

import (
	"testing"
)

func TestRetryDelayAtLeastInitial(t *testing.T) {
	for attempt := 2; attempt <= 5; attempt++ {
		d := retryDelay(attempt, nil)
		if d < InitialRetryDelay {
			t.Errorf("retryDelay(attempt=%d) = %v, want >= %v", attempt, d, InitialRetryDelay)
		}
	}
}

func TestRetryDelayGrowsWithAttempt(t *testing.T) {
	prev := retryDelay(1, nil)
	for attempt := 2; attempt <= 6; attempt++ {
		cur := retryDelay(attempt, nil)
		if cur < prev {
			t.Errorf("retryDelay not monotone: attempt %d gave %v < attempt %d gave %v",
				attempt, cur, attempt-1, prev)
		}
		prev = cur
	}
}
