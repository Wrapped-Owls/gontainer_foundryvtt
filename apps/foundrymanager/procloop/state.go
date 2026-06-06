package procloop

import (
	"context"
	"log/slog"

	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/profile"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundryruntime/jsruntime"
)

type State struct {
	DataPath    string
	InstallRoot string
	MainScript  string
	JSRuntime   jsruntime.Runtime
	Port        int
	Version     string
	Profiles    []profile.Profile
}

type Activator interface {
	Switch(ctx context.Context, logger *slog.Logger, p profile.Profile) (State, error)
}
