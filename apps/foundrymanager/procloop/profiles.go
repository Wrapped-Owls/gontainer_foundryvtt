package procloop

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"

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
// AdminKey, AdminPasswordSalt and ManifestPath are cleared: as in applyOverrides,
// they come from file/env only.
func (r *Runner) CreateProfile(p profile.Profile) error {
	if p.Name == "" || p.DataPath == "" {
		return fmt.Errorf("%w: name and dataPath are required", profile.ErrInvalid)
	}
	if !filepath.IsAbs(p.DataPath) || hasParentSegment(p.DataPath) {
		return fmt.Errorf("%w: dataPath must be an absolute path with no .. segment", profile.ErrInvalid)
	}
	p.AdminKey = ""
	p.AdminPasswordSalt = ""
	p.ManifestPath = ""
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.indexOf(p.Name) >= 0 {
		return fmt.Errorf("%w: %q", profile.ErrExists, p.Name)
	}
	return r.persistProfiles(append(slices.Clone(r.state.Profiles), p))
}

// UpdateProfile merges the non-empty fields of p onto the existing profile named
// name and persists the result. It returns profile.ErrNotFound when absent. If
// name is the active profile and its version or world actually changed, the
// edit is applied to a stale, already-running session; queue a switch so the
// new version/world takes effect instead of silently going inert.
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
	return slices.IndexFunc(r.state.Profiles, func(p profile.Profile) bool { return p.Name == name })
}

// hasParentSegment finds a ".." element, not the substring inside a longer name.
func hasParentSegment(path string) bool {
	return slices.Contains(strings.Split(path, string(filepath.Separator)), "..")
}

// applyOverrides copies the editable non-empty fields of src onto dst. Only the
// label, version and world are editable; the data location, manifest path and
// admin secrets are deliberately immutable here so an edit cannot repoint a
// profile at another disk or change credentials - those come from file/env only.
func applyOverrides(dst *profile.Profile, src profile.Profile) {
	for _, f := range []struct {
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
