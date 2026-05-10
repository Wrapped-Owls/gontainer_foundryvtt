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

var (
	ErrNoBuildNumber = errors.New("release: cannot derive build number from version")
	ErrEmptyURL      = errors.New("release: server returned empty URL")
)

type FetchOptions struct {
	Retries int
	Sleep   func(ctx context.Context, d time.Duration) error
	Rand    *rand.Rand
}

type releaseURLResp struct { // mirrors the FoundryVTT release URL endpoint's JSON response
	URL      string `json:"url"`
	Lifetime int    `json:"lifetime"`
}
