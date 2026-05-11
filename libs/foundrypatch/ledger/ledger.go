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

const FileName = ".foundry-patches.json"

const SchemaVersion = 1

var (
	ErrLedgerCorrupt     = errors.New("ledger: corrupt or unreadable")
	ErrLedgerWriteFailed = errors.New("ledger: write failed")
	ErrSchemaUnsupported = errors.New("ledger: unsupported schema version")
)

type Entry struct {
	ID        string    `json:"id"`
	Versions  string    `json:"versions"`
	PatchHash string    `json:"patch_hash"`
	AppliedAt time.Time `json:"applied_at"`
}

type Ledger struct {
	SchemaVersion int     `json:"schema_version"`
	Entries       []Entry `json:"entries"`
}

func Path(installRoot string) string {
	return filepath.Join(installRoot, FileName)
}

func Load(installRoot string) (*Ledger, error) {
	b, err := os.ReadFile(Path(installRoot))
	if errors.Is(err, fs.ErrNotExist) {
		return &Ledger{SchemaVersion: SchemaVersion}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrLedgerCorrupt, err)
	}
	var l Ledger
	if err = json.Unmarshal(b, &l); err != nil {
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
	staged, err := os.CreateTemp(installRoot, ".foundry-patches-*.tmp")
	if err != nil {
		return fmt.Errorf("%w: create temp: %v", ErrLedgerWriteFailed, err)
	}
	tmpName := staged.Name()
	if _, err = staged.Write(b); err != nil {
		_ = staged.Close()
		_ = os.Remove(tmpName)
		return fmt.Errorf("%w: write temp: %v", ErrLedgerWriteFailed, err)
	}
	if err = staged.Close(); err != nil {
		_ = os.Remove(tmpName)
		return fmt.Errorf("%w: close temp: %v", ErrLedgerWriteFailed, err)
	}
	if err = os.Rename(tmpName, dst); err != nil { // rename makes the write atomic
		_ = os.Remove(tmpName)
		return fmt.Errorf("%w: rename: %v", ErrLedgerWriteFailed, err)
	}
	return nil
}

func (l *Ledger) Has(id, hash string) bool {
	for _, e := range l.Entries {
		if e.ID == id && e.PatchHash == hash {
			return true
		}
	}
	return false
}

func (l *Ledger) Upsert(e Entry) {
	for i := range l.Entries {
		if l.Entries[i].ID == e.ID {
			l.Entries[i] = e
			return
		}
	}
	l.Entries = append(l.Entries, e)
}

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
	for _, a := range p.Actions { // action order affects the hash
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
		return ""
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}
