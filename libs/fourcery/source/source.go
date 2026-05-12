package source

import (
	"context"
	"errors"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/fourcery/version"
)

// Kind enumerates the concrete Source implementations.
type Kind string

const (
	KindURL     Kind = "url"
	KindSession Kind = "session"
	KindZip     Kind = "zip"
	KindFolder  Kind = "folder"
)

// Source is the strategy interface for one way of obtaining a Foundry
// install. Implementations must be stateless with respect to the
// destination directory: Materialise must work with any empty dst
// handed to it by the orchestrator.
type Source interface {
	// Kind returns the Source's concrete kind, used by the resolver
	// for preference ordering.
	Kind() Kind
	// Describe returns a human-readable label safe for logs (no
	// credentials, no presigned URLs).
	Describe() string
	// Probe returns the version this Source would install. It must
	// not write to disk. Returns ErrVersionUnknown when the version
	// cannot be determined without materialising.
	Probe(ctx context.Context) (version.Version, error)
	// Materialise places the install tree into dst. dst must already
	// exist and be empty. The returned Result reports the version
	// actually written (which may differ from a Probe result for
	// sources whose version is only knowable post-fetch).
	Materialise(ctx context.Context, dst string) (Result, error)
}

// Result reports the outcome of a successful Materialise call.
type Result struct {
	Version version.Version
	Kind    Kind
}

// Sentinel errors.
var (
	// ErrVersionUnknown is returned by Probe when the implementation
	// cannot determine a version without fetching/extracting.
	ErrVersionUnknown = errors.New("source: version not determinable without materialising")
	// ErrNoMatch is returned by callers when no Source satisfies the
	// requested version.
	ErrNoMatch = errors.New("source: nothing matches the request")
	// ErrEmptyInput is returned by Materialise when a Source requires
	// input (URL, credentials) that is missing.
	ErrEmptyInput = errors.New("source: required input is empty")
)
