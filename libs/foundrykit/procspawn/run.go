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

// Run launches spec's child, forwards signals to its process group, and blocks
// until the child exits or ctx is done. A cancelled ctx signals the child and
// still waits for it, so the reported code is always the child's own. A child
// that fails to start reports -1 with a non-nil error.
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
		if code, ok := exitCodeFromWaitStatus(ws); ok {
			return code, nil
		}
		return -1, exitErr
	}
	return -1, err
}

// signalExitBase mirrors the shell convention for a signal death.
const signalExitBase = 128

// exitCodeFromWaitStatus reports the child's exit code, or ok false when the
// status is neither an exit nor a signal death.
func exitCodeFromWaitStatus(ws syscall.WaitStatus) (code int, ok bool) {
	switch {
	case ws.Signaled():
		return signalExitBase + int(ws.Signal()), true
	case ws.Exited():
		return ws.ExitStatus(), true
	default:
		return -1, false
	}
}
