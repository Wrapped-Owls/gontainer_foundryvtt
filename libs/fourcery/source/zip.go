package source

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/fourcery/archive"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/fourcery/internal/probe"
)

type zipSource struct {
	path          string
	cachedVersion string
}

func NewZip(path string) Source { return &zipSource{path: path} }

func (z *zipSource) Kind() Kind { return KindZip }

func (z *zipSource) Describe() string { return "zip " + filepath.Base(z.path) }

func (z *zipSource) Probe(_ context.Context) (string, error) {
	if z.cachedVersion != "" {
		return z.cachedVersion, nil
	}
	if v, err := probe.Filename(filepath.Base(z.path)); err == nil {
		z.cachedVersion = v
		return v, nil
	}
	v, err := probe.Zip(z.path)
	if err != nil {
		if errors.Is(err, probe.ErrNoVersion) {
			return "", ErrVersionUnknown
		}
		return "", fmt.Errorf("zip probe: %w", err)
	}
	z.cachedVersion = v
	return v, nil
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
