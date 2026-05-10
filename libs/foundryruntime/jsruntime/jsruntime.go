package jsruntime

import (
	"fmt"
	"os/exec"
)

func Resolve(c Config, lookPath func(string) (string, error)) (Runtime, error) {
	if lookPath == nil {
		lookPath = exec.LookPath
	}
	kind := c.Kind
	if kind == "" {
		kind = Default
	}
	switch kind {
	case Bun, Node:
	default:
		return Runtime{}, fmt.Errorf("%w: %q", ErrUnsupported, kind)
	}
	if c.Path != "" {
		return Runtime{Kind: kind, Path: c.Path}, nil
	}
	p, err := lookPath(string(kind))
	if err != nil {
		return Runtime{}, fmt.Errorf("jsruntime: locate %s: %w", kind, err)
	}
	return Runtime{Kind: kind, Path: p}, nil
}
