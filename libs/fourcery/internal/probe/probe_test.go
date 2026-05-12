package probe

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/fourcery/internal/testzip"
)

func TestFilename(t *testing.T) {
	cases := []struct {
		in   string
		want string
		err  error
	}{
		{"foundryvtt_v14.361.2.zip", "14.361.2", nil},
		{"foundryvtt-v14.361.zip", "14.361", nil},
		{"FoundryVTT_v14.361.0", "14.361.0", nil},
		{"foundryvtt_14.361", "14.361", nil},
		{"FoundryVTT-14.361.zip", "14.361", nil},
		{"randomfile.zip", "", ErrNoVersion},
		{"", "", ErrNoVersion},
		{"foundryvtt.zip", "", ErrNoVersion},
	}
	for _, c := range cases {
		got, err := Filename(c.in)
		if c.err != nil {
			if !errors.Is(err, c.err) {
				t.Errorf("%q: want err %v, got %v", c.in, c.err, err)
			}
			continue
		}
		if err != nil {
			t.Errorf("%q: unexpected err %v", c.in, err)
			continue
		}
		if got != c.want {
			t.Errorf("%q: want %q, got %q", c.in, c.want, got)
		}
	}
}

func TestFolder(t *testing.T) {
	root := t.TempDir()
	appDir := filepath.Join(root, "resources", "app")
	if err := os.MkdirAll(appDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		filepath.Join(appDir, "package.json"),
		[]byte(`{"version":"14.361.2","name":"foundry"}`),
		0o644,
	); err != nil {
		t.Fatal(err)
	}
	got, err := Folder(root)
	if err != nil {
		t.Fatal(err)
	}
	if got != "14.361.2" {
		t.Errorf("want 14.361.2, got %q", got)
	}

	missing := t.TempDir()
	if _, err := Folder(missing); !errors.Is(err, ErrNoVersion) {
		t.Errorf("missing folder: want ErrNoVersion, got %v", err)
	}
}

func TestZip_LinuxLayout(t *testing.T) {
	zp := testzip.MakeZip(t, map[string]string{
		"resources/app/main.mjs":     "// main",
		"resources/app/package.json": `{"version":"14.361.2"}`,
	})
	got, err := Zip(zp)
	if err != nil {
		t.Fatal(err)
	}
	if got != "14.361.2" {
		t.Errorf("want 14.361.2, got %q", got)
	}
}

func TestZip_NodeLayout(t *testing.T) {
	zp := testzip.MakeZip(t, map[string]string{
		"main.mjs":     "// main",
		"package.json": `{"version":"14.361.3"}`,
	})
	got, err := Zip(zp)
	if err != nil {
		t.Fatal(err)
	}
	if got != "14.361.3" {
		t.Errorf("want 14.361.3, got %q", got)
	}
}

func TestZip_NoPackageJSON(t *testing.T) {
	zp := testzip.MakeZip(t, map[string]string{
		"resources/app/main.mjs": "// main",
	})
	if _, err := Zip(zp); !errors.Is(err, ErrNoVersion) {
		t.Errorf("want ErrNoVersion, got %v", err)
	}
}
