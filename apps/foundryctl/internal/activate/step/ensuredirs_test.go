package step

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	appconfig "github.com/wrapped-owls/gontainer_foundryvtt/apps/foundryctl/config"
)

func TestEnsureDirsStepCreatesEveryConfiguredDir(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	testCases := []struct {
		name string
		want appconfig.PathsConfig
	}{
		{
			name: "all three dirs configured",
			want: appconfig.PathsConfig{
				DataPath:    filepath.Join(root, "data"),
				InstallRoot: filepath.Join(root, "foundry"),
				SourcesDir:  filepath.Join(root, "foundry", "sources"),
			},
		},
		{
			name: "empty entries are skipped, not created as the cwd",
			want: appconfig.PathsConfig{
				DataPath:    filepath.Join(root, "only-data"),
				InstallRoot: "",
				SourcesDir:  "",
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			s := &State{App: appconfig.Config{Paths: testCase.want}}
			err := ensureDirsStep{}.Apply(context.Background(), s, discardLogger())
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			for _, dir := range []string{
				testCase.want.DataPath, testCase.want.InstallRoot, testCase.want.SourcesDir,
			} {
				if dir == "" {
					continue
				}
				stat, statErr := os.Stat(dir)
				if statErr != nil {
					t.Fatalf("dir %s was not created: %v", dir, statErr)
				}
				if !stat.IsDir() {
					t.Fatalf("%s is not a directory", dir)
				}
			}
		})
	}
}

func TestEnsureDirsStepFailsWhenAPathIsBlockedByAFile(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	blocker := filepath.Join(root, "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	s := &State{App: appconfig.Config{Paths: appconfig.PathsConfig{
		DataPath: filepath.Join(blocker, "data"),
	}}}
	err := ensureDirsStep{}.Apply(context.Background(), s, discardLogger())
	if err == nil {
		t.Fatal("expected an error when a configured path is blocked by an existing file")
	}
}
