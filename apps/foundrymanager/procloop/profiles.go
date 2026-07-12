package procloop

import (
	"fmt"
	"slices"

	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/profile"
	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/profloader"
)

func (r *Runner) ListProfiles() []profile.Profile {
	return r.currentProfiles()
}

func (r *Runner) GetProfile(name string) (profile.Profile, bool) {
	return r.findProfile(name)
}

func (r *Runner) CreateProfile(p profile.Profile) error {
	if p.Name == "" || p.DataPath == "" {
		return fmt.Errorf("%w: name and dataPath are required", profile.ErrInvalid)
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.indexOf(p.Name) >= 0 {
		return fmt.Errorf("%w: %q", profile.ErrExists, p.Name)
	}
	return r.persistProfiles(append(slices.Clone(r.state.Profiles), p))
}

func (r *Runner) UpdateProfile(name string, p profile.Profile) error {
	r.mu.Lock()
	idx := r.indexOf(name)
	if idx < 0 {
		r.mu.Unlock()
		return fmt.Errorf("%w: %q", profile.ErrNotFound, name)
	}
	before := r.state.Profiles[idx]
	updated := slices.Clone(r.state.Profiles)
	applyOverrides(&updated[idx], p)
	after := updated[idx]
	err := r.persistProfiles(updated)
	r.mu.Unlock()
	if err != nil {
		return err
	}
	if name == r.ctrl.Active() && (before.Version != after.Version || before.World != after.World) {
		r.ctrl.RequestSwitch(name)
	}
	return nil
}

func (r *Runner) DeleteProfile(name string) error {
	if name == r.ctrl.Active() {
		return fmt.Errorf("%w: cannot delete the active profile", profile.ErrInvalid)
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	idx := r.indexOf(name)
	if idx < 0 {
		return fmt.Errorf("%w: %q", profile.ErrNotFound, name)
	}
	return r.persistProfiles(slices.Delete(slices.Clone(r.state.Profiles), idx, idx+1))
}

func (r *Runner) persistProfiles(profiles []profile.Profile) error {
	if err := profloader.WriteProfiles(r.cfg.ProfilesFile, profiles); err != nil {
		return fmt.Errorf("persist profiles: %w", err)
	}
	r.state.Profiles = profiles
	return nil
}

func (r *Runner) indexOf(name string) int {
	for i := range r.state.Profiles {
		if r.state.Profiles[i].Name == name {
			return i
		}
	}
	return -1
}

func applyOverrides(dst *profile.Profile, src profile.Profile) {
	for _, f := range []struct { // only label/version/world editable; disk path and secrets are not
		dst *string
		src string
	}{
		{&dst.Label, src.Label},
		{&dst.Version, src.Version},
		{&dst.World, src.World},
	} {
		if f.src != "" {
			*f.dst = f.src
		}
	}
}
