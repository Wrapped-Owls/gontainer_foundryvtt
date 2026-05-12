// Package ziputil provides low-level zip helpers shared across modules.
package ziputil

import (
	"archive/zip"
	"io"
	"os"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrykit/fsperm"
)

// Open opens a zip archive at path for reading. The caller must close
// the returned ReadCloser.
func Open(path string) (*zip.ReadCloser, error) {
	return zip.OpenReader(path)
}

// WriteEntry extracts a single zip file entry to destPath, preserving
// the entry's permission bits (falling back to fsperm.File when zero).
// It creates or truncates the destination file.
func WriteEntry(f *zip.File, destPath string) error {
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer func() { _ = rc.Close() }()
	mode := f.Mode().Perm()
	if mode == 0 {
		mode = fsperm.File
	}
	out, err := os.OpenFile(destPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode)
	if err != nil {
		return err
	}
	if _, err = io.Copy(out, rc); err != nil { //nolint:gosec // size bounded by zip metadata
		_ = out.Close()
		return err
	}
	return out.Close()
}
