package procloop

import (
	"context"

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

// runSession supervises one profile session: registers the cancel function and
// delegates to the restart loop.
func (r *Runner) runSession(ctx context.Context) (int, bool) {
	profileCtx, cancelProfile := context.WithCancelCause(ctx)
	defer cancelProfile(nil)
	r.ctrl.SetCancel(cancelProfile)
	return r.restartLoop(ctx, profileCtx)
}

// restartLoop is the backoff restart loop for the current profile session.
// Returns (exitCode, true) when a switch was requested, (exitCode, false) on
// clean shutdown or fatal error.
func (r *Runner) restartLoop(ctx, profileCtx context.Context) (int, bool) {
	mgr := backoff.NewFromConfig(r.backoffCfg)
	for {
		code, err := procspawn.Run(profileCtx, r.buildSpec())
		if err != nil {
			r.logger.Error("child failed to start", "err", err)
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
		dec, decErr := mgr.OnFailure(code)
		if decErr != nil {
			r.logger.Error("backoff state failed", "err", decErr)
			return code, false
		}
		switched, earlyExit := r.handleBackoff(ctx, profileCtx, dec)
		if earlyExit {
			return code, switched
		}
	}
}

// handleBackoff applies the backoff decision. Returns (switched, stop): if
// stop is true the caller should return immediately with that switched value.
func (r *Runner) handleBackoff(
	ctx, profileCtx context.Context,
	dec backoff.Decision,
) (switched, stop bool) {
	switch dec.Mode {
	case backoff.ModeKubernetes:
		return false, true
	case backoff.ModeNoCache:
		<-ctx.Done()
		return false, true
	case backoff.ModeBackoff:
		if dec.Delay == 0 {
			return false, true
		}
		r.logger.Info(
			"backoff",
			"delay", dec.Delay,
			"consecutive_failures", dec.State.ConsecutiveFailures,
		)
		if err := backoff.Sleep(profileCtx, dec.Delay); err != nil {
			if ctx.Err() != nil {
				return false, true
			}
			return true, true
		}
	}
	return false, false
}
