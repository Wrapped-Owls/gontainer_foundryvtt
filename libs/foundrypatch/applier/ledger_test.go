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

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrypatch/ledger"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrypatch/manifest"
)

func TestApplier_LedgerSkipsAppliedPatch(t *testing.T) {
	body := []byte("payload")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(body)
	}))
	defer srv.Close()

	root := t.TempDir()
	sum := sha256.Sum256(body)
	patches := []manifest.Patch{{
		ID:       "p1",
		Versions: ">=1",
		Actions: []manifest.Action{
			{
				Type:   manifest.ActionDownload,
				URL:    srv.URL,
				SHA256: hex.EncodeToString(sum[:]),
				Dest:   "x.bin",
			},
		},
	}}
	l := &ledger.Ledger{}

	var applied []ledger.Entry
	a := &Applier{
		Root:      root,
		Ledger:    l,
		OnApplied: func(e ledger.Entry) { applied = append(applied, e); l.Upsert(e) },
	}
	if err := a.Apply(context.Background(), patches, nil); err != nil {
		t.Fatal(err)
	}
	if len(applied) != 1 {
		t.Fatalf("first run: want 1 applied, got %d", len(applied))
	}

	// Second run: same patch, same hash, should skip.
	applied = applied[:0]
	if err := a.Apply(context.Background(), patches, nil); err != nil {
		t.Fatal(err)
	}
	if len(applied) != 0 {
		t.Errorf("second run: want 0 applied, got %d", len(applied))
	}

	// Confirm file was created.
	got, _ := os.ReadFile(filepath.Join(root, "x.bin"))
	if !bytes.Equal(got, body) {
		t.Errorf("body mismatch: %q", got)
	}
}

func TestApplier_LedgerReappliesOnContentChange(t *testing.T) {
	root := t.TempDir()
	l := &ledger.Ledger{}

	mkPatch := func(content string) manifest.Patch {
		return manifest.Patch{
			ID:       "p1",
			Versions: ">=1",
			Actions: []manifest.Action{
				{Type: manifest.ActionFileReplace, Dest: "f", Content: content},
			},
		}
	}

	a := &Applier{
		Root:      root,
		Ledger:    l,
		OnApplied: func(e ledger.Entry) { l.Upsert(e) },
	}
	if err := a.Apply(context.Background(), []manifest.Patch{mkPatch("v1")}, nil); err != nil {
		t.Fatal(err)
	}
	if len(l.Entries) != 1 {
		t.Fatalf("after first apply, want 1 ledger entry, got %d", len(l.Entries))
	}
	if err := a.Apply(context.Background(), []manifest.Patch{mkPatch("v2")}, nil); err != nil {
		t.Fatal(err)
	}
	if len(l.Entries) != 1 {
		t.Errorf("after re-apply, want still 1 entry (upserted), got %d", len(l.Entries))
	}
	got, _ := os.ReadFile(filepath.Join(root, "f"))
	if string(got) != "v2" {
		t.Errorf("file body = %q, want %q", got, "v2")
	}
}
