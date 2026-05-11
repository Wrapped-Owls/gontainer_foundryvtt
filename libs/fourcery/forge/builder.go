package forge

import (
	"errors"
	"log/slog"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/fourcery/source"
)

type Builder struct {
	installRoot string
	sources     []source.Source
	observer    Observer
	logger      *slog.Logger
}

func New(installRoot string) *Builder {
	return &Builder{installRoot: installRoot}
}

func (b *Builder) WithSources(srcs ...source.Source) *Builder {
	b.sources = srcs
	return b
}

func (b *Builder) WithLogger(l *slog.Logger) *Builder {
	b.logger = l
	return b
}

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
