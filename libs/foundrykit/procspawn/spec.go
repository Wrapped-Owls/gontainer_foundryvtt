package procspawn

import (
	"io"
	"os"
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
