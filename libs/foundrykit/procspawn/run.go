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

// Run launches the child described by spec, forwards configured signals
// from the parent to the child's process group, blocks until either the
// child exits or ctx is done, and returns the child exit code.
//
// Exit code semantics:
//   - normal exit              → child exit code (0..255)
//   - killed by signal N       → 128 + N (mirrors bash convention)
//   - failed to start          → -1, err non-nil
//   - ctx cancelled before exit → child is signalled (SIGTERM) and we still
//     wait for it; the returned exit code is whatever the child reports
func Run(ctx context.Context, spec Spec) (int, error) {
	if spec.Path == "" {
		return -1, errors.New("procspawn: Spec.Path is required")
	}
	if spec.Env == nil {
		spec.Env = FilterEnv(os.Environ(), DefaultPasslist)
	}
	if spec.ForwardSignals == nil {
		spec.ForwardSignals = []os.Signal{syscall.SIGTERM, syscall.SIGINT}
	}
	if spec.Stdin == nil {
		spec.Stdin = os.Stdin
	}
	if spec.Stdout == nil {
		spec.Stdout = os.Stdout
	}
	if spec.Stderr == nil {
		spec.Stderr = os.Stderr
	}

	cmd := exec.CommandContext(ctx, spec.Path, spec.Args...)
	cmd.Env = spec.Env
	cmd.Dir = spec.Dir
	cmd.Stdin = spec.Stdin
	cmd.Stdout = spec.Stdout
	cmd.Stderr = spec.Stderr
	// Put the child in its own process group so we can signal the whole
	// group (Foundry forks workers).
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	// Override CommandContext's default Cancel (which sends SIGKILL) with
	// a polite SIGTERM so the JS runtime can shut down cleanly. Returning
	// os.ErrProcessDone tells exec to surface the child's natural exit
	// status rather than synthesising "signal: killed".
	cmd.Cancel = func() error {
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
		switch {
		case ws.Signaled():
			return 128 + int(ws.Signal()), nil
		case ws.Exited():
			return ws.ExitStatus(), nil
		default:
			return -1, exitErr
		}
	}
	return -1, err
}
