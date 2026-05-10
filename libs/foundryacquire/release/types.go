package release

import (
	"context"
	"errors"
	"math/rand/v2"
	"time"
)

const (
	InitialRetryDelay = 120 * time.Second
)

// Errors surfaced by Fetch.
var (
	ErrNoBuildNumber = errors.New("release: cannot derive build number from version")
	ErrEmptyURL      = errors.New("release: server returned empty URL")
)

// FetchOptions tunes retry behaviour.
type FetchOptions struct {
	// Retries is the number of additional attempts after the first
	// failure. 0 means "try once". Default 0.
	Retries int
	// Sleep is the back-off implementation. Default uses backoff.Sleep
	// (honours ctx cancellation).
	Sleep func(ctx context.Context, d time.Duration) error
	// Now allows tests to inject a clock for jitter calculations.
	Rand *rand.Rand
}

// releaseURLResp is the JSON payload returned by the FoundryVTT
// release URL endpoint.
type releaseURLResp struct {
	URL      string `json:"url"`
	Lifetime int    `json:"lifetime"`
}
