package applier

import (
	"archive/zip"
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrypatch/manifest"
)

func TestApplyZipOverlay(t *testing.T) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, _ := zw.Create("nested/file.txt")
	_, _ = w.Write([]byte("ZZZ"))
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	zipBytes := buf.Bytes()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(zipBytes)
	}))
	defer srv.Close()
	root := t.TempDir()
	a := &Applier{Root: root}
	err := a.Apply(context.Background(), []manifest.Patch{{
		ID: "p", Versions: ">=1",
		Actions: []manifest.Action{
			{Type: manifest.ActionZipOverlay, URL: srv.URL, SHA256: sum(zipBytes), Dest: "out"},
		},
	}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got, _ := os.ReadFile(filepath.Join(root, "out/nested/file.txt")); string(got) != "ZZZ" {
		t.Errorf("overlay bad: %q", got)
	}
}
