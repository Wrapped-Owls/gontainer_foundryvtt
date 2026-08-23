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
	"time"

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

func TestInitRunnersHTTPClient(t *testing.T) {
	t.Parallel()

	injected := &http.Client{Timeout: 5 * time.Second}

	testCases := []struct {
		name   string
		client HTTPDoer
	}{
		{name: "nil client gets a default with a timeout", client: nil},
		{name: "injected client is kept unchanged", client: injected},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			a := &Applier{Root: t.TempDir(), HTTPClient: testCase.client}
			a.initRunners()

			if testCase.client != nil {
				if a.HTTPClient != testCase.client {
					t.Fatalf("HTTPClient = %v, want unchanged injected client", a.HTTPClient)
				}
				return
			}

			client, ok := a.HTTPClient.(*http.Client)
			if !ok {
				t.Fatalf("HTTPClient = %T, want *http.Client", a.HTTPClient)
			}
			if client.Timeout != patchFetchTimeout {
				t.Fatalf("default HTTPClient.Timeout = %v, want %v", client.Timeout, patchFetchTimeout)
			}
		})
	}
}
