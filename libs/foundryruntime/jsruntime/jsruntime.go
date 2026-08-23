package jsruntime

import (
	"cmp"
	"fmt"
)

// FoundryMajor is the resolved Foundry semver major; 0 means unparsed.
type FoundryMajor uint64

// Resolve picks the binary for c.Kind. See docs/rules/supervision.md for the rule
// that ties a Node major to a Foundry major.
func Resolve(
	c Config,
	foundry FoundryMajor,
	lookPath func(string) (string, error),
) (Runtime, error) {
	kind := cmp.Or(c.Kind, Default)
	if kind != Bun && kind != Node {
		return Runtime{}, fmt.Errorf("%w: %q", ErrUnsupported, kind)
	}
	if c.Path != "" {
		return Runtime{Kind: kind, Path: c.Path}, nil
	}

	name, err := binaryName(kind, foundry)
	if err != nil {
		return Runtime{}, err
	}
	path, err := lookPath(name)
	if err != nil {
		return Runtime{}, fmt.Errorf(
			"%w: %q is not on PATH; this image ships bun, node22 and node24, "+
				"or set %s to an explicit binary",
			ErrNotFound, name, envJSRuntimePath,
		)
	}
	return Runtime{Kind: kind, Path: path}, nil
}

const (
	binNode22 = "node22"
	binNode24 = "node24"
)

type nodeFloor struct {
	minFoundry FoundryMajor
	binary     string
}

// nodeFloors is checked newest-first; 0 matches nothing and errors rather than guess.
// Extend with a row, never a branch.
//
//nolint:mnd // the Foundry majors are the table's content
var nodeFloors = [...]nodeFloor{
	{minFoundry: 14, binary: binNode24},
	{minFoundry: 1, binary: binNode22},
}

func binaryName(kind Kind, foundry FoundryMajor) (string, error) {
	if kind != Node {
		return string(kind), nil
	}
	for _, floor := range nodeFloors {
		if foundry >= floor.minFoundry {
			return floor.binary, nil
		}
	}
	return "", fmt.Errorf(
		"%w: set FOUNDRY_VERSION to a semver version, or %s to an explicit binary",
		ErrUnknownFoundry, envJSRuntimePath,
	)
}
