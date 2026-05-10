package colorlog

import (
	"bytes"
	"strings"
	"testing"
	"testing/synctest"
	"time"
)

func TestFormatOutput(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var buf bytes.Buffer
		noColor := false
		log := NewWithOptions(Options{
			Name:  "Entrypoint",
			Level: LevelDebug,
			Out:   &buf,
			Color: &noColor,
		})

		now := time.Now()
		log.Info("hello", "key", "val")
		got := buf.String()
		want := "Entrypoint | " + now.Format(time.DateTime) + " | [info] hello key=val\n"
		if got != want {
			t.Fatalf("format mismatch:\n got: %q\nwant: %q", got, want)
		}
	})
}

func TestLevelFiltering(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var buf bytes.Buffer
		noColor := false
		log := NewWithOptions(Options{
			Name: "X", Level: LevelWarn, Out: &buf, Color: &noColor,
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
	})
}

func TestColorEnabled(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var buf bytes.Buffer
		yes := true
		log := NewWithOptions(Options{
			Name: "X", Level: LevelInfo, Out: &buf, Color: &yes,
		})
		log.Info("hi")
		got := buf.String()
		if !strings.Contains(got, "\x1b[32m") {
			t.Fatalf("expected ANSI green code in output: %q", got)
		}
	})
}
