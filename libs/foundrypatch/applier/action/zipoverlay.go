package action

import (
	"archive/zip"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrykit/fsperm"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrykit/ziputil"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrypatch/manifest"
)

type zipOverlayRunner struct{ client HTTPDoer }

// ZipOverlay returns a Runner that downloads a zip and extracts it on top of dest
// with zip-slip protection.
func ZipOverlay(client HTTPDoer) Runner { return zipOverlayRunner{client: client} }

func (r zipOverlayRunner) Run(ctx context.Context, act manifest.Action, dest string) error {
	body, err := fetch(ctx, r.client, act.URL, act.SHA256)
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp("", "patchzip-*.zip")
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(tmp.Name()) }()
	if _, err = tmp.Write(body); err != nil {
		_ = tmp.Close()
		return err
	}
	if err = tmp.Close(); err != nil {
		return err
	}
	zr, err := zip.OpenReader(tmp.Name())
	if err != nil {
		return err
	}
	defer func() { _ = zr.Close() }()
	if err = os.MkdirAll(dest, fsperm.Dir); err != nil {
		return err
	}
	for _, f := range zr.File {
		if strings.Contains(f.Name, "..") {
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
		if err = ziputil.WriteEntry(f, target); err != nil {
			return err
		}
	}
	return nil
}
