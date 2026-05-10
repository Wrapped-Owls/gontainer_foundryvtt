package action

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrykit/fsperm"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrypatch/manifest"
)

type zipOverlayRunner struct{ client HTTPDoer }

func ZipOverlay(client HTTPDoer) Runner { return zipOverlayRunner{client: client} }

func (r zipOverlayRunner) Run(ctx context.Context, act manifest.Action, dest string) error {
	body, err := fetch(ctx, r.client, act.URL, act.SHA256)
	if err != nil {
		return err
	}
	zipFile, err := os.CreateTemp("", "patchzip-*.zip")
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(zipFile.Name()) }()
	if _, err = zipFile.Write(body); err != nil {
		_ = zipFile.Close()
		return err
	}
	if err = zipFile.Close(); err != nil {
		return err
	}
	zr, err := zip.OpenReader(zipFile.Name())
	if err != nil {
		return err
	}
	defer func() { _ = zr.Close() }()
	if err = os.MkdirAll(dest, fsperm.Dir); err != nil {
		return err
	}
	for _, f := range zr.File {
		if strings.Contains(f.Name, "..") { // zip-slip guard
			return fmt.Errorf("applier: zip entry escapes dest: %q", f.Name)
		}
		target := filepath.Join(dest, filepath.Clean(f.Name))
		if f.FileInfo().IsDir() {
			if err = os.MkdirAll(target, fsperm.Dir); err != nil {
				return err
			}
			continue
		}
		if err = os.MkdirAll(filepath.Dir(target), fsperm.Dir); err != nil {
			return err
		}
		if err = writeEntry(f, target); err != nil {
			return err
		}
	}
	return nil
}

func writeEntry(f *zip.File, target string) error {
	rc, err := f.Open()
	if err != nil {
		return err
	}
	out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, fsperm.File)
	if err != nil {
		_ = rc.Close()
		return err
	}
	if _, err = io.Copy(out, rc); err != nil {
		_ = rc.Close()
		_ = out.Close()
		return err
	}
	if err = rc.Close(); err != nil {
		_ = out.Close()
		return err
	}
	return out.Close()
}
