package manifest

import (
	"errors"
	"fmt"
	"os"

	"github.com/Masterminds/semver/v3"
	"gopkg.in/yaml.v3"
)

// Load parses a manifest from path. A missing file is treated as the
// empty manifest (no patches), matching the "no hotfixes by default"
// policy.
func Load(path string) (*File, error) {
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return &File{Version: SchemaVersion}, nil
	}
	if err != nil {
		return nil, err
	}
	return Parse(b)
}

// Parse decodes raw bytes into a validated *File.
func Parse(b []byte) (*File, error) {
	var f File
	if err := yaml.Unmarshal(b, &f); err != nil {
		return nil, fmt.Errorf("manifest: parse: %w", err)
	}
	if f.Version == 0 {
		f.Version = SchemaVersion
	}
	if err := f.Validate(); err != nil {
		return nil, err
	}
	return &f, nil
}

// Validate runs structural checks on the manifest.
func (f *File) Validate() error {
	if f.Version > SchemaVersion {
		return fmt.Errorf("%w: got %d, max %d", ErrUnsupportedSchema, f.Version, SchemaVersion)
	}
	for i, p := range f.Patches {
		if p.ID == "" {
			return fmt.Errorf("patch[%d]: %w", i, ErrEmptyID)
		}
		if p.Versions == "" {
			return fmt.Errorf("patch[%d] %q: %w", i, p.ID, ErrEmptyVersions)
		}
		if _, err := semver.NewConstraint(p.Versions); err != nil {
			return fmt.Errorf("patch[%d] %q: %w: %v", i, p.ID, ErrInvalidConstraint, err)
		}
		for j, a := range p.Actions {
			if err := validateAction(a); err != nil {
				return fmt.Errorf("patch[%d] %q action[%d]: %w", i, p.ID, j, err)
			}
		}
	}
	return nil
}

func validateAction(a Action) error {
	if a.Dest == "" {
		return ErrMissingDest
	}
	switch a.Type {
	case ActionDownload, ActionZipOverlay:
		if a.URL == "" || a.SHA256 == "" {
			return ErrDownloadNeedsURL
		}
	case ActionFileReplace:
		// Content may legitimately be empty (truncate file).
	default:
		return fmt.Errorf("%w: %q", ErrUnknownAction, a.Type)
	}
	return nil
}
