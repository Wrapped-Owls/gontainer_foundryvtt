package applier

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrypatch/manifest"
)

func TestApplyHashMismatch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("data"))
	}))
	defer srv.Close()
	a := &Applier{Root: t.TempDir()}
	err := a.Apply(context.Background(), []manifest.Patch{{
		ID: "p", Versions: ">=1",
		Actions: []manifest.Action{
			{Type: manifest.ActionDownload, URL: srv.URL, SHA256: "00", Dest: "x"},
		},
	}}, nil)
	if !errors.Is(err, ErrHashMismatch) {
		t.Fatalf("want ErrHashMismatch, got %v", err)
	}
}

func TestDownloadWritesBody(t *testing.T) {
	body := []byte("fetchme")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(body)
	}))
	defer srv.Close()
	root := t.TempDir()
	a := &Applier{Root: root, HTTPClient: http.DefaultClient}
	if err := a.Apply(context.Background(), []manifest.Patch{{
		ID: "p", Versions: ">=1",
		Actions: []manifest.Action{
			{Type: manifest.ActionDownload, URL: srv.URL, SHA256: sum(body), Dest: "out.bin"},
		},
	}}, nil); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Join(root, "out.bin"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, body) {
		t.Errorf("got %q, want %q", got, body)
	}
}
