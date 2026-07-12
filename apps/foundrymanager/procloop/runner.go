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

const statusTimeout = 2 * time.Second

var _ dashboard.Switcher = (*Runner)(nil)

type Runner struct {
	mu         sync.RWMutex
	state      State
	activator  Activator
	backoffCfg backoff.Config
	cfg        config.Config
	logger     *slog.Logger
	ctrl       *controller.SwitchController
	status     *foundrystatus.Client
}

type Params struct {
	Initial       State
	InitialActive string
	Activator     Activator
	Config        config.Config
	Backoff       backoff.Config
	Logger        *slog.Logger
}

func New(params Params) *Runner {
	ctrl := controller.New()
	if params.InitialActive != "" {
		ctrl.SetActive(params.InitialActive)
	}
	return &Runner{
		state:      params.Initial,
		activator:  params.Activator,
		cfg:        params.Config,
		backoffCfg: params.Backoff,
		logger:     params.Logger,
		ctrl:       ctrl,
		status:     foundrystatus.NewClient(&http.Client{Timeout: statusTimeout}),
	}
}

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

func (r *Runner) RequestSwitch(name string) error {
	if _, ok := r.findProfile(name); !ok {
		return fmt.Errorf("unknown profile %q", name)
	}
	r.ctrl.RequestSwitch(name)
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

func (r *Runner) currentProfiles() []profile.Profile {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.state.Profiles
}
