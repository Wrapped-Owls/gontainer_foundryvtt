package source

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/fourcery/archive"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/fourcery/internal/probe"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/fourcery/version"
)

// zipSource installs from a local *.zip file.
type zipSource struct {
	path string
	// cachedVersion is filled on first successful Probe and reused.
	cachedVersion version.Version
}

// NewZip constructs a zipSource for the given absolute zip path.
func NewZip(path string) Source { return &zipSource{path: path} }

func (z *zipSource) Kind() Kind { return KindZip }

func (z *zipSource) Describe() string { return "zip " + filepath.Base(z.path) }

func (z *zipSource) Probe(_ context.Context) (version.Version, error) {
	if !z.cachedVersion.IsZero() {
		return z.cachedVersion, nil
	}
	if raw, err := probe.Filename(filepath.Base(z.path)); err == nil {
		z.cachedVersion = version.Parse(raw)
		return z.cachedVersion, nil
	}
	raw, err := probe.Zip(z.path)
	if err != nil {
		if errors.Is(err, probe.ErrNoVersion) {
			return version.Version{}, ErrVersionUnknown
		}
		return version.Version{}, fmt.Errorf("zip probe: %w", err)
	}
	z.cachedVersion = version.Parse(raw)
	return z.cachedVersion, nil
}

func (z *zipSource) Materialise(_ context.Context, dst string) (Result, error) {
	if z.path == "" {
		return Result{}, fmt.Errorf("%w: zip path", ErrEmptyInput)
	}
	if _, err := archive.Extract(z.path, dst); err != nil {
		return Result{}, fmt.Errorf("zip extract: %w", err)
	}
	v, _ := z.Probe(context.Background())
	return Result{Kind: KindZip, Version: v}, nil
}
