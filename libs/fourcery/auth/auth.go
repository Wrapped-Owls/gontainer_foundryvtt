package auth

import (
	"errors"
	"net/http"
)

const BaseURL = "https://foundryvtt.com"

const LoginPath = "/auth/login/"

const DefaultUserAgent = "node-fetch"

var (
	ErrCSRFNotFound         = errors.New("auth: csrfmiddlewaretoken not found in login form")
	ErrCommunityLinkMissing = errors.New(
		"auth: community URL not found after login (likely bad credentials)",
	)
	ErrSessionCookieMissing = errors.New(
		"auth: no sessionid cookie set after login (likely bad credentials)",
	)
)

type Options struct {
	HTTPClient *http.Client
	UserAgent  string
}
