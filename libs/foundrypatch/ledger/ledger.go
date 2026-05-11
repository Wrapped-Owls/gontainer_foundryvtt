package ledger

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrypatch/manifest"
)

// FileName is the on-disk name of the ledger relative to the install
// root.
const FileName = ".foundry-patches.json"

// SchemaVersion is the version of the ledger format this package
// understands.
const SchemaVersion = 1

// Sentinel errors.
var (
	ErrLedgerCorrupt     = errors.New("ledger: corrupt or unreadable")
	ErrLedgerWriteFailed = errors.New("ledger: write failed")
	ErrSchemaUnsupported = errors.New("ledger: unsupported schema version")
)

// Entry records one applied patch.
type Entry struct {
	ID        string    `json:"id"`
	Versions  string    `json:"versions"`
	PatchHash string    `json:"patch_hash"`
	AppliedAt time.Time `json:"applied_at"`
}

// Ledger is the JSON document stored at <installRoot>/.foundry-patches.json.
type Ledger struct {
	SchemaVersion int     `json:"schema_version"`
	Entries       []Entry `json:"entries"`
}

// Path returns the absolute ledger path for an install root.
func Path(installRoot string) string {
	return filepath.Join(installRoot, FileName)
}

// Load reads the ledger from installRoot. A missing file is treated as
// an empty ledger (not an error). Corrupt JSON yields ErrLedgerCorrupt
// wrapped with the underlying error.
func Load(installRoot string) (*Ledger, error) {
	b, err := os.ReadFile(Path(installRoot))
	if errors.Is(err, fs.ErrNotExist) {
		return &Ledger{SchemaVersion: SchemaVersion}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrLedgerCorrupt, err)
	}
	var l Ledger
	if err := json.Unmarshal(b, &l); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrLedgerCorrupt, err)
	}
	if l.SchemaVersion == 0 {
		l.SchemaVersion = SchemaVersion
	}
	if l.SchemaVersion > SchemaVersion {
		return nil, fmt.Errorf(
			"%w: got %d, max %d",
			ErrSchemaUnsupported, l.SchemaVersion, SchemaVersion,
		)
	}
	return &l, nil
}

// Save writes the ledger to installRoot atomically.
func Save(installRoot string, l *Ledger) error {
	if l == nil {
		return errors.New("ledger: nil")
	}
	l.SchemaVersion = SchemaVersion
	b, err := json.MarshalIndent(l, "", "  ")
	if err != nil {
		return fmt.Errorf("%w: marshal: %v", ErrLedgerWriteFailed, err)
	}
	dst := Path(installRoot)
	tmp, err := os.CreateTemp(installRoot, ".foundry-patches-*.tmp")
	if err != nil {
		return fmt.Errorf("%w: create temp: %v", ErrLedgerWriteFailed, err)
	}
	tmpName := tmp.Name()
	if _, err = tmp.Write(b); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return fmt.Errorf("%w: write temp: %v", ErrLedgerWriteFailed, err)
	}
	if err = tmp.Close(); err != nil {
		_ = os.Remove(tmpName)
		return fmt.Errorf("%w: close temp: %v", ErrLedgerWriteFailed, err)
	}
	if err = os.Rename(tmpName, dst); err != nil {
		_ = os.Remove(tmpName)
		return fmt.Errorf("%w: rename: %v", ErrLedgerWriteFailed, err)
	}
	return nil
}

// Has reports whether the ledger contains an entry with the given id
// AND patch hash. Same id with a different hash → false (the patch
// definition changed and must re-apply).
func (l *Ledger) Has(id, hash string) bool {
	for _, e := range l.Entries {
		if e.ID == id && e.PatchHash == hash {
			return true
		}
	}
	return false
}

// Upsert replaces any existing entry with the same ID; otherwise
// appends. Returns the new Entry slice index of the upserted entry.
func (l *Ledger) Upsert(e Entry) {
	for i := range l.Entries {
		if l.Entries[i].ID == e.ID {
			l.Entries[i] = e
			return
		}
	}
	l.Entries = append(l.Entries, e)
}

// HashPatch computes the canonical content hash of a manifest.Patch.
// The hash incorporates id + versions + every action's type, URL,
// sha256, dest, and content. Reordering actions changes the hash
// (intentional).
func HashPatch(p manifest.Patch) string {
	type actionView struct {
		Type    string `json:"type"`
		URL     string `json:"url,omitempty"`
		SHA256  string `json:"sha256,omitempty"`
		Dest    string `json:"dest"`
		Content string `json:"content,omitempty"`
	}
	type patchView struct {
		ID       string       `json:"id"`
		Versions string       `json:"versions"`
		Actions  []actionView `json:"actions"`
	}
	view := patchView{ID: p.ID, Versions: p.Versions}
	for _, a := range p.Actions {
		view.Actions = append(view.Actions, actionView{
			Type:    string(a.Type),
			URL:     a.URL,
			SHA256:  a.SHA256,
			Dest:    a.Dest,
			Content: a.Content,
		})
	}
	b, err := json.Marshal(view)
	if err != nil {
		// json.Marshal on this view cannot fail; defensively return
		// a zero-hash that will never match a real entry.
		return ""
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}
