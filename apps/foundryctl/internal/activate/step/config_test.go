package step

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestAppConfigStepFieldMapping(t *testing.T) {
	testCases := []struct {
		name string
		port string
		want int
	}{
		{name: "port overlaid from app runtime config", port: "31000", want: 31000},
		{name: "default port kept when unset", port: "", want: 30000},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Setenv("CONF_FILE", filepath.Join(t.TempDir(), "absent.json"))
			t.Setenv("FOUNDRY_DATA_PATH", "/mnt/data")
			if testCase.port != "" {
				t.Setenv("FOUNDRY_PORT", testCase.port)
			}

			s := &State{}
			err := appConfigStep{}.Apply(context.Background(), s, discardLogger())
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if s.App.Paths.DataPath != "/mnt/data" {
				t.Fatalf("App.Paths.DataPath = %q, want %q", s.App.Paths.DataPath, "/mnt/data")
			}
			if s.Runtime.DataPath != "/mnt/data" {
				t.Fatalf("Runtime.DataPath = %q, want the App.Paths.DataPath overlay", s.Runtime.DataPath)
			}
			if s.App.Runtime.Port != testCase.want {
				t.Fatalf("App.Runtime.Port = %d, want %d", s.App.Runtime.Port, testCase.want)
			}
			if s.Runtime.Port != testCase.want {
				t.Fatalf("Runtime.Port = %d, want the App.Runtime.Port overlay %d", s.Runtime.Port, testCase.want)
			}
		})
	}
}

func TestAppConfigStepFailsOnMalformedConfigFile(t *testing.T) {
	t.Setenv("CONF_FILE", writeBadJSON(t))

	err := appConfigStep{}.Apply(context.Background(), &State{}, discardLogger())
	if err == nil {
		t.Fatal("expected an error for a malformed config file")
	}
}

func TestAppConfigStepFailsOnInvalidRuntimeEnv(t *testing.T) {
	t.Setenv("CONF_FILE", filepath.Join(t.TempDir(), "absent.json"))
	t.Setenv("FOUNDRY_DEMO_CONFIG", "{not json")

	err := appConfigStep{}.Apply(context.Background(), &State{}, discardLogger())
	if err == nil {
		t.Fatal("expected an error for an invalid FOUNDRY_DEMO_CONFIG value")
	}
}

func writeBadJSON(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "bad.json")
	if err := os.WriteFile(path, []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}
