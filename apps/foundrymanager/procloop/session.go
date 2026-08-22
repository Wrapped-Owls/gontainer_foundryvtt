package procloop

import (
	"context"
	"time"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrykit/backoff"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrykit/procspawn"
)

// profileLoop drives the outer loop: one session per profile, retrying on switch.
func (r *Runner) profileLoop(ctx context.Context) int {
	for {
		code, switched := r.runSession(ctx)
		if !switched {
			return code
		}
		if err := r.applySwitch(ctx); err != nil {
			r.logger.Error("profile switch failed, resuming with current config", "err", err)
		}
	}
}

func (r *Runner) runSession(ctx context.Context) (int, bool) {
	profileCtx, cancelProfile := context.WithCancelCause(ctx)
	defer cancelProfile(nil)
	r.ctrl.SetCancel(cancelProfile)
	return r.restartLoop(ctx, profileCtx)
}

// restartLoop returns (exitCode, switched): switched is true when the session
// ended for a reason that warrants re-entering the profile loop.
func (r *Runner) restartLoop(ctx, profileCtx context.Context) (int, bool) {
	mgr := backoff.NewFromConfig(r.backoffCfg)
	for {
		startedAt := time.Now()
		code, err := procspawn.Run(profileCtx, r.buildSpec())
		if err != nil {
			// A missing or non-executable binary cannot be fixed by respawning here.
			r.logger.Error("child failed to start", "err", err)
			r.logs.RecordCrash(code)
			return 1, false
		}
		if ctx.Err() != nil {
			r.logger.Info("shutdown requested; exiting", "exit_code", code)
			return code, false
		}
		if profileCtx.Err() != nil {
			return code, true
		}
		r.logger.Info("child exited", "exit_code", code)
		if code != 0 {
			r.logs.RecordCrash(code)
		}
		dec := mgr.OnFailure(code, time.Since(startedAt))
		switched, shouldStop := r.handleBackoff(ctx, profileCtx, dec)
		if shouldStop {
			return code, switched
		}
	}
}

// handleBackoff applies the backoff decision. Stopping exits the whole supervisor;
// see docs/rules/supervision.md for when that is the right answer.
func (r *Runner) handleBackoff(
	ctx, profileCtx context.Context,
	dec backoff.Decision,
) (switched, shouldStop bool) {
	if dec.Mode == backoff.ModeKubernetes {
		return false, true
	}
	if dec.IsExhausted() {
		r.logger.Error(
			"restart budget exhausted; exiting so the container can be recreated",
			"consecutive_failures", dec.State.ConsecutiveFailures,
		)
		return false, true
	}
	if dec.Delay > 0 {
		r.logger.Info(
			"backoff",
			"delay", dec.Delay,
			"consecutive_failures", dec.State.ConsecutiveFailures,
		)
		if err := backoff.Sleep(profileCtx, dec.Delay); err != nil {
			if ctx.Err() != nil {
				return false, true
			}
			r.logger.Info("backoff cut short", "cause", context.Cause(profileCtx))
			return true, true
		}
	}
	return false, false
}
