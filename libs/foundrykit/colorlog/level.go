package colorlog

import "log/slog"

const (
	envLogLevel = "CONTAINER_LOG_LEVEL"
	envVerbose  = "CONTAINER_VERBOSE"
)

// Level is the log level type used by colorlog (debug, info, warn, error).
type Level = slog.Level

const (
	LevelDebug = slog.LevelDebug
	LevelInfo  = slog.LevelInfo
	LevelWarn  = slog.LevelWarn
	LevelError = slog.LevelError
)

// LevelFromEnv reads the log level from environment variables.
// Prefer LoadFromEnv for new code.
func LevelFromEnv() Level {
	c := Default()
	_ = LoadFromEnv(&c)
	return c.Level
}
