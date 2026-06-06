// Command foundrymanager is a standalone entrypoint for the Foundry profile
// dashboard. When imported by foundryctl, only the manager/ and profile/
// packages are used directly.
package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/cmd"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrykit/colorlog"
)

func main() {
	logger := colorlog.New("foundrymanager", colorlog.LevelFromEnv())
	slog.SetDefault(logger)

	args := os.Args[1:]
	sub := "run"
	if len(args) > 0 && args[0][0] != '-' {
		sub, args = args[0], args[1:]
	}

	const exitUsage = 2
	switch sub {
	case "run":
		os.Exit(cmd.Run(args, logger))
	default:
		fmt.Fprintf(os.Stderr, "foundrymanager: unknown subcommand %q\n", sub)
		os.Exit(exitUsage)
	}
}
