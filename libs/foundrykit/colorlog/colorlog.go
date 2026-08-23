package colorlog

import (
	"io"
	"log/slog"
	"os"

	"golang.org/x/term"
)

// Options configures a logger. The zero value is valid and produces a
// stderr text logger at info level with auto-detected colour.
type Options struct {
	// Name is the LOG_NAME prefix (e.g. "Entrypoint", "Launcher").
	Name string
	// Level controls minimum emitted level. Defaults to LevelInfo.
	Level Level
	// Out is the destination writer. Defaults to os.Stderr.
	Out io.Writer
	// Color forces colour on/off. Nil = auto-detect (TTY on Out).
	Color *bool
}

// New constructs a *slog.Logger with the colorlog format.
func New(name string, level Level) *slog.Logger {
	return NewWithOptions(Options{Name: name, Level: level})
}

// NewWithOptions is the configurable constructor.
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

// ANSI colour codes used by the handler.
const (
	ansiReset  = "\x1b[0m"
	ansiBlue   = "\x1b[34m" // debug
	ansiGreen  = "\x1b[32m" // info
	ansiYellow = "\x1b[33m" // warn
	ansiRed    = "\x1b[31m" // error
)
