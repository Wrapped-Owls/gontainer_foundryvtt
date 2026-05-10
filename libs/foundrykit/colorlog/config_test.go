package colorlog

import (
	"log/slog"
	"testing"
)

func TestConfigDefault(t *testing.T) {
	cfg := Default()
	if cfg.Level != LevelInfo {
		t.Errorf("Default Level = %v, want %v", cfg.Level, LevelInfo)
	}
	if cfg.Color != nil {
		t.Errorf("Default Color should be nil (auto-detect), got %v", cfg.Color)
	}
}

func TestLoadFromEnvLogLevel(t *testing.T) {
	cases := []struct {
		name  string
		level string
		want  Level
	}{
		{"debug", "debug", LevelDebug},
		{"info", "info", LevelInfo},
		{"warn", "warn", LevelWarn},
		{"error", "error", LevelError},
		{"ERROR-upper", "ERROR", LevelError},
		{"unknown-falls-back", "bogus", LevelInfo},
		{"empty-uses-default", "", LevelInfo},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("CONTAINER_LOG_LEVEL", tc.level)
			t.Setenv("CONTAINER_VERBOSE", "")
			c := Default()
			if err := LoadFromEnv(&c); err != nil {
				t.Fatal(err)
			}
			if c.Level != tc.want {
				t.Errorf("Level = %v, want %v", c.Level, tc.want)
			}
		})
	}
}

func TestLoadFromEnvVerbose(t *testing.T) {
	t.Setenv("CONTAINER_VERBOSE", "1")
	t.Setenv("CONTAINER_LOG_LEVEL", "")
	c := Default()
	if err := LoadFromEnv(&c); err != nil {
		t.Fatal(err)
	}
	if c.Level != LevelDebug {
		t.Errorf("VERBOSE=1 should set debug level, got %v", c.Level)
	}
}

func TestLoadFromEnvLogLevelOverridesVerbose(t *testing.T) {
	// LOG_LEVEL takes precedence over VERBOSE.
	t.Setenv("CONTAINER_VERBOSE", "1")
	t.Setenv("CONTAINER_LOG_LEVEL", "warn")
	c := Default()
	if err := LoadFromEnv(&c); err != nil {
		t.Fatal(err)
	}
	if c.Level != LevelWarn {
		t.Errorf("explicit LOG_LEVEL should win, got %v", c.Level)
	}
}

func TestParseLevelValues(t *testing.T) {
	cases := []struct {
		in   string
		want Level
	}{
		{"debug", slog.LevelDebug},
		{"DEBUG", slog.LevelDebug},
		{"info", slog.LevelInfo},
		{"warn", slog.LevelWarn},
		{"warning", slog.LevelWarn},
		{"error", slog.LevelError},
		{"err", slog.LevelError},
		{"unknown", slog.LevelInfo},
	}
	for _, tc := range cases {
		got, err := parseLevel(tc.in)
		if err != nil {
			t.Errorf("parseLevel(%q) unexpected error: %v", tc.in, err)
		}
		if got != tc.want {
			t.Errorf("parseLevel(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}
}
