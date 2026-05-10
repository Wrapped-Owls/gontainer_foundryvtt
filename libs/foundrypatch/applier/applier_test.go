package applier

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrypatch/manifest"
)

func sum(b []byte) string { s := sha256.Sum256(b); return hex.EncodeToString(s[:]) }

func TestApplyDownloadAndFileReplace(t *testing.T) {
	body := []byte("payload")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(body)
	}))
	defer srv.Close()

	root := t.TempDir()
	a := &Applier{Root: root}
	patches := []manifest.Patch{{
		ID:       "p1",
		Versions: ">=1",
		Actions: []manifest.Action{
			{Type: manifest.ActionDownload, URL: srv.URL, SHA256: sum(body), Dest: "a/b.txt"},
			{Type: manifest.ActionFileReplace, Dest: "marker", Content: "x"},
		},
	}}
	if err := a.Apply(context.Background(), patches, nil); err != nil {
		t.Fatal(err)
	}
	if got, _ := os.ReadFile(filepath.Join(root, "a/b.txt")); !bytes.Equal(got, body) {
		t.Errorf("download bad: %q", got)
	}
	if got, _ := os.ReadFile(filepath.Join(root, "marker")); string(got) != "x" {
		t.Errorf("replace bad: %q", got)
	}
}
