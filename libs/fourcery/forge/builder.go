package forge

import (
	"errors"
	"log/slog"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/fourcery/source"
)

// Builder constructs a Forge via a fluent API. All With* methods
// return the receiver so calls can be chained; Build returns the
// configured Forge (or an error when required inputs are missing).
type Builder struct {
	installRoot string
	sources     []source.Source
	observer    Observer
	logger      *slog.Logger
}

// New starts a Builder rooted at installRoot. installRoot must be a
// non-empty absolute path; an error is returned at Build time
// otherwise.
func New(installRoot string) *Builder {
	return &Builder{installRoot: installRoot}
}

// WithSources sets the candidate sources the resolver may pick from.
// Subsequent calls replace the previous slice.
func (b *Builder) WithSources(srcs ...source.Source) *Builder {
	b.sources = srcs
	return b
}

// WithLogger sets a logger.
func (b *Builder) WithLogger(l *slog.Logger) *Builder {
	b.logger = l
	return b
}

// Build returns the configured Forge.
func (b *Builder) Build() (*Forge, error) {
	if b.installRoot == "" {
		return nil, errors.New("forge: install root is required")
	}
	obs := b.observer
	if obs == nil {
		if b.logger != nil {
			obs = SlogObserver{Logger: b.logger}
		} else {
			obs = noopObserver{}
		}
	}
	return &Forge{
		installRoot: b.installRoot,
		sources:     b.sources,
		observer:    obs,
		resolver:    NewResolver(b.installRoot),
	}, nil
}
