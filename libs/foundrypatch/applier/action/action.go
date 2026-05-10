// Package action provides the strategy implementations for each patch action type.
package action

import (
	"context"
	"io/fs"
	"net/http"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrypatch/manifest"
)

const (
	DirPerm  fs.FileMode = 0o755
	FilePerm fs.FileMode = 0o644
)

// HTTPDoer abstracts HTTP client calls for testability.
type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Runner is the strategy interface for a single action type.
type Runner interface {
	Run(ctx context.Context, act manifest.Action, dest string) error
}
