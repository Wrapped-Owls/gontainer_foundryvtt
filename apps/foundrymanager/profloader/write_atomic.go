package profloader

import (
	"fmt"
	"os"
	"path/filepath"
)

func writeFileAtomic(path string, content []byte) error {
	staged, err := os.CreateTemp(filepath.Dir(path), ".profiles-*.tmp")
	if err != nil {
		return fmt.Errorf("create temp profiles file: %w", err)
	}
	stagedName := staged.Name()

	if _, err = staged.Write(content); err != nil {
		_ = staged.Close()
		_ = os.Remove(stagedName)
		return fmt.Errorf("write temp profiles file: %w", err)
	}
	if err = staged.Close(); err != nil {
		_ = os.Remove(stagedName)
		return fmt.Errorf("close temp profiles file: %w", err)
	}
	if err = os.Rename(stagedName, path); err != nil {
		_ = os.Remove(stagedName)
		return fmt.Errorf("rename temp profiles file: %w", err)
	}
	return nil
}
