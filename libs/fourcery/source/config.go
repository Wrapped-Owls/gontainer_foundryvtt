package source

import (
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrykit/jsonhttp"
)

// HTTPDoer is the minimal HTTP interface the URL/session sources need.
// It matches *http.Client and any test fake.
type HTTPDoer interface {
	Do(*http.Request) (*http.Response, error)
}

// Config is the inputs from app config that a Registry turns into a
// list of Source values. SourcesDir, ReleaseURL, Version, and the
// auth fields may each be empty; the registry picks the strategies
// that the present fields can satisfy.
type Config struct {
	SourcesDir string
	ReleaseURL string
	Version    string
	Session    string
	Username   string
	Password   string
}

// Options collects optional dependencies that override defaults during
// Registry construction.
type Options struct {
	// HTTPClient is used for URL downloads and release-fetch calls.
	// Defaults to http.DefaultClient with a 30 minute download
	// timeout via httpClientWithTimeout.
	HTTPClient HTTPDoer
}

// Option is the functional-option signature used by NewRegistry.
type Option func(*Options)
