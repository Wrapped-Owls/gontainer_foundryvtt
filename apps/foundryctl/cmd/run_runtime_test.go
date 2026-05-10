package cmd

import (
	"reflect"
	"testing"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundryruntime/jsruntime"
)

func TestRuntimeArgsForBun(t *testing.T) {
	got := runtimeArgs(jsruntime.Bun, "/foundry/resources/app/main.mjs", "/data", 30000)
	want := []string{"run", "/foundry/resources/app/main.mjs", "--dataPath=/data", "--port=30000"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("args = %#v, want %#v", got, want)
	}
}

func TestRuntimeArgsForNode(t *testing.T) {
	got := runtimeArgs(jsruntime.Node, "/foundry/resources/app/main.mjs", "/data", 30000)
	want := []string{"/foundry/resources/app/main.mjs", "--dataPath=/data", "--port=30000"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("args = %#v, want %#v", got, want)
	}
}
