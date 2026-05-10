package confloader_test

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrykit/confloader"
)

type testConfig struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

func defaultTestConfig() testConfig {
	return testConfig{Host: "localhost", Port: 8080}
}

// TestLoadMissingFileUsesDefaults verifies that a non-existent config file
// does not cause an error and that defaults are preserved.
func TestLoadMissingFileUsesDefaults(t *testing.T) {
	cfg, err := confloader.Load(
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

// TestLoadAppliesEnvViaBindField verifies that Load calls the updater which
// applies env-var overrides via BindField.
func TestLoadAppliesEnvViaBindField(t *testing.T) {
	t.Setenv("TEST_HOST", "remotehost")

	cfg, err := confloader.Load(
		filepath.Join(t.TempDir(), "nonexistent.json"),
		defaultTestConfig(),
		func(c *testConfig) error {
			return confloader.BindEnv(
				confloader.BindField(&c.Host, "TEST_HOST", nil),
			)
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Host != "remotehost" {
		t.Fatalf("expected Host=remotehost, got %q", cfg.Host)
	}
	// Port should remain at default
	if cfg.Port != 8080 {
		t.Fatalf("expected Port=8080, got %d", cfg.Port)
	}
}

// TestLoadReadsJSONFile verifies that an existing JSON config file is parsed.
func TestLoadReadsJSONFile(t *testing.T) {
	dir := t.TempDir()
	cfgFile := filepath.Join(dir, "conf.json")
	if err := os.WriteFile(cfgFile, []byte(`{"host":"filehost","port":9090}`), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := confloader.Load(
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

// TestBindEnvStopsOnFirstError verifies that BindEnv returns immediately on
// the first binder that returns an error.
func TestBindEnvStopsOnFirstError(t *testing.T) {
	callCount := 0
	sentinel := errors.New("first binder error")

	err := confloader.BindEnv(
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

// TestBindRequiredMissingVar verifies that BindRequired errors on a missing var.
func TestBindRequiredMissingVar(t *testing.T) {
	if err := os.Unsetenv("REQUIRED_TEST_VAR"); err != nil {
		t.Fatal(err)
	}
	var s string
	err := confloader.BindEnv(confloader.BindRequired(&s, "REQUIRED_TEST_VAR", nil))
	if err == nil {
		t.Fatal("expected error for missing required var")
	}
}

// TestBindRequiredPresentVar verifies that BindRequired sets the field.
func TestBindRequiredPresentVar(t *testing.T) {
	t.Setenv("REQUIRED_TEST_VAR", "hello")
	var s string
	err := confloader.BindEnv(confloader.BindRequired(&s, "REQUIRED_TEST_VAR", nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s != "hello" {
		t.Fatalf("expected s=hello, got %q", s)
	}
}
