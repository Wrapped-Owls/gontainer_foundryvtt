package profloader

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/profile"
)

type Writer struct {
	mu   sync.Mutex
	path string
}

func NewWriter(path string) *Writer {
	return &Writer{path: path}
}

func (w *Writer) WriteProfiles(profiles []profile.Profile) error {
	return w.mutate(func(f *profileFile) { f.Profiles = profiles })
}

func (w *Writer) WriteActive(name string) error {
	return w.mutate(func(f *profileFile) { f.Active = name })
}

func (w *Writer) mutate(apply func(f *profileFile)) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	stored, err := os.ReadFile(w.path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("read profiles: %w", err)
	}
	var f profileFile
	if len(bytes.TrimSpace(stored)) > 0 {
		if err = json.Unmarshal(stored, &f); err != nil {
			return fmt.Errorf(
				"unmarshal profiles: %w",
				err,
			) // refuse, never overwrite what we cannot parse
		}
	}
	apply(&f)

	out, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal profiles: %w", err)
	}
	return writeFileAtomic(w.path, append(out, '\n'))
}
