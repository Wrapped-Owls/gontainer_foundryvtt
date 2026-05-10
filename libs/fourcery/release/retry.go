package release

import (
	"math/rand/v2"
	"time"
)

func retryDelay(attempt int, r *rand.Rand) time.Duration {
	exp := 1 << (attempt - 1)
	var jitter time.Duration
	if r != nil {
		jitter = time.Duration(r.Int64N(int64(InitialRetryDelay)))
	} else {
		jitter = time.Duration(rand.Int64N(int64(InitialRetryDelay)))
	}
	return InitialRetryDelay*time.Duration(exp) + jitter
}
