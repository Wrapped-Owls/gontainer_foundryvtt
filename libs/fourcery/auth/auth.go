// Package auth handles foundryvtt.com authentication and is the source
// of truth for the authenticated *http.Client used by the release and
// license packages.
//
// It performs a two-step login flow: GET the homepage to pick up the
// CSRF middleware token + initial cookies, then POST to /auth/login/ with
// the credentials. On success the http.CookieJar holds the sessionid
// cookie and the resolved (case-corrected, e-mail-resolved) username is
// returned.
//
// The Session struct can be persisted to / restored from disk so a
// build-stage tool can reuse a runtime-stage authentication and vice
// versa.
package auth

import (
	"errors"
	"net/http"
)

// BaseURL is the foundryvtt.com origin used for every request.
const BaseURL = "https://foundryvtt.com"

// LoginPath is the form endpoint.
const LoginPath = "/auth/login/"

// DefaultUserAgent is the User-Agent header sent with every request.
const DefaultUserAgent = "node-fetch"

// Errors surfaced by Login.
var (
	ErrCSRFNotFound         = errors.New("auth: csrfmiddlewaretoken not found in login form")
	ErrCommunityLinkMissing = errors.New(
		"auth: community URL not found after login (likely bad credentials)",
	)
	ErrSessionCookieMissing = errors.New(
		"auth: no sessionid cookie set after login (likely bad credentials)",
	)
)

// Options configures Login and Restore.
type Options struct {
	// HTTPClient overrides the underlying transport. The cookie jar is
	// always replaced; only Transport / Timeout are inherited. Nil → a
	// new client with a 30s timeout is used.
	HTTPClient *http.Client
	// UserAgent sent with every request.
	UserAgent string
}
