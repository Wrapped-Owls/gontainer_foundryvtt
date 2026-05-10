package colorlog

import "log/slog"

const (
	envLogLevel = "CONTAINER_LOG_LEVEL"
	envVerbose  = "CONTAINER_VERBOSE"
)

type Level = slog.Level

const (
	LevelDebug = slog.LevelDebug
	LevelInfo  = slog.LevelInfo
	LevelWarn  = slog.LevelWarn
	LevelError = slog.LevelError
)

func LevelFromEnv() Level {
	c := Default()
	_ = LoadFromEnv(&c)
	return c.Level
}
