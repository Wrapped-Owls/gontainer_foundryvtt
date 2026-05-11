package backoff

import (
	"testing"
)

func TestDefault(t *testing.T) {
	cfg := Default()
	if cfg.CacheDir != defaultCacheDir {
		t.Errorf("Default CacheDir = %q, want %q", cfg.CacheDir, defaultCacheDir)
	}
	if cfg.KubernetesBypass {
		t.Error("Default KubernetesBypass should be false")
	}
}
