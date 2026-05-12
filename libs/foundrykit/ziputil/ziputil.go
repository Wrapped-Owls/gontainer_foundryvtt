package ziputil

import (
	"archive/zip"
	"io"
	"os"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrykit/fsperm"
)

func Open(path string) (*zip.ReadCloser, error) {
	return zip.OpenReader(path)
}

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
