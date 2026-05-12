package source

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// Registry is the factory that turns a Config + filesystem state into
// the ordered list of Source values fourcery will consider. The order
// is deterministic for testability: folders alphabetical, zips
// alphabetical, then session (if credentials), then URL (if set).
type Registry struct {
	cfg  Config
	opts Options
}

// NewRegistry constructs a Registry. All Options are optional; sane
// defaults apply.
func NewRegistry(cfg Config, opts ...Option) *Registry {
	r := &Registry{cfg: cfg}
	for _, opt := range opts {
		opt(&r.opts)
	}
	return r
}

// Enumerate walks the configured sources directory and combines the
// discovered local sources with credential-derived ones. Errors are
// returned for unexpected I/O failures only; an absent sources
// directory is treated as "no local sources".
func (r *Registry) Enumerate(_ context.Context) ([]Source, error) {
	var folders, zips []Source
	if r.cfg.SourcesDir != "" {
		f, z, err := scanSources(r.cfg.SourcesDir)
		if err != nil {
			return nil, err
		}
		folders, zips = f, z
	}
	out := make([]Source, 0, len(folders)+len(zips)+2)
	out = append(out, folders...)
	out = append(out, zips...)
	if r.cfg.Version != "" && r.hasAuth() {
		out = append(out, NewSession(
			r.cfg.Version, r.cfg.Session, r.cfg.Username, r.cfg.Password,
		))
	}
	if r.cfg.ReleaseURL != "" {
		out = append(
			out,
			NewURL(r.cfg.ReleaseURL, r.opts.HTTPClient, r.cfg.Version, r.cfg.SourcesDir),
		)
	}
	return out, nil
}

func (r *Registry) hasAuth() bool {
	if r.cfg.Session != "" {
		return true
	}
	return r.cfg.Username != "" && r.cfg.Password != ""
}

func scanSources(dir string) ([]Source, []Source, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, nil
		}
		return nil, nil, fmt.Errorf("source scan %s: %w", dir, err)
	}
	var folderNames, zipNames []string
	for _, e := range entries {
		name := e.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}
		switch {
		case e.IsDir():
			folderNames = append(folderNames, name)
		case strings.EqualFold(filepath.Ext(name), ".zip"):
			zipNames = append(zipNames, name)
		}
	}
	slices.Sort(folderNames)
	slices.Sort(zipNames)
	folders := make([]Source, 0, len(folderNames))
	for _, n := range folderNames {
		folders = append(folders, NewFolder(filepath.Join(dir, n)))
	}
	zips := make([]Source, 0, len(zipNames))
	for _, n := range zipNames {
		zips = append(zips, NewZip(filepath.Join(dir, n)))
	}
	return folders, zips, nil
}
