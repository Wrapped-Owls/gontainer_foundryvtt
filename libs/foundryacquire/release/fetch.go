package release

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundryacquire/auth"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrykit/jsonhttp"
)

// Fetch returns a presigned URL for the requested version.
func Fetch(
	ctx context.Context,
	sess *auth.Session,
	version string,
	opts FetchOptions,
) (string, error) {
	build, err := buildNumber(version)
	if err != nil {
		return "", err
	}
	if opts.Sleep == nil {
		opts.Sleep = sleepCtx
	}

	releaseURL := FetchURL(build)
	totalAttempts := 1 + opts.Retries
	var lastErr error
	for attempt := 1; attempt <= totalAttempts; attempt++ {
		if attempt > 1 {
			delay := backoff(attempt, opts.Rand)
			if err = opts.Sleep(ctx, delay); err != nil {
				return "", err
			}
		}
		url, err := fetchOnce(ctx, sess, releaseURL)
		if err == nil {
			return url, nil
		}
		lastErr = err
	}
	return "", fmt.Errorf("release: failed after %d attempts: %w", totalAttempts, lastErr)
}

// FetchURL is the bare URL we hit (exported for tests of release.Fetch
// with an injected base URL).
func FetchURL(build string) string {
	return fmt.Sprintf(
		"%s/releases/download?build=%s&platform=node&response_type=json",
		auth.BaseURL,
		build,
	)
}

func fetchOnce(ctx context.Context, sess *auth.Session, releaseURL string) (string, error) {
	path := strings.TrimPrefix(releaseURL, auth.BaseURL)
	result, err := jsonhttp.Request[releaseURLResp, struct{}](ctx,
		jsonhttp.ClientConfig{
			BaseURL: auth.BaseURL,
			Headers: map[string]string{
				"User-Agent": sess.UserAgent,
				"Referer":    auth.BaseURL,
			},
			HTTP: sess.Client(),
		},
		jsonhttp.RequestConfig[struct{}]{
			Method: http.MethodGet,
			Path:   path,
		},
	)
	if err != nil {
		return "", err
	}
	if result.URL == "" {
		return "", ErrEmptyURL
	}
	return result.URL, nil
}

// buildNumber pulls the trailing build component from a Foundry version
// like "14.361" → "361". A bare numeric input is accepted as-is.
func buildNumber(version string) (string, error) {
	v := strings.TrimSpace(version)
	if v == "" {
		return "", ErrNoBuildNumber
	}
	parts := strings.Split(v, ".")
	last := parts[len(parts)-1]
	if last == "" {
		return "", ErrNoBuildNumber
	}
	return last, nil
}
