package source

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/fourcery/internal/copytree"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/fourcery/internal/probe"
)

// folderSource installs from a pre-extracted directory.
type folderSource struct {
	path          string
	cachedVersion string
}

// NewFolder constructs a folderSource for the given absolute directory.
func NewFolder(path string) Source { return &folderSource{path: path} }

func (f *folderSource) Kind() Kind { return KindFolder }

func (f *folderSource) Describe() string { return "folder " + filepath.Base(f.path) }

func (f *folderSource) Probe(_ context.Context) (string, error) {
	if f.cachedVersion != "" {
		return f.cachedVersion, nil
	}
	if v, err := probe.Filename(filepath.Base(f.path)); err == nil {
		f.cachedVersion = v
		return v, nil
	}
	v, err := probe.Folder(f.path)
	if err != nil {
		if errors.Is(err, probe.ErrNoVersion) {
			return "", ErrVersionUnknown
		}
		return "", fmt.Errorf("folder probe: %w", err)
	}
	f.cachedVersion = v
	return v, nil
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
