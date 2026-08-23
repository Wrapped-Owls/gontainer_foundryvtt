package confloader

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

type testConfig struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

func defaultTestConfig() testConfig {
	return testConfig{Host: "localhost", Port: 8080}
}

func TestLoadMissingFileUsesDefaults(t *testing.T) {
	cfg, err := Load(
		filepath.Join(t.TempDir(), "nonexistent.json"),
		defaultTestConfig(),
		func(c *testConfig) error { return nil },
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Host != "localhost" || cfg.Port != 8080 {
		t.Fatalf("expected defaults, got %+v", cfg)
	}
}

func TestLoadAppliesEnvViaBindField(t *testing.T) {
	t.Setenv("TEST_HOST", "remotehost")

	cfg, err := Load(
		filepath.Join(t.TempDir(), "nonexistent.json"),
		defaultTestConfig(),
		func(c *testConfig) error {
			return BindEnv(
				BindField(&c.Host, "TEST_HOST", nil),
			)
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Host != "remotehost" {
		t.Fatalf("expected Host=remotehost, got %q", cfg.Host)
	}
	if cfg.Port != 8080 {
		t.Fatalf("expected Port=8080, got %d", cfg.Port)
	}
}

func TestLoadReadsJSONFile(t *testing.T) {
	dir := t.TempDir()
	cfgFile := filepath.Join(dir, "conf.json")
	if err := os.WriteFile(cfgFile, []byte(`{"host":"filehost","port":9090}`), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(
		cfgFile,
		defaultTestConfig(),
		func(c *testConfig) error { return nil },
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Host != "filehost" || cfg.Port != 9090 {
		t.Fatalf("expected filehost:9090, got %+v", cfg)
	}
}

func TestBindEnvStopsOnFirstError(t *testing.T) {
	callCount := 0
	sentinel := errors.New("first binder error")

	err := BindEnv(
		func() error { callCount++; return sentinel },
		func() error { callCount++; return fmt.Errorf("second binder error") },
	)

	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
	if callCount != 1 {
		t.Fatalf("expected exactly 1 binder called, got %d", callCount)
	}
}

func TestBindRequiredMissingVar(t *testing.T) {
	if err := os.Unsetenv("REQUIRED_TEST_VAR"); err != nil {
		t.Fatal(err)
	}
	var s string
	err := BindEnv(BindRequired(&s, "REQUIRED_TEST_VAR", nil))
	if err == nil {
		t.Fatal("expected error for missing required var")
	}
}

func TestBindRequiredPresentVar(t *testing.T) {
	t.Setenv("REQUIRED_TEST_VAR", "hello")
	var s string
	err := BindEnv(BindRequired(&s, "REQUIRED_TEST_VAR", nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s != "hello" {
		t.Fatalf("expected s=hello, got %q", s)
	}
}
