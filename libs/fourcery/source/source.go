package source

import (
	"context"
	"errors"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/fourcery/version"
)

type Kind string

const (
	KindURL     Kind = "url"
	KindSession Kind = "session"
	KindZip     Kind = "zip"
	KindFolder  Kind = "folder"
)

type Source interface {
	Kind() Kind
	Describe() string
	Probe(ctx context.Context) (version.Version, error)
	Materialise(ctx context.Context, dst string) (Result, error)
}

type Result struct {
	Version version.Version
	Kind    Kind
}

var (
	ErrVersionUnknown = errors.New("source: version not determinable without materialising")

	ErrNoMatch = errors.New("source: nothing matches the request")

	ErrEmptyInput = errors.New("source: required input is empty")
)
