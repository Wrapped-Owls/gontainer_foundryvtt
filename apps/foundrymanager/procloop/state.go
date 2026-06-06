// Package procloop runs the Foundry process in a supervised restart loop,
// handling backoff delays and profile switches.
package procloop

import (
	"context"
	"log/slog"

	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/profile"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundryruntime/jsruntime"
)

// State holds the resolved runtime information needed to launch and manage
// the Foundry process.
type State struct {
	DataPath    string
	InstallRoot string
	MainScript  string
	JSRuntime   jsruntime.Runtime
	Port        int
	Version     string
	Profiles    []profile.Profile
}

// Activator abstracts the activation pipeline, allowing the Runner to
// request a profile switch without importing app-layer packages.
type Activator interface {
	Switch(ctx context.Context, logger *slog.Logger, p profile.Profile) (State, error)
}
