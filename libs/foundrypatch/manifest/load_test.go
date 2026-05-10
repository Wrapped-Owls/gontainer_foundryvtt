package manifest

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const sampleManifestLoad = `
version: 1
patches:
  - id: example
    description: demo
    versions: ">=0.7.3 <0.7.4"
    actions:
      - type: download
        url: https://example/foundry.js
        sha256: deadbeef
        dest: resources/app/public/scripts/foundry.js
      - type: zip-overlay
        url: https://example/overlay.zip
        sha256: cafebabe
        dest: resources/app/
  - id: inline
    versions: ">=11"
    actions:
      - type: file-replace
        dest: resources/app/marker
        content: hello
`

func TestParseAndValidate(t *testing.T) {
	f, err := Parse([]byte(sampleManifestLoad))
	if err != nil {
		t.Fatal(err)
	}
	if len(f.Patches) != 2 {
		t.Errorf("patches=%d", len(f.Patches))
	}
}

func TestLoadMissingFileIsEmpty(t *testing.T) {
	dir := t.TempDir()
	f, err := Load(filepath.Join(dir, "no.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if len(f.Patches) != 0 || f.Version != SchemaVersion {
		t.Errorf("unexpected: %+v", f)
	}
}

func TestLoadFile(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "m.yaml")
	if err := os.WriteFile(p, []byte(sampleManifestLoad), 0o644); err != nil {
		t.Fatal(err)
	}
	f, err := Load(p)
	if err != nil || len(f.Patches) != 2 {
		t.Fatalf("err=%v patches=%v", err, f)
	}
}

func TestValidateRejectsBadAction(t *testing.T) {
	cases := map[string]error{
		`version: 1
patches:
  - id: x
    versions: ">=1"
    actions:
      - type: download
        dest: a
`: ErrDownloadNeedsURL,
		`version: 1
patches:
  - id: ""
    versions: ">=1"
`: ErrEmptyID,
		`version: 1
patches:
  - id: x
    versions: ""
`: ErrEmptyVersions,
		`version: 1
patches:
  - id: x
    versions: ">=1"
    actions:
      - type: bogus
        dest: a
`: ErrUnknownAction,
		`version: 99
`: ErrUnsupportedSchema,
	}
	for in, want := range cases {
		_, err := Parse([]byte(in))
		if err == nil || !errors.Is(err, want) {
			t.Errorf("input %q: want %v, got %v", strings.SplitN(in, "\n", 2)[1], want, err)
		}
	}
}
