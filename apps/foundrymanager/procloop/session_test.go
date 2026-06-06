package procloop

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/internal/controller"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrykit/backoff"
)

func newTestRunner() *Runner {
	return &Runner{
		logger: slog.Default(),
		ctrl:   controller.New(),
	}
}

func TestHandleBackoff_kubernetes(t *testing.T) {
	r := newTestRunner()
	ctx := context.Background()
	dec := backoff.Decision{Mode: backoff.ModeKubernetes}
	switched, stop := r.handleBackoff(ctx, ctx, dec)
	if switched || !stop {
		t.Errorf("ModeKubernetes: got switched=%v stop=%v, want false/true", switched, stop)
	}
}

func TestHandleBackoff_backoffZeroDelay(t *testing.T) {
	r := newTestRunner()
	ctx := context.Background()
	dec := backoff.Decision{Mode: backoff.ModeBackoff, Delay: 0}
	switched, stop := r.handleBackoff(ctx, ctx, dec)
	if switched || !stop {
		t.Errorf("ModeBackoff delay=0: got switched=%v stop=%v, want false/true", switched, stop)
	}
}

func TestHandleBackoff_backoffCancelledBySwitch(t *testing.T) {
	r := newTestRunner()
	ctx := context.Background()
	profileCtx, cancel := context.WithCancelCause(ctx)

	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel(errors.New("switch"))
	}()

	dec := backoff.Decision{Mode: backoff.ModeBackoff, Delay: 5 * time.Second}
	switched, stop := r.handleBackoff(ctx, profileCtx, dec)
	if !switched || !stop {
		t.Errorf("switch cancel: got switched=%v stop=%v, want true/true", switched, stop)
	}
}
