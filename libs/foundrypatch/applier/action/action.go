package action

import (
	"context"
	"net/http"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrypatch/manifest"
)

type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

type Runner interface {
	Run(ctx context.Context, act manifest.Action, dest string) error
}
