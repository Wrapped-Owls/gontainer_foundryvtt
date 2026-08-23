package procspawn

import (
	"io"
	"os"
	"syscall"
)

type Spec struct {
	Path           string
	Args           []string
	Env            []string
	Dir            string
	ForwardSignals []os.Signal

	Stdin  *os.File
	Stdout io.Writer
	Stderr io.Writer
}

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
