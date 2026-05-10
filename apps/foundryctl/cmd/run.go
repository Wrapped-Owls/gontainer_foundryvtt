package cmd

import (
	"context"
	"errors"
	"log/slog"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundryctl/internal/activate"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrykit/backoff"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrykit/procspawn"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundryruntime/jsruntime"
)

func Run(_ []string, logger *slog.Logger) int {
	ctx, cancel := context.WithCancelCause(context.Background())
	defer cancel(nil)
	ctx, stop := signalNotifyContext(ctx, syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	state, err := activate.Prepare(ctx, logger)
	if err != nil {
		logger.Error("activation failed", "err", err, "install_root", state.App.Paths.InstallRoot)
		return 1
	}
	startHealthServer(ctx, logger, state.App.Paths.HealthAddr, state.Runtime.Port)
	logger.Info("js runtime selected", "kind", state.JSRuntime.Kind, "path", state.JSRuntime.Path)

	mgr := backoff.NewFromConfig(state.App.Backoff)
	mainScript := filepath.Join(state.Install.Root, state.App.Paths.MainScript)
	for {
		spec := procspawn.Spec{
			Path: state.JSRuntime.Path,
			Args: runtimeArgs(
				state.JSRuntime.Kind,
				mainScript,
				state.App.Paths.DataPath,
				state.Runtime.Port,
			),
			Dir: state.Install.Root,
		}
		logger.Info("starting foundry", "argv", spec.Args, "dir", spec.Dir)

		var code int
		if code, err = procspawn.Run(ctx, spec); err != nil {
			logger.Error("child failed to start", "err", err)
			return 1
		}
		if errors.Is(ctx.Err(), context.Canceled) {
			logger.Info("shutdown requested; exiting", "exit_code", code)
			return code
		}
		logger.Info("child exited", "exit_code", code)
		dec, err := mgr.OnFailure(code)
		if err != nil {
			logger.Error("backoff state failed", "err", err)
			return code
		}
		switch dec.Mode {
		case backoff.ModeKubernetes:
			return code
		case backoff.ModeNoCache:
			<-ctx.Done()
			return code
		case backoff.ModeBackoff:
			if dec.Delay == 0 {
				return code
			}
			logger.Info(
				"backoff",
				"delay",
				dec.Delay,
				"consecutive_failures",
				dec.State.ConsecutiveFailures,
			)
			if err = backoff.Sleep(ctx, dec.Delay); err != nil {
				return code
			}
		}
	}
}

func runtimeArgs(kind jsruntime.Kind, mainScript, dataPath string, port int) []string {
	args := []string{
		mainScript,
		"--dataPath=" + dataPath,
		"--port=" + strconv.Itoa(port),
	}
	if kind == jsruntime.Bun {
		return append([]string{"run"}, args...)
	}
	return args
}
