package procloop

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/config"
	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/internal/controller"
	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/internal/dashboard"
	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/internal/foundrystatus"
	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/profile"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrykit/backoff"
)

// statusTimeout bounds the live status probe against the local Foundry server.
const statusTimeout = 2 * time.Second

var _ dashboard.Switcher = (*Runner)(nil)

// Runner runs the Foundry process, handles backoff restarts, and applies
// profile switches requested via the dashboard.
type Runner struct {
	mu         sync.RWMutex
	state      State
	activator  Activator
	backoffCfg backoff.Config
	cfg        config.Config
	logger     *slog.Logger
	ctrl       *controller.SwitchController
	status     *foundrystatus.Client
	versions   dashboard.VersionManager
}

// New creates a Runner ready to run. The dashboard is started internally when
// Run is called. initialActive, if non-empty, is pre-set as the active profile
// name so the dashboard reports the correct state from the first request.
func New(
	initial State,
	initialActive string,
	activator Activator,
	versions dashboard.VersionManager,
	cfg config.Config,
	backoffCfg backoff.Config,
	logger *slog.Logger,
) *Runner {
	ctrl := controller.New()
	if initialActive != "" {
		ctrl.SetActive(initialActive)
	}
	return &Runner{
		state:      initial,
		activator:  activator,
		versions:   versions,
		cfg:        cfg,
		backoffCfg: backoffCfg,
		logger:     logger,
		ctrl:       ctrl,
		status:     foundrystatus.NewClient(&http.Client{Timeout: statusTimeout}),
	}
}

// Run starts the dashboard and the process loop. Blocks until clean shutdown.
// Returns the Foundry process exit code.
func (r *Runner) Run(ctx context.Context) int {
	dashCtx, cancelDash := context.WithCancel(ctx)
	var wg sync.WaitGroup
	wg.Go(func() {
		errCh := dashboard.Start(dashCtx, r.logger, r.cfg.DashboardAddr, r, r.versions, r)
		if err := <-errCh; err != nil {
			r.logger.Error("dashboard server stopped unexpectedly", "err", err)
		}
	})

	code := r.profileLoop(ctx)
	cancelDash()
	wg.Wait()
	return code
}

// RequestSwitch validates and enqueues a profile switch from external callers
// (e.g. the dashboard HTTP handler).
func (r *Runner) RequestSwitch(name string) error {
	if _, ok := r.findProfile(name); !ok {
		return fmt.Errorf("unknown profile %q", name)
	}
	r.ctrl.RequestSwitch(name)
	return nil
}

// Active returns the name of the currently active profile (empty for base config).
func (r *Runner) Active() string {
	return r.ctrl.Active()
}

// Version returns the version string of the currently running Foundry instance.
func (r *Runner) Version() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.state.Version
}

// FoundryStatus probes the running Foundry server for its live status. It
// returns an error when the server is unreachable (for example between
// restarts), which callers treat as "no users connected".
func (r *Runner) FoundryStatus(ctx context.Context) (foundrystatus.Status, error) {
	r.mu.RLock()
	port := r.state.Port
	r.mu.RUnlock()

	ctx, cancel := context.WithTimeout(ctx, statusTimeout)
	defer cancel()
	// simplification: assumes no route prefix; local probes hit the bare port.
	// Upgrade path: thread Runtime.RoutePrefix into procloop.State if needed.
	return r.status.Fetch(ctx, fmt.Sprintf("http://127.0.0.1:%d", port))
}

func (r *Runner) currentProfiles() []profile.Profile {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.state.Profiles
}
