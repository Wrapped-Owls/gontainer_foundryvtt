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

var ErrNoSession = errors.New("procloop: no running session")

var (
	_ dashboard.Supervisor   = (*Runner)(nil)
	_ dashboard.ProfileStore = (*Runner)(nil)
	_ dashboard.LogReader    = (*Runner)(nil)
)

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

type Params struct {
	Initial       State
	InitialActive string
	Activator     Activator
	Versions      dashboard.VersionManager
	Config        config.Config
	Backoff       backoff.Config
	Logger        *slog.Logger
}

func New(params Params) *Runner {
	ctrl := controller.New()
	if params.InitialActive != "" {
		ctrl.SetActive(params.InitialActive)
	}
	cfg := params.Config
	return &Runner{
		state:      params.Initial,
		activator:  params.Activator,
		versions:   params.Versions,
		cfg:        cfg,
		backoffCfg: params.Backoff,
		logger:     params.Logger,
		ctrl:       ctrl,
		status:     foundrystatus.NewClient(&http.Client{Timeout: statusTimeout}),
		logs: logstore.New(
			logstore.DefaultBufferLines,
			logstore.DefaultEventBuffer,
			cfg.LogAlertPatterns,
		),
	}
}

func (r *Runner) Run(ctx context.Context) int {
	dashCtx, cancelDash := context.WithCancel(ctx)
	var wg sync.WaitGroup
	wg.Go(func() {
		errCh := dashboard.Start(dashCtx, dashboard.Params{
			Logger:     r.logger,
			Addr:       r.cfg.DashboardAddr,
			Supervisor: r,
			Versions:   r.versions,
			Profiles:   r,
			Logs:       r,
		})
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

func (r *Runner) RequestRestart() error {
	if !r.ctrl.RequestRestart() {
		return ErrNoSession
	}
	if err := backoff.NewFromConfig(r.backoffCfg).Reset(); err != nil {
		r.logger.Warn("could not clear the backoff history", "err", err)
	}
	return nil
}

func (r *Runner) Active() string {
	return r.ctrl.Active()
}

func (r *Runner) Version() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.state.Version
}

func (r *Runner) FoundryStatus(ctx context.Context) (foundrystatus.Status, error) {
	r.mu.RLock()
	port := r.state.Port
	r.mu.RUnlock()

	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, statusTimeout)
	defer cancel()
	return r.status.Fetch(ctx, fmt.Sprintf("http://127.0.0.1:%d", port))
}

func (r *Runner) Logs(n int) []string { return r.logs.Tail(n) }

func (r *Runner) Events(cursor int) ([]logstore.Event, int) { return r.logs.EventsSince(cursor) }

func (r *Runner) currentProfiles() []profile.Profile {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.state.Profiles
}
