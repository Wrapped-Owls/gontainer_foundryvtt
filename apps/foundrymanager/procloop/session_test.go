package procloop

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"testing/synctest"
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

func TestHandleBackoff(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		decision      backoff.Decision
		cancelParent  bool
		cancelProfile bool
		wantSwitched  bool
		wantStop      bool
	}{
		{
			name:     "kubernetes bypass exits so crashloopbackoff throttles",
			decision: backoff.Decision{Mode: backoff.ModeKubernetes},
			wantStop: true,
		},
		{
			name: "restart budget exhausted exits for a container recreate",
			decision: backoff.Decision{
				Mode:  backoff.ModeBackoff,
				Delay: time.Second,
				State: backoff.State{ConsecutiveFailures: backoff.MaxConsecutiveFailures},
			},
			wantStop: true,
		},
		{
			name: "first failure respawns immediately instead of exiting",
			decision: backoff.Decision{
				Mode:  backoff.ModeBackoff,
				Delay: 0,
				State: backoff.State{ConsecutiveFailures: 1},
			},
		},
		{
			name: "unwritable cache respawns instead of blocking forever",
			decision: backoff.Decision{
				Mode:  backoff.ModeNoCache,
				Delay: 0,
				State: backoff.State{ConsecutiveFailures: 1},
			},
		},
		{
			name: "switch during the wait restarts with the new profile",
			decision: backoff.Decision{
				Mode:  backoff.ModeBackoff,
				Delay: time.Hour,
				State: backoff.State{ConsecutiveFailures: 3},
			},
			cancelProfile: true,
			wantSwitched:  true,
			wantStop:      true,
		},
		{
			name: "shutdown during the wait exits without switching",
			decision: backoff.Decision{
				Mode:  backoff.ModeBackoff,
				Delay: time.Hour,
				State: backoff.State{ConsecutiveFailures: 3},
			},
			cancelParent: true,
			wantStop:     true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancelParent := context.WithCancel(context.Background())
			defer cancelParent()
			profileCtx, cancelProfile := context.WithCancelCause(ctx)
			defer cancelProfile(nil)
			if testCase.cancelParent {
				cancelParent()
			}
			if testCase.cancelProfile {
				cancelProfile(errors.New("switch requested"))
			}

			switched, stop := newTestRunner().handleBackoff(ctx, profileCtx, testCase.decision)
			if switched != testCase.wantSwitched || stop != testCase.wantStop {
				t.Fatalf(
					"got switched=%v stop=%v, want switched=%v stop=%v",
					switched, stop, testCase.wantSwitched, testCase.wantStop,
				)
			}
		})
	}
}

func TestHandleBackoffWaitsOutTheDelayThenRespawns(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx := context.Background()
		dec := backoff.Decision{
			Mode:  backoff.ModeBackoff,
			Delay: backoff.MaxDelay,
			State: backoff.State{ConsecutiveFailures: 9},
		}

		switched, stop := newTestRunner().handleBackoff(ctx, ctx, dec)
		if switched || stop {
			t.Fatalf("got switched=%v stop=%v, want false/false", switched, stop)
		}
	})
}
