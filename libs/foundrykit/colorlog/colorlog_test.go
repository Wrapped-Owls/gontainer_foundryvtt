package colorlog

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func fixedTime() time.Time {
	return time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)
}

func TestFormatOutput(t *testing.T) {
	var buf bytes.Buffer
	noColor := false
	log := NewWithOptions(Options{
		Name:  "Entrypoint",
		Level: LevelDebug,
		Out:   &buf,
		Color: &noColor,
		Now:   fixedTime,
	})

	log.Info("hello", "key", "val")
	got := buf.String()
	want := "Entrypoint | 2025-01-02 03:04:05 | [info] hello key=val\n"
	if got != want {
		t.Fatalf("format mismatch:\n got: %q\nwant: %q", got, want)
	}
}

func TestLevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	noColor := false
	log := NewWithOptions(Options{
		Name: "X", Level: LevelWarn, Out: &buf, Color: &noColor, Now: fixedTime,
	})
	log.Debug("d")
	log.Info("i")
	log.Warn("w")
	log.Error("e")
	out := buf.String()
	for _, want := range []string{"[warn] w", "[error] e"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
	for _, bad := range []string{"[debug]", "[info]"} {
		if strings.Contains(out, bad) {
			t.Errorf("unexpected %q in:\n%s", bad, out)
		}
	}
}

func TestColorEnabled(t *testing.T) {
	var buf bytes.Buffer
	yes := true
	log := NewWithOptions(Options{
		Name: "X", Level: LevelInfo, Out: &buf, Color: &yes, Now: fixedTime,
	})
	log.Info("hi")
	if !strings.Contains(buf.String(), "\x1b[32m") {
		t.Fatalf("expected ANSI green code in output: %q", buf.String())
	}
}
