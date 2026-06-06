package procloop

import (
	"testing"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundryruntime/jsruntime"
)

const testMainScript = "/foundry/main.mjs"

func TestBuildArgs(t *testing.T) {
	tests := []struct {
		name       string
		kind       jsruntime.Kind
		mainScript string
		dataPath   string
		port       int
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildArgs(tt.kind, tt.mainScript, tt.dataPath, tt.port)
			if len(got) != len(tt.want) {
				t.Fatalf("len mismatch: got %v, want %v", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("arg[%d]: got %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}
