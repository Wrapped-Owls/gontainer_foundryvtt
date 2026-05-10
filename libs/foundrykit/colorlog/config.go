package colorlog

import (
	"strings"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrykit/confloader"
)

type Config struct {
	Level Level

	Name  string
	Color *bool
}

func Default() Config {
	return Config{Level: LevelInfo}
}

func LoadFromEnv(c *Config) error {
	return confloader.BindEnv(
		confloader.BindField(&c.Level, envVerbose, func(v string) (Level, error) {
			if strings.TrimSpace(v) != "" {
				return LevelDebug, nil
			}
			return c.Level, nil
		}),
		confloader.BindField(&c.Level, envLogLevel, parseLevel), // must win, run after VERBOSE
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
