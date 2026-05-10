package cmd

import (
	"fmt"
	"runtime/debug"

	appconfig "github.com/wrapped-owls/gontainer_foundryvtt/apps/foundryctl/config"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundryruntime/lifecycle"
)

func Version() {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		fmt.Println("foundryctl: build info unavailable")
		return
	}
	fmt.Printf("foundryctl %s (%s)\n", bi.Main.Version, bi.GoVersion)
	cfg, err := appconfig.Load()
	if err != nil {
		cfg = appconfig.Default()
	}
	if info, derr := lifecycle.DetectInstalled(cfg.Paths.InstallRoot); derr == nil && info.Present {
		fmt.Printf("foundry installed: %s\n", info.Version)
	}
}
