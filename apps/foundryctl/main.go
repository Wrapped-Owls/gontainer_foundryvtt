// Command foundryctl is the PID 1 entrypoint for the foundryvtt-docker
// container. Sub-commands: run (default), healthcheck, options, version.
package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundryctl/cmd"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrykit/colorlog"
)

func main() {
	logger := colorlog.New("foundryctl", colorlog.LevelFromEnv())
	slog.SetDefault(logger)

	args := os.Args[1:]
	sub := "run"
	if len(args) > 0 && !startsWithFlag(args[0]) {
		sub, args = args[0], args[1:]
	}

	switch sub {
	case "run":
		os.Exit(cmd.Run(args, logger))
	case "healthcheck":
		os.Exit(cmd.Healthcheck(args, logger))
	case "options":
		os.Exit(cmd.Options(args, logger))
	case "version":
		cmd.Version()
	default:
		_, _ = fmt.Fprintf(os.Stderr, "foundryctl: unknown subcommand %q\n", sub)
		os.Exit(2)
	}
}

func startsWithFlag(s string) bool { return len(s) > 0 && s[0] == '-' }
