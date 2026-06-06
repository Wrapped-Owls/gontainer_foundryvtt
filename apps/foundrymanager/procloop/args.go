package procloop

import (
	"strconv"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundryruntime/jsruntime"
)

func BuildArgs(kind jsruntime.Kind, mainScript, dataPath string, port int) []string {
	args := []string{
		mainScript,
		"--dataPath=" + dataPath,
		"--port=" + strconv.Itoa(port),
	}
	if kind == jsruntime.Bun {
		return append([]string{"run"}, args...)
	}
	return args
}
