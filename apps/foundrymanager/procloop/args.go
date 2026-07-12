package procloop

import (
	"strconv"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundryruntime/jsruntime"
)

func BuildArgs(kind jsruntime.Kind, mainScript, dataPath string, port int, world string) []string {
	args := []string{
		mainScript,
		"--dataPath=" + dataPath,
		"--port=" + strconv.Itoa(port),
	}
	if world != "" {
		args = append(args, "--world="+world)
	}
	if kind == jsruntime.Bun {
		return append([]string{"run"}, args...)
	}
	return args
}
