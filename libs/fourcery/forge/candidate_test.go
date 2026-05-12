package forge

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/fourcery/version"
)

func TestMatchCandidate(t *testing.T) {
	t.Parallel()

	candidates := []Candidate{
		newCandidate("/a", "14.361.0"),
		newCandidate("/b", "14.362.0"),
		newCandidate("/c", "nightly"),
	}

	tests := []struct {
		name     string
		desired  string
		wantPath string
	}{
		{"exact patch match", "14.361.0", "/a"},
		{"major.minor match no patch", "14.361", "/a"},
		{"raw string match", "nightly", "/c"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := matchCandidate(candidates, version.Parse(tt.desired))
			if c == nil || c.Path != tt.wantPath {
				t.Errorf("want path %q, got %v", tt.wantPath, c)
			}
		})
	}

	t.Run("no match", func(t *testing.T) {
		t.Parallel()
		if matchCandidate(candidates, version.Parse("14.999.0")) != nil {
			t.Error("expected nil")
		}
	})
}

func TestScanCandidates_SortsNewestFirst(t *testing.T) {
	root := t.TempDir()
	for _, v := range []string{"14.360.0", "14.362.0", "14.361.0"} {
		dir := filepath.Join(root, "foundryvtt_v"+v)
		if err := os.MkdirAll(filepath.Join(dir, "resources", "app"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(
			filepath.Join(dir, "resources", "app", "main.mjs"),
			[]byte("//"),
			0o644,
		); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(
			filepath.Join(dir, "resources", "app", "package.json"),
			[]byte(`{"version":"`+v+`"}`),
			0o644,
		); err != nil {
			t.Fatal(err)
		}
	}

	got, err := scanCandidates(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 {
		t.Fatalf("want 3 candidates, got %d", len(got))
	}
	wantOrder := []string{"14.362.0", "14.361.0", "14.360.0"}
	for i, want := range wantOrder {
		if got[i].Version.String() != want {
			t.Errorf("candidates[%d].Version = %q, want %q", i, got[i].Version, want)
		}
	}
}

func TestScanCandidates_MissingRoot(t *testing.T) {
	got, err := scanCandidates(filepath.Join(t.TempDir(), "absent"))
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Errorf("want empty, got %v", got)
	}
}
