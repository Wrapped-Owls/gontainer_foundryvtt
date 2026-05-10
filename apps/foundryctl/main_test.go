package main

import "testing"

// Smoke test for tiny helpers. The bulk of foundryctl logic lives in
// libs/foundryruntime/lifecycle and is covered there.
// envOr and orDefault are tested in internal/cmd.

func TestStartsWithFlag(t *testing.T) {
	if !startsWithFlag("-x") || startsWithFlag("run") || startsWithFlag("") {
		t.Fatal("startsWithFlag misbehaves")
	}
}
