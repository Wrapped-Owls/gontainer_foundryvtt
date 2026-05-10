package colorlog

import (
	"strings"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrykit/confloader"
)

// Config configures a colorlog logger.
type Config struct {
	// Level is the minimum emitted log level. Default: LevelInfo.
	Level Level
	// Name is the LOG_NAME prefix (e.g. "foundryctl").
	Name string
	// Color forces colour output on/off. Nil = auto-detect (TTY).
	Color *bool
}

// Default returns a Config with sensible defaults.
func Default() Config {
	return Config{Level: LevelInfo}
}

// LoadFromEnv overlays environment variables onto c using confloader.
// CONTAINER_LOG_LEVEL (debug|info|warn|error) and CONTAINER_VERBOSE (any non-empty = debug).
// CONTAINER_LOG_LEVEL takes precedence over CONTAINER_VERBOSE when both are set.
func LoadFromEnv(c *Config) error {
	return confloader.BindEnv(
		// VERBOSE runs first (lower priority); LOG_LEVEL runs second and overwrites.
		confloader.BindField(&c.Level, envVerbose, func(v string) (Level, error) {
			if strings.TrimSpace(v) != "" {
				return LevelDebug, nil
			}
			return c.Level, nil
		}),
		confloader.BindField(&c.Level, envLogLevel, parseLevel),
	)
}

func parseLevel(v string) (Level, error) {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "debug":
		return LevelDebug, nil
	case "info":
		return LevelInfo, nil
	case "warn", "warning":
		return LevelWarn, nil
	case "error", "err":
		return LevelError, nil
	default:
		return LevelInfo, nil
	}
}
