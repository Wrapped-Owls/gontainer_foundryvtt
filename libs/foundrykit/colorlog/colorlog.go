package colorlog

import (
	"io"
	"log/slog"
	"os"

	"golang.org/x/term"
)

type Options struct {
	Name string

	Level Level

	Out   io.Writer
	Color *bool
}

func New(name string, level Level) *slog.Logger {
	return NewWithOptions(Options{Name: name, Level: level})
}

func NewWithOptions(opts Options) *slog.Logger {
	if opts.Out == nil {
		opts.Out = os.Stderr
	}
	color := autoColor(opts.Out)
	if opts.Color != nil {
		color = *opts.Color
	}
	return slog.New(&handler{
		name:  opts.Name,
		level: opts.Level,
		out:   opts.Out,
		color: color,
	})
}

func autoColor(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	return term.IsTerminal(int(f.Fd()))
}

const (
	ansiBlue   = "\x1b[34m"
	ansiGreen  = "\x1b[32m"
	ansiYellow = "\x1b[33m"
	ansiRed    = "\x1b[31m"
)
