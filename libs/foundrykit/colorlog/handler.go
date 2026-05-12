package colorlog

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"
)

type handler struct {
	name  string
	level Level
	out   io.Writer
	now   func() time.Time
	color bool
	attrs []slog.Attr
	group string
}

func (h *handler) Enabled(_ context.Context, lvl Level) bool { return lvl >= h.level }

func (h *handler) Handle(_ context.Context, r slog.Record) error {
	ts := h.now().Format(time.DateTime)
	level := strings.ToLower(r.Level.String())
	colored := level
	if h.color {
		colored = colorize(r.Level) + level + ansiReset
	}
	var b strings.Builder
	b.Grow(64 + len(r.Message))
	b.WriteString(h.name)
	b.WriteString(" | ")
	b.WriteString(ts)
	b.WriteString(" | [")
	b.WriteString(colored)
	b.WriteString("] ")
	b.WriteString(r.Message)
	// append structured attrs as "key=value" pairs at the end of the line.
	appendAttrs(&b, h.attrs)
	r.Attrs(func(a slog.Attr) bool {
		appendAttr(&b, a)
		return true
	})
	b.WriteByte('\n')
	_, err := io.WriteString(h.out, b.String())
	return err
}

func (h *handler) WithAttrs(as []slog.Attr) slog.Handler {
	c := *h
	c.attrs = append(append([]slog.Attr{}, h.attrs...), as...)
	return &c
}

func (h *handler) WithGroup(name string) slog.Handler {
	c := *h
	if c.group == "" {
		c.group = name
	} else {
		c.group = c.group + "." + name
	}
	return &c
}

func appendAttrs(b *strings.Builder, as []slog.Attr) {
	for _, a := range as {
		appendAttr(b, a)
	}
}

func appendAttr(b *strings.Builder, a slog.Attr) {
	if a.Equal(slog.Attr{}) {
		return
	}
	b.WriteByte(' ')
	b.WriteString(a.Key)
	b.WriteByte('=')
	fmt.Fprintf(b, "%v", a.Value.Any())
}

func colorize(lvl Level) string {
	switch {
	case lvl <= LevelDebug:
		return ansiBlue
	case lvl < LevelWarn:
		return ansiGreen
	case lvl < LevelError:
		return ansiYellow
	default:
		return ansiRed
	}
}
