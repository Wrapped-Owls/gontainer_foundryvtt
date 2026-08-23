package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/cmd"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrykit/cliargs"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrykit/colorlog"
)

func main() {
	logger := colorlog.New("foundrymanager", colorlog.LevelFromEnv())
	slog.SetDefault(logger)

	sub, args := cliargs.SplitSubcommand(os.Args[1:], "run")

	const exitUsage = 2
	switch sub {
	case "run":
		os.Exit(cmd.Run(args, logger))
	default:
		fmt.Fprintf(os.Stderr, "foundrymanager: unknown subcommand %q\n", sub)
		os.Exit(exitUsage)
	}
}
