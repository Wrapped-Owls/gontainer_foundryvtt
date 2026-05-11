package source

import (
	"net/http"
)

type HTTPDoer interface {
	Do(*http.Request) (*http.Response, error)
}

type Config struct {
	SourcesDir string
	ReleaseURL string
	Version    string
	Session    string
	Username   string
	Password   string
}

type Options struct {
	HTTPClient HTTPDoer
}

type Option func(*Options)

func WithHTTPClient(c HTTPDoer) Option { return func(o *Options) { o.HTTPClient = c } }
