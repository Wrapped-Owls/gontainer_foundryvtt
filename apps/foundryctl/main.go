package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundryctl/cmd"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrykit/cliargs"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrykit/colorlog"
)

func main() {
	logger := colorlog.New("foundryctl", colorlog.LevelFromEnv())
	slog.SetDefault(logger)

	sub, args := cliargs.SplitSubcommand(os.Args[1:], "run")

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
