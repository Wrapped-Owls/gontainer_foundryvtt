package procspawn

import (
	"io"
	"os"
	"syscall"
)

// Spec describes how to execute a child process.
type Spec struct {
	// Path is the absolute path to the binary to exec.
	Path string
	// Args are passed verbatim (Args[0] becomes argv[0]).
	Args []string
	// Env is the full child environment. When nil, FilterEnv(os.Environ(),
	// DefaultPasslist) is used.
	Env []string
	// Dir is the child working directory. Empty means inherit.
	Dir string
	// ForwardSignals, when non-nil, is the set of signals captured in the
	// parent and forwarded to the child's process group. Defaults to
	// SIGTERM + SIGINT when nil.
	ForwardSignals []os.Signal
	// Stdin defaults to os.Stdin. Stdout/Stderr default to os.Stdout/os.Stderr
	// and may be any writer (e.g. an io.MultiWriter that also captures logs).
	Stdin  *os.File
	Stdout io.Writer
	Stderr io.Writer
}

// withDefaults returns a copy of s with every unset field replaced by its
// documented default. It touches no process and can be exercised without
// spawning a child.
func (s Spec) withDefaults() Spec {
	if s.Env == nil {
		s.Env = FilterEnv(os.Environ(), DefaultPasslist)
	}
	if s.ForwardSignals == nil {
		s.ForwardSignals = []os.Signal{syscall.SIGTERM, syscall.SIGINT}
	}
	if s.Stdin == nil {
		s.Stdin = os.Stdin
	}
	if s.Stdout == nil {
		s.Stdout = os.Stdout
	}
	if s.Stderr == nil {
		s.Stderr = os.Stderr
	}
	return s
}
