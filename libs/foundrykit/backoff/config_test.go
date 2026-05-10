package backoff

import (
	"os"
	"testing"
)

func TestNewFromEnvDefaults(t *testing.T) {
	t.Setenv("CONTAINER_CACHE", "")
	t.Setenv("KUBERNETES_SERVICE_HOST", "")
	m := NewFromEnv()
	if m.CacheDir != "" {
		t.Errorf("explicit empty CONTAINER_CACHE should disable cache, got %q", m.CacheDir)
	}
	if err := os.Unsetenv("CONTAINER_CACHE"); err != nil {
		t.Fatal(err)
	}
	m = NewFromEnv()
	if m.CacheDir != "/data/container_cache" {
		t.Errorf("default cache = %q, want /data/container_cache", m.CacheDir)
	}
	t.Setenv("KUBERNETES_SERVICE_HOST", "10.0.0.1")
	m = NewFromEnv()
	if !m.KubernetesBypass {
		t.Error("KUBERNETES_SERVICE_HOST should enable bypass")
	}
}

func TestDefault(t *testing.T) {
	cfg := Default()
	if cfg.CacheDir != defaultCacheDir {
		t.Errorf("Default CacheDir = %q, want %q", cfg.CacheDir, defaultCacheDir)
	}
	if cfg.KubernetesBypass {
		t.Error("Default KubernetesBypass should be false")
	}
}
