package source

import (
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrykit/jsonhttp"
)

type HTTPDoer = jsonhttp.HTTPDoer

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
