package procloop

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/profile"
	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/profloader"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrykit/procspawn"
)

// applySwitch reads the pending switch request and updates r.state.
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
		r.state = newState
		r.mu.Unlock()
		r.ctrl.SetActive(name)
		if err := profloader.WriteActive(r.cfg.ProfilesFile, name); err != nil {
			r.logger.Warn("failed to persist active profile", "profile", name, "err", err)
		}
		return nil
	default:
		return nil
	}
}

func (r *Runner) findProfile(name string) (profile.Profile, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, p := range r.state.Profiles {
		if p.Name == name {
			return p, true
		}
	}
	return profile.Profile{}, false
}

func (r *Runner) buildSpec() procspawn.Spec {
	r.mu.RLock()
	s := r.state
	r.mu.RUnlock()
	mainScript := filepath.Join(s.InstallRoot, s.MainScript)
	return procspawn.Spec{
		Path: s.JSRuntime.Path,
		Args: BuildArgs(s.JSRuntime.Kind, mainScript, s.DataPath, s.Port),
		Dir:  s.InstallRoot,
	}
}
