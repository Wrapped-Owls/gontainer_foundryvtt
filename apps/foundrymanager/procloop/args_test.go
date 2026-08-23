package procloop

import (
	"slices"
	"testing"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundryruntime/jsruntime"
)

const (
	testMainScript = "/foundry/resources/app/main.mjs"
	testDataArg    = "--dataPath=/data"
	testPortArg    = "--port=30000"
)

func sessionOn(kind jsruntime.Kind, world string) State {
	return State{
		InstallRoot: "/foundry",
		MainScript:  "resources/app/main.mjs",
		JSRuntime:   jsruntime.Runtime{Kind: kind},
		DataPath:    "/data",
		Port:        30000,
		World:       world,
	}
}

func TestBuildArgs(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		session State
		want    []string
	}{
		{
			name:    "node takes the script directly",
			session: sessionOn(jsruntime.Node, ""),
			want:    []string{testMainScript, testDataArg, testPortArg},
		},
		{
			name:    "bun needs the run prefix node rejects",
			session: sessionOn(jsruntime.Bun, ""),
			want:    []string{bunRun, testMainScript, testDataArg, testPortArg},
		},
		{
			name:    "a named world boots straight into it",
			session: sessionOn(jsruntime.Node, "my-world"),
			want: []string{
				testMainScript, testDataArg, testPortArg, "--world=my-world",
			},
		},
		{
			name:    "bun with a world keeps both",
			session: sessionOn(jsruntime.Bun, "my-world"),
			want: []string{
				bunRun, testMainScript, testDataArg, testPortArg, "--world=my-world",
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			if got := BuildArgs(testCase.session); !slices.Equal(got, testCase.want) {
				t.Fatalf("BuildArgs = %v, want %v", got, testCase.want)
			}
		})
	}
}
