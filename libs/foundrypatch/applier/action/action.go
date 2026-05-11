package action

import (
	"context"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrykit/jsonhttp"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrypatch/manifest"
)

type HTTPDoer = jsonhttp.HTTPDoer

type Runner interface {
	Run(ctx context.Context, act manifest.Action, dest string) error
}
