package step

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	appconfig "github.com/wrapped-owls/gontainer_foundryvtt/apps/foundryctl/config"
)

func TestSecretsStepNeverAbortsTheSequence(t *testing.T) {
	testCases := []struct {
		name string
		body string
	}{
		{name: "missing secrets file", body: ""},
		{name: "malformed secrets file", body: "{not json"},
		{name: "valid secrets file with a known key", body: `{"foundry_admin_key":"topsecret"}`},
		{name: "valid secrets file with an unknown key", body: `{"unrelated_key":"value"}`},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Setenv("FOUNDRY_ADMIN_KEY", "")

			path := filepath.Join(t.TempDir(), "secrets.json")
			if testCase.body != "" {
				if err := os.WriteFile(path, []byte(testCase.body), 0o644); err != nil {
					t.Fatal(err)
				}
			}

			s := &State{App: appconfig.Config{}}
			s.App.Secrets.Path = path

			err := secretsStep{}.Apply(context.Background(), s, discardLogger())
			if err != nil {
				t.Fatalf("secretsStep must never return an error, got: %v", err)
			}
		})
	}
}

func TestSecretsStepAppliesAKnownKeyAsAnEnvVar(t *testing.T) {
	t.Setenv("FOUNDRY_ADMIN_KEY", "")

	path := filepath.Join(t.TempDir(), "secrets.json")
	if err := os.WriteFile(path, []byte(`{"foundry_admin_key":"topsecret"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	s := &State{App: appconfig.Config{}}
	s.App.Secrets.Path = path

	err := secretsStep{}.Apply(context.Background(), s, discardLogger())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := os.Getenv("FOUNDRY_ADMIN_KEY"); got != "topsecret" {
		t.Fatalf("FOUNDRY_ADMIN_KEY = %q, want %q", got, "topsecret")
	}
}
