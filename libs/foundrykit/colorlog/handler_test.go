package colorlog

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
	"testing/synctest"
	"time"
)

func TestHandlerEnabled(t *testing.T) {
	h := &handler{level: LevelWarn}
	if h.Enabled(context.TODO(), LevelDebug) {
		t.Error("debug should be disabled at warn level")
	}
	if h.Enabled(context.TODO(), LevelInfo) {
		t.Error("info should be disabled at warn level")
	}
	if !h.Enabled(context.TODO(), LevelWarn) {
		t.Error("warn should be enabled")
	}
	if !h.Enabled(context.TODO(), LevelError) {
		t.Error("error should be enabled")
	}
}

func TestHandlerFormatLine(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var buf bytes.Buffer
		noColor := false
		h := &handler{
			name:  "Test",
			level: LevelInfo,
			out:   &buf,
			color: noColor,
		}
		now := time.Now()
		rec := slog.NewRecord(now, slog.LevelInfo, "hello world", 0)
		if err := h.Handle(context.TODO(), rec); err != nil {
			t.Fatal(err)
		}
		got := buf.String()
		want := "Test | " + now.Format(time.DateTime) + " | [info] hello world\n"
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
}

func TestHandlerAttrs(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var buf bytes.Buffer
		noColor := false
		h := &handler{
			name:  "X",
			level: LevelInfo,
			out:   &buf,
			color: noColor,
		}
		now := time.Now()
		rec := slog.NewRecord(now, slog.LevelInfo, "msg", 0)
		rec.AddAttrs(slog.String("k", "v"))
		if err := h.Handle(context.TODO(), rec); err != nil {
			t.Fatal(err)
		}
		got := buf.String()
		if !strings.Contains(got, "k=v") {
			t.Errorf("expected k=v in output: %q", got)
		}
	})
}

func TestHandlerWithAttrs(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var buf bytes.Buffer
		noColor := false
		h := &handler{
			name:  "X",
			level: LevelInfo,
			out:   &buf,
			color: noColor,
		}
		prefixed := h.WithAttrs([]slog.Attr{slog.String("pre", "attached")})
		now := time.Now()
		rec := slog.NewRecord(now, slog.LevelInfo, "msg", 0)
		if err := prefixed.Handle(context.TODO(), rec); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(buf.String(), "pre=attached") {
			t.Errorf("expected pre=attached in output: %q", buf.String())
		}
	})
}

func TestColorize(t *testing.T) {
	cases := []struct {
		lvl  Level
		code string
	}{
		{LevelDebug, ansiBlue},
		{LevelInfo, ansiGreen},
		{LevelWarn, ansiYellow},
		{LevelError, ansiRed},
	}
	for _, tc := range cases {
		got := colorize(tc.lvl)
		if got != tc.code {
			t.Errorf("colorize(%v) = %q, want %q", tc.lvl, got, tc.code)
		}
	}
}
