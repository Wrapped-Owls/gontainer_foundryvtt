package backoff

import "time"

func computeDelay(n int) time.Duration {
	const cappedAtFailures = 9 // failure count where BaseDelay*2^(n-2) would exceed MaxDelay

	if n <= 1 {
		return 0
	}
	if n >= cappedAtFailures {
		return MaxDelay
	}
	shift := uint(n - 2)
	return BaseDelay * (1 << shift)
}
