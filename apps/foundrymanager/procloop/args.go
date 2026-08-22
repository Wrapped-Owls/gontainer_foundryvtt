package procloop

import (
	"path/filepath"
	"strconv"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundryruntime/jsruntime"
)

const bunRun = "run"

func BuildArgs(s State) []string {
	args := []string{
		filepath.Join(s.InstallRoot, s.MainScript),
		"--dataPath=" + s.DataPath,
		"--port=" + strconv.Itoa(s.Port),
	}
	if s.World != "" {
		args = append(args, "--world="+s.World)
	}
	if s.JSRuntime.Kind == jsruntime.Bun {
		return append([]string{bunRun}, args...)
	}
	return args
}
