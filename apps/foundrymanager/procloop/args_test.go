package procloop

import (
	"slices"
	"testing"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundryruntime/jsruntime"
)

const testMainScript = "/foundry/main.mjs"

func TestBuildArgs(t *testing.T) {
	testCases := []struct {
		name       string
		kind       jsruntime.Kind
		mainScript string
		dataPath   string
		port       int
		world      string
		want       []string
	}{
		{
			name:       "node",
			kind:       jsruntime.Node,
			mainScript: testMainScript,
			dataPath:   "/data",
			port:       30000,
			want:       []string{testMainScript, "--dataPath=/data", "--port=30000"},
		},
		{
			name:       "bun",
			kind:       jsruntime.Bun,
			mainScript: testMainScript,
			dataPath:   "/data",
			port:       30000,
			want:       []string{"run", testMainScript, "--dataPath=/data", "--port=30000"},
		},
		{
			name:       "node with world",
			kind:       jsruntime.Node,
			mainScript: testMainScript,
			dataPath:   "/data",
			port:       30000,
			world:      "my-world",
			want: []string{
				testMainScript, "--dataPath=/data", "--port=30000", "--world=my-world",
			},
		},
		{
			name:       "bun with world",
			kind:       jsruntime.Bun,
			mainScript: testMainScript,
			dataPath:   "/data",
			port:       30000,
			world:      "my-world",
			want: []string{
				"run", testMainScript, "--dataPath=/data", "--port=30000", "--world=my-world",
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			got := BuildArgs(
				testCase.kind,
				testCase.mainScript,
				testCase.dataPath,
				testCase.port,
				testCase.world,
			)
			if !slices.Equal(got, testCase.want) {
				t.Errorf("BuildArgs = %v, want %v", got, testCase.want)
			}
		})
	}
}
