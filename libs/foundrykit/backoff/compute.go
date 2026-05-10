package backoff

import "time"

// computeDelay returns the exponential delay for n consecutive failures.
//
//	n == 1 → 0          (first failure exits immediately)
//	n == 2 → 10s
//	n == 3 → 20s
//	n >= 9 → 960s       (capped)
func computeDelay(n int) time.Duration {
	if n <= 1 {
		return 0
	}
	if n >= 9 {
		return MaxDelay
	}
	shift := uint(n - 2)
	return BaseDelay * (1 << shift)
}
