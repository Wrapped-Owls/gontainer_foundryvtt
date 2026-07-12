package procloop

import (
	"strconv"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundryruntime/jsruntime"
)

// BuildArgs builds the JS runtime argv for the given kind, script, data path,
// port, and optional world. When world is non-empty a --world flag is appended
// so Foundry boots straight into that world instead of the setup screen.
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
