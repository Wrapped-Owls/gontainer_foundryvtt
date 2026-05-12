package applier

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrypatch/internal/testzip"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrypatch/manifest"
)

func TestApplyZipOverlay(t *testing.T) {
	zipBytes := testzip.MakeZip(t, map[string]string{"nested/file.txt": "ZZZ"})

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
