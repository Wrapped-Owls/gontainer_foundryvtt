package source

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/fourcery/internal/copytree"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/fourcery/internal/probe"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/fourcery/version"
)

// folderSource installs from a pre-extracted directory.
type folderSource struct {
	path          string
	cachedVersion version.Version
}

// NewFolder constructs a folderSource for the given absolute directory.
func NewFolder(path string) Source { return &folderSource{path: path} }

func (f *folderSource) Kind() Kind { return KindFolder }

func (f *folderSource) Describe() string { return "folder " + filepath.Base(f.path) }

func (f *folderSource) Probe(_ context.Context) (version.Version, error) {
	if !f.cachedVersion.IsZero() {
		return f.cachedVersion, nil
	}
	if raw, err := probe.Filename(filepath.Base(f.path)); err == nil {
		f.cachedVersion = version.Parse(raw)
		return f.cachedVersion, nil
	}
	raw, err := probe.Folder(f.path)
	if err != nil {
		if errors.Is(err, probe.ErrNoVersion) {
			return version.Version{}, ErrVersionUnknown
		}
		return version.Version{}, fmt.Errorf("folder probe: %w", err)
	}
	f.cachedVersion = version.Parse(raw)
	return f.cachedVersion, nil
}

func (f *folderSource) Materialise(_ context.Context, dst string) (Result, error) {
	if f.path == "" {
		return Result{}, fmt.Errorf("%w: folder path", ErrEmptyInput)
	}
	if err := copytree.Copy(f.path, dst); err != nil {
		return Result{}, fmt.Errorf("folder copy: %w", err)
	}
	v, _ := f.Probe(context.Background())
	return Result{Kind: KindFolder, Version: v}, nil
}
