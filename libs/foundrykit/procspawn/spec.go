package procspawn

import "os"

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
	// Stdin/Stdout/Stderr default to os.Std*.
	Stdin  *os.File
	Stdout *os.File
	Stderr *os.File
}
