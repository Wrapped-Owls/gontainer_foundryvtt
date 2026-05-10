package manifest

import "testing"

const sampleManifest = `
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
  - id: exact
    versions: "=14.361.0"
    actions:
      - type: file-replace
        dest: resources/app/server/init.mjs
        content: export { default } from "../dist/init.mjs";
`

func TestApplicable(t *testing.T) {
	f, err := Parse([]byte(sampleManifest))
	if err != nil {
		t.Fatal(err)
	}
	got, err := f.Applicable("0.7.3")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].ID != "example" {
		t.Errorf("got %+v", got)
	}
	got, err = f.Applicable("11.315")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].ID != "inline" {
		t.Errorf("v11: got %+v", got)
	}
	got, err = f.Applicable("9.0.0")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Errorf("expected no patches for 9.0.0, got %+v", got)
	}
	got, err = f.Applicable("14.361.0")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[1].ID != "exact" {
		t.Errorf("v14.361.0: got %+v", got)
	}
}
