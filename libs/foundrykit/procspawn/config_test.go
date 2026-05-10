package procspawn

import (
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
	// Must include SIGTERM.
	found := false
	for _, s := range cfg.ForwardSignals {
		if s == syscall.SIGTERM {
			found = true
		}
	}
	if !found {
		t.Error("Default ForwardSignals must include SIGTERM")
	}
}

func TestLoadFromEnvIsNoOp(t *testing.T) {
	cfg := Default()
	if err := LoadFromEnv(&cfg); err != nil {
		t.Fatalf("LoadFromEnv returned error: %v", err)
	}
}
