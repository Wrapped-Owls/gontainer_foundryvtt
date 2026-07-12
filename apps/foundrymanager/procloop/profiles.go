package procloop

import (
	"fmt"
	"slices"

	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/profile"
	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/profloader"
)

// ListProfiles returns a copy of the configured profiles.
func (r *Runner) ListProfiles() []profile.Profile {
	return r.currentProfiles()
}

// GetProfile returns the named profile and whether it exists.
func (r *Runner) GetProfile(name string) (profile.Profile, bool) {
	return r.findProfile(name)
}

// CreateProfile validates and appends a new profile, then persists the list.
// It returns profile.ErrInvalid for missing required fields and
// profile.ErrExists when a profile with the same name already exists.
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

// UpdateProfile merges the non-empty fields of p onto the existing profile named
// name and persists the result. It returns profile.ErrNotFound when absent.
func (r *Runner) UpdateProfile(name string, p profile.Profile) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	idx := r.indexOf(name)
	if idx < 0 {
		return fmt.Errorf("%w: %q", profile.ErrNotFound, name)
	}
	updated := slices.Clone(r.state.Profiles)
	applyOverrides(&updated[idx], p)
	return r.persistProfiles(updated)
}

// DeleteProfile removes the named profile and persists the list. It refuses to
// delete the active profile (profile.ErrInvalid) and returns profile.ErrNotFound
// when absent.
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

// persistProfiles writes profiles to disk and, on success, swaps them into the
// running state. The caller must hold r.mu.
func (r *Runner) persistProfiles(profiles []profile.Profile) error {
	if err := profloader.WriteProfiles(r.cfg.ProfilesFile, profiles); err != nil {
		return fmt.Errorf("persist profiles: %w", err)
	}
	r.state.Profiles = profiles
	return nil
}

// indexOf returns the position of the named profile, or -1. Caller holds r.mu.
func (r *Runner) indexOf(name string) int {
	for i := range r.state.Profiles {
		if r.state.Profiles[i].Name == name {
			return i
		}
	}
	return -1
}

// applyOverrides copies the non-empty fields of src onto dst, leaving the name
// and any omitted fields (including secrets) untouched.
func applyOverrides(dst *profile.Profile, src profile.Profile) {
	for _, f := range []struct {
		dst *string
		src string
	}{
		{&dst.Label, src.Label},
		{&dst.DataPath, src.DataPath},
		{&dst.AdminKey, src.AdminKey},
		{&dst.AdminPasswordSalt, src.AdminPasswordSalt},
		{&dst.Version, src.Version},
		{&dst.World, src.World},
		{&dst.ManifestPath, src.ManifestPath},
	} {
		if f.src != "" {
			*f.dst = f.src
		}
	}
}
