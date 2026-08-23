package procspawn

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

func Run(ctx context.Context, spec Spec) (int, error) {
	if spec.Path == "" {
		return -1, errors.New("procspawn: Spec.Path is required")
	}
	spec = spec.withDefaults()

	cmd := exec.CommandContext(ctx, spec.Path, spec.Args...)
	cmd.Env = spec.Env
	cmd.Dir = spec.Dir
	cmd.Stdin = spec.Stdin
	cmd.Stdout = spec.Stdout
	cmd.Stderr = spec.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true} // Foundry forks workers, signal them all
	cmd.Cancel = func() error {
		// SIGTERM, not the default SIGKILL, lets the JS runtime shut down cleanly; ErrProcessDone
		// tells exec the child was killed on purpose, so it won't report a fake "signal: killed"
		_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM)
		return os.ErrProcessDone
	}

	if err := cmd.Start(); err != nil {
		return -1, fmt.Errorf("procspawn: start %s: %w", spec.Path, err)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, spec.ForwardSignals...)
	defer signal.Stop(sigCh)

	done := make(chan struct{})
	go func() {
		for {
			select {
			case <-done:
				return
			case s := <-sigCh:
				if cmd.Process == nil {
					return
				}
				ss, ok := s.(syscall.Signal)
				if !ok {
					ss = syscall.SIGTERM
				}
				_ = syscall.Kill(-cmd.Process.Pid, ss)
			}
		}
	}()

	err := cmd.Wait()
	close(done)

	if err == nil {
		return cmd.ProcessState.ExitCode(), nil
	}
	if exitErr, ok := errors.AsType[*exec.ExitError](err); ok {
		ws := exitErr.Sys().(syscall.WaitStatus)
		if code, ok := exitCodeFromWaitStatus(ws); ok {
			return code, nil
		}
		return -1, exitErr
	}
	return -1, err
}

func exitCodeFromWaitStatus(ws syscall.WaitStatus) (code int, ok bool) {
	const signalExitBase = 128 // mirrors the shell convention for a signal death

	switch {
	case ws.Signaled():
		return signalExitBase + int(ws.Signal()), true
	case ws.Exited():
		return ws.ExitStatus(), true
	default:
		return -1, false
	}
}
