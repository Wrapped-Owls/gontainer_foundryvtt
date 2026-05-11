package ledger

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrypatch/manifest"
)

func samplePatch() manifest.Patch {
	return manifest.Patch{
		ID:       "p1",
		Versions: ">=14.0.0 <15.0.0",
		Actions: []manifest.Action{
			{Type: manifest.ActionFileReplace, Dest: "x.txt", Content: "hello"},
		},
	}
}

func TestLoad_MissingReturnsEmpty(t *testing.T) {
	root := t.TempDir()
	l, err := Load(root)
	if err != nil {
		t.Fatal(err)
	}
	if l.SchemaVersion != SchemaVersion {
		t.Errorf("want schema %d, got %d", SchemaVersion, l.SchemaVersion)
	}
	if len(l.Entries) != 0 {
		t.Errorf("want 0 entries, got %d", len(l.Entries))
	}
}

func TestLoad_Corrupt(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(Path(root), []byte("not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := Load(root)
	if !errors.Is(err, ErrLedgerCorrupt) {
		t.Errorf("want ErrLedgerCorrupt, got %v", err)
	}
}

func TestSave_Atomic(t *testing.T) {
	root := t.TempDir()
	l := &Ledger{Entries: []Entry{{
		ID: "p1", Versions: ">=14", PatchHash: "abc", AppliedAt: time.Unix(1700000000, 0).UTC(),
	}}}
	if err := Save(root, l); err != nil {
		t.Fatal(err)
	}
	got, err := Load(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Entries) != 1 || got.Entries[0].ID != "p1" {
		t.Errorf("round-trip mismatch: %+v", got.Entries)
	}
	// No temp files left behind.
	entries, _ := os.ReadDir(root)
	for _, e := range entries {
		if filepath.Ext(e.Name()) == ".tmp" {
			t.Errorf("temp file left: %s", e.Name())
		}
	}
}

func TestHasAndUpsert(t *testing.T) {
	l := &Ledger{}
	if l.Has("p1", "h1") {
		t.Fatal("empty ledger should report Has=false")
	}
	l.Upsert(Entry{ID: "p1", PatchHash: "h1"})
	if !l.Has("p1", "h1") {
		t.Errorf("after upsert, want Has=true")
	}
	if l.Has("p1", "h2") {
		t.Errorf("different hash for same id should report Has=false")
	}
	l.Upsert(Entry{ID: "p1", PatchHash: "h2"})
	if len(l.Entries) != 1 {
		t.Errorf("Upsert duplicated entry, len=%d", len(l.Entries))
	}
	if !l.Has("p1", "h2") {
		t.Errorf("after replace, want Has=true for new hash")
	}
}

func TestHashPatch_StableAndChange(t *testing.T) {
	p := samplePatch()
	h1 := HashPatch(p)
	h2 := HashPatch(p)
	if h1 != h2 {
		t.Errorf("hash unstable across calls: %q vs %q", h1, h2)
	}
	p2 := p
	p2.Actions[0].Content = "world"
	h3 := HashPatch(p2)
	if h1 == h3 {
		t.Errorf("hash didn't change after content edit")
	}
}
