package procloop

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/config"
	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/internal/controller"
	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/internal/dashboard"
	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/internal/foundrystatus"
	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/internal/logstore"
	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/profile"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrykit/backoff"
)

const statusTimeout = 2 * time.Second

// ErrNoSession reports that no Foundry session was running to act on.
var ErrNoSession = errors.New("procloop: no running session")

var (
	_ dashboard.Supervisor   = (*Runner)(nil)
	_ dashboard.ProfileStore = (*Runner)(nil)
	_ dashboard.LogReader    = (*Runner)(nil)
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
	status     *foundrystatus.Client
	versions   dashboard.VersionManager
	logs       *logstore.Store
}

// New seeds the dashboard's reported active profile from initialActive, if set.
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
		logs: logstore.New(
			logstore.DefaultBufferLines,
			logstore.DefaultEventBuffer,
			cfg.LogAlertPatterns,
		),
	}
}

// Run blocks until shutdown and returns the Foundry exit code.
func (r *Runner) Run(ctx context.Context) int {
	dashCtx, cancelDash := context.WithCancel(ctx)
	var wg sync.WaitGroup
	wg.Go(func() {
		errCh := dashboard.Start(dashCtx, r.logger, r.cfg.DashboardAddr, r, r.versions, r, r)
		if err := <-errCh; err != nil {
			r.logger.Error("dashboard server stopped unexpectedly", "err", err)
		}
	})

	code := r.profileLoop(ctx)
	cancelDash()
	wg.Wait()
	return code
}

func (r *Runner) RequestSwitch(name string) error {
	if _, ok := r.findProfile(name); !ok {
		return fmt.Errorf("unknown profile %q", name)
	}
	r.ctrl.RequestSwitch(name)
	return nil
}

// RequestRestart clears the failure history and cancels the session so the child
// respawns immediately, instead of forcing operators to cycle the active profile
// just to cut a backoff wait short.
func (r *Runner) RequestRestart() error {
	if !r.ctrl.RequestRestart() {
		return ErrNoSession
	}
	// The wait is already cut short; a history we cannot clear only costs the next
	// failure a longer delay, so it must not fail the request.
	if err := backoff.NewFromConfig(r.backoffCfg).Reset(); err != nil {
		r.logger.Warn("could not clear the backoff history", "err", err)
	}
	return nil
}

// Active returns the name of the currently active profile (empty for base config).
func (r *Runner) Active() string {
	return r.ctrl.Active()
}

func (r *Runner) Version() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.state.Version
}

// FoundryStatus errors when the server is unreachable, which callers treat as
// "no users connected".
func (r *Runner) FoundryStatus(ctx context.Context) (foundrystatus.Status, error) {
	r.mu.RLock()
	port := r.state.Port
	r.mu.RUnlock()

	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, statusTimeout)
	defer cancel()
	// simplification: assumes no route prefix; local probes hit the bare port.
	// Upgrade path: thread Runtime.RoutePrefix into procloop.State if needed.
	return r.status.Fetch(ctx, fmt.Sprintf("http://127.0.0.1:%d", port))
}

func (r *Runner) Logs(n int) []string { return r.logs.Tail(n) }

// Events returns detected error/crash events at or after cursor and the next cursor.
func (r *Runner) Events(cursor int) ([]logstore.Event, int) { return r.logs.EventsSince(cursor) }

func (r *Runner) currentProfiles() []profile.Profile {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.state.Profiles
}
