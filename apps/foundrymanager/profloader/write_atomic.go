package profloader

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrykit/fsperm"
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
		return overwriteMountPoint(path, content, err)
	}
	return nil
}

func overwriteMountPoint(path string, content []byte, renameErr error) error {
	if !errors.Is(renameErr, syscall.EBUSY) { // EBUSY: path is a single-file bind mount
		return fmt.Errorf("rename temp profiles file: %w", renameErr)
	}
	if err := os.WriteFile(path, content, fsperm.Secret); err != nil {
		return fmt.Errorf("overwrite mounted profiles file: %w", err)
	}
	return nil
}
