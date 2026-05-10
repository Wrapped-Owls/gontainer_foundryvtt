package release

import (
	"context"
	"math/rand/v2"
	"time"
)

func backoff(attempt int, r *rand.Rand) time.Duration {
	exp := 1 << (attempt - 1)
	jitter := time.Duration(0)
	if r != nil {
		jitter = time.Duration(r.Int64N(int64(InitialRetryDelay)))
	} else {
		jitter = time.Duration(rand.Int64N(int64(InitialRetryDelay)))
	}
	return InitialRetryDelay*time.Duration(exp) + jitter
}

func sleepCtx(ctx context.Context, d time.Duration) error {
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
