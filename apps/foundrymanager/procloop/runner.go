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

type Runner struct {
	mu         sync.RWMutex
	state      State
	activator  Activator
	backoffCfg backoff.Config
	cfg        config.Config
	logger     *slog.Logger
	ctrl       *controller.SwitchController
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

func (r *Runner) currentProfiles() []profile.Profile {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.state.Profiles
}
