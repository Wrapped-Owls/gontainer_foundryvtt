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
	t.Parallel()

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
		h := &handler{name: "Test", level: LevelInfo, out: &buf, color: false}
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
		h := &handler{name: "X", level: LevelInfo, out: &buf, color: false}
		rec := slog.NewRecord(time.Now(), slog.LevelInfo, "msg", 0)
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
		h := &handler{name: "X", level: LevelInfo, out: &buf, color: false}
		prefixed := h.WithAttrs([]slog.Attr{slog.String("pre", "attached")})
		rec := slog.NewRecord(time.Now(), slog.LevelInfo, "msg", 0)
		if err := prefixed.Handle(context.TODO(), rec); err != nil {
			t.Fatal(err)
		}
		got := buf.String()
		if !strings.Contains(got, "pre=attached") {
			t.Errorf("expected pre=attached in output: %q", got)
		}
	})
}

func TestColorize(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		lvl  Level
		code string
	}{
		{LevelDebug, ansiBlue},
		{LevelInfo, ansiGreen},
		{LevelWarn, ansiYellow},
		{LevelError, ansiRed},
	}
	for _, testCase := range testCases {
		got := colorize(testCase.lvl)
		if got != testCase.code {
			t.Errorf("colorize(%v) = %q, want %q", testCase.lvl, got, testCase.code)
		}
	}
}
