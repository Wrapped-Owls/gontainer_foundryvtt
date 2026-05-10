package jsruntime

import (
	"errors"
	"fmt"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrykit/confloader"
)

type Kind string

const (
	Bun  Kind = "bun"
	Node Kind = "node"
)

const Default = Bun

const (
	envJSRuntime     = "FOUNDRY_JS_RUNTIME"
	envJSRuntimePath = "FOUNDRY_JS_RUNTIME_PATH"
)

var ErrUnsupported = errors.New("jsruntime: unsupported runtime kind")

type Config struct {
	Kind Kind
	Path string
}

type Runtime struct {
	Kind Kind
	Path string
}

func DefaultConfig() Config {
	return Config{Kind: Default}
}

func LoadFromEnv(c *Config) error {
	return confloader.BindEnv(
		confloader.BindField(&c.Kind, envJSRuntime, parseKind),
		confloader.BindField(&c.Path, envJSRuntimePath, nil),
	)
}

func parseKind(v string) (Kind, error) {
	k := Kind(v)
	switch k {
	case Bun, Node:
		return k, nil
	default:
		return "", fmt.Errorf("%w: %q", ErrUnsupported, v)
	}
}
