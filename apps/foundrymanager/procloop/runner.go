package procloop

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/config"
	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/internal/controller"
	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/internal/dashboard"
	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/profile"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrykit/backoff"
)

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
}

// New creates a Runner ready to run. The dashboard is started internally when
// Run is called.
func New(
	initial State,
	activator Activator,
	cfg config.Config,
	backoffCfg backoff.Config,
	logger *slog.Logger,
) *Runner {
	return &Runner{
		state:      initial,
		activator:  activator,
		cfg:        cfg,
		backoffCfg: backoffCfg,
		logger:     logger,
		ctrl:       controller.New(),
	}
}

// Run starts the dashboard and the process loop. Blocks until clean shutdown.
// Returns the Foundry process exit code.
func (r *Runner) Run(ctx context.Context) int {
	dashCtx, cancelDash := context.WithCancel(ctx)
	var wg sync.WaitGroup
	wg.Go(func() {
		errCh := dashboard.Start(dashCtx, r.logger, r.cfg.DashboardAddr, r.currentProfiles(), r)
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

func (r *Runner) currentProfiles() []profile.Profile {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.state.Profiles
}
