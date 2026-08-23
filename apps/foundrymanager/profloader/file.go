package profloader

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/profile"
)

type profileFile struct {
	Active   string            `json:"active,omitempty"`
	Profiles []profile.Profile `json:"profiles"`
}

// mu serializes the read-modify-write cycles: profileLoop and HTTP handlers race.
var mu sync.Mutex

// FromFile treats a missing or empty file as no profiles, not an error: the image
// pre-creates it, and empty-as-malformed would fail activation before Foundry starts.
func FromFile(path string) (profiles []profile.Profile, active string, err error) {
	var stored []byte
	stored, err = os.ReadFile(path) //nolint:gosec // path is sourced from operator-controlled config
	if errors.Is(err, os.ErrNotExist) {
		return nil, "", nil
	}
	if err != nil {
		return nil, "", err
	}
	if len(bytes.TrimSpace(stored)) == 0 {
		return nil, "", nil
	}
	var f profileFile
	if err = json.Unmarshal(stored, &f); err != nil {
		return nil, "", err
	}
	return f.Profiles, f.Active, nil
}

// WriteProfiles preserves the recorded active name; it creates the file if absent.
func WriteProfiles(path string, profiles []profile.Profile) error {
	return mutateFile(path, func(f *profileFile) {
		f.Profiles = profiles
	})
}

// WriteActive preserves any existing profiles array; it creates the file if absent.
func WriteActive(path, name string) error {
	return mutateFile(path, func(f *profileFile) {
		f.Active = name
	})
}

// mutateFile refuses malformed existing content rather than discarding it.
func mutateFile(path string, apply func(f *profileFile)) error {
	mu.Lock()
	defer mu.Unlock()

	stored, err := os.ReadFile(path) //nolint:gosec // path is sourced from operator-controlled config
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("read profiles: %w", err)
	}

	var f profileFile
	if len(bytes.TrimSpace(stored)) > 0 {
		if err = json.Unmarshal(stored, &f); err != nil {
			return fmt.Errorf("unmarshal profiles: %w", err)
		}
	}
	apply(&f)

	out, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal profiles: %w", err)
	}
	return writeFileAtomic(path, append(out, '\n'))
}

// writeFileAtomic renames over path so readers never see a truncated write.
func writeFileAtomic(path string, content []byte) error {
	staged, err := os.CreateTemp(filepath.Dir(path), ".profiles-*.tmp")
	if err != nil {
		return fmt.Errorf("create temp profiles file: %w", err)
	}
	stagedName := staged.Name()

	if _, err = staged.Write(content); err != nil {
		_ = staged.Close()
		_ = os.Remove(stagedName)
		return fmt.Errorf("write temp profiles file: %w", err)
	}
	if err = staged.Close(); err != nil {
		_ = os.Remove(stagedName)
		return fmt.Errorf("close temp profiles file: %w", err)
	}
	if err = os.Rename(stagedName, path); err != nil {
		_ = os.Remove(stagedName)
		return fmt.Errorf("rename temp profiles file: %w", err)
	}
	return nil
}
