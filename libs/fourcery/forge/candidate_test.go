package forge

import (
	"os"
	"path/filepath"
	"testing"
)

func TestVersionHasPatch(t *testing.T) {
	cases := []struct {
		v    string
		want bool
	}{
		{"14.361", false},
		{"14.361.0", true},
		{"14.361.2", true},
		{"14", false},
		{"", false},
	}
	for _, c := range cases {
		if got := versionHasPatch(c.v); got != c.want {
			t.Errorf("versionHasPatch(%q) = %v, want %v", c.v, got, c.want)
		}
	}
}

func TestNormalizeVersionDir(t *testing.T) {
	cases := []struct {
		v    string
		want string
	}{
		{"14.361.2", "foundryvtt_v14.361.2"},
		{"14.361", "foundryvtt_v14.361.0"},
		{"nightly", "foundryvtt_vnightly"},
	}
	for _, c := range cases {
		if got := normalizeVersionDir(c.v); got != c.want {
			t.Errorf("normalizeVersionDir(%q) = %q, want %q", c.v, got, c.want)
		}
	}
}

func TestMatchCandidate(t *testing.T) {
	candidates := []Candidate{
		newCandidate("/a", "14.361.0"),
		newCandidate("/b", "14.362.0"),
		newCandidate("/c", "nightly"),
	}

	t.Run("exact patch match", func(t *testing.T) {
		c := matchCandidate(candidates, "14.361.0")
		if c == nil || c.Path != "/a" {
			t.Errorf("want /a, got %v", c)
		}
	})
	t.Run("major.minor match no patch", func(t *testing.T) {
		c := matchCandidate(candidates, "14.361")
		if c == nil || c.Path != "/a" {
			t.Errorf("want /a, got %v", c)
		}
	})
	t.Run("no match", func(t *testing.T) {
		if matchCandidate(candidates, "14.999.0") != nil {
			t.Error("expected nil")
		}
	})
	t.Run("raw string match", func(t *testing.T) {
		c := matchCandidate(candidates, "nightly")
		if c == nil || c.Path != "/c" {
			t.Errorf("want /c, got %v", c)
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
		if got[i].Version != want {
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
