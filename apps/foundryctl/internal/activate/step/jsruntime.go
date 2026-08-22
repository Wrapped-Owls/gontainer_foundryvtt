package step

import (
	"context"
	"fmt"
	"log/slog"
	"os/exec"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundryruntime/jsruntime"
)

type jsRuntimeStep struct {
	lookPath func(string) (string, error)
}

func JSRuntime() Step { return jsRuntimeStep{lookPath: exec.LookPath} }

func (js jsRuntimeStep) Apply(_ context.Context, s *State, logger *slog.Logger) error {
	jsCfg := jsruntime.DefaultConfig()
	if err := jsruntime.LoadFromEnv(&jsCfg); err != nil {
		return fmt.Errorf("load js runtime config: %w", err)
	}
	rt, err := jsruntime.Resolve(
		jsCfg,
		jsruntime.FoundryMajor(s.Install.Version.Major()),
		js.lookPath,
	)
	if err != nil {
		return fmt.Errorf("resolve js runtime: %w", err)
	}
	s.JSRuntime = rt
	logger.Info(
		"js runtime selected",
		"kind", rt.Kind,
		"path", rt.Path,
		"foundry_version", s.Install.Version.String(),
	)
	return nil
}
