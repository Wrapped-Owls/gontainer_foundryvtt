package procloop

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/profile"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrykit/procspawn"
)

func (r *Runner) applySwitch(ctx context.Context) error {
	select {
	case name := <-r.ctrl.SwitchCh:
		p, ok := r.findProfile(name)
		if !ok {
			return fmt.Errorf("unknown profile %q", name)
		}
		newState, err := r.activator.Switch(ctx, r.logger, p)
		if err != nil {
			return fmt.Errorf("switch to %q: %w", name, err)
		}
		r.mu.Lock()
		newState.Profiles = r.state.Profiles
		r.state = newState
		r.mu.Unlock()
		r.ctrl.SetActive(name)
		if writeErr := r.profilesFile.WriteActive(name); writeErr != nil {
			r.logger.Warn("failed to persist active profile", "profile", name, "err", writeErr)
		}
		return nil
	default:
		return nil
	}
}

func (r *Runner) findProfile(name string) (profile.Profile, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return profile.ByName(r.state.Profiles, name)
}

func (r *Runner) buildSpec() procspawn.Spec {
	r.mu.RLock()
	s := r.state
	r.mu.RUnlock()
	return procspawn.Spec{
		Path:   s.JSRuntime.Path,
		Args:   BuildArgs(s),
		Dir:    s.InstallRoot,
		Stdout: io.MultiWriter(os.Stdout, r.logs),
		Stderr: io.MultiWriter(os.Stderr, r.logs),
	}
}
