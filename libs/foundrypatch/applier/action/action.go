// Package action provides the strategy implementations for each patch action type.
package action

import (
	"context"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrykit/jsonhttp"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrypatch/manifest"
)

// HTTPDoer abstracts HTTP client calls for testability.
type HTTPDoer = jsonhttp.HTTPDoer

// Runner is the strategy interface for a single action type.
type Runner interface {
	Run(ctx context.Context, act manifest.Action, dest string) error
}
