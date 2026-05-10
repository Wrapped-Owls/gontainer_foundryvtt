package step

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundryruntime/jsruntime"
)

type jsRuntimeStep struct{}

// JSRuntime returns a Step that resolves the JavaScript runtime (Node.js or Bun).
func JSRuntime() Step { return jsRuntimeStep{} }

func (jsRuntimeStep) Apply(_ context.Context, s *State, _ *slog.Logger) error {
	jsCfg := jsruntime.DefaultConfig()
	if err := jsruntime.LoadFromEnv(&jsCfg); err != nil {
		return fmt.Errorf("load js runtime config: %w", err)
	}
	rt, err := jsruntime.Resolve(jsCfg, nil)
	if err != nil {
		return fmt.Errorf("resolve js runtime: %w", err)
	}
	s.JSRuntime = rt
	return nil
}
