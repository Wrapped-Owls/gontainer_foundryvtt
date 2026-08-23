package procspawn

import (
	"os"
	"slices"
	"syscall"
	"testing"
)

func TestConfigDefault(t *testing.T) {
	cfg := Default()
	if len(cfg.Passlist) == 0 {
		t.Error("Default Passlist should not be empty")
	}
	if len(cfg.ForwardSignals) == 0 {
		t.Error("Default ForwardSignals should not be empty")
	}
	if !slices.Contains(cfg.ForwardSignals, os.Signal(syscall.SIGTERM)) {
		t.Error("Default ForwardSignals must include SIGTERM")
	}
}
