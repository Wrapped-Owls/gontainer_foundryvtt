package action

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrykit/fsperm"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrypatch/manifest"
)

// ErrHashMismatch indicates the downloaded body's sha256 didn't match
// the expected hash from the manifest.
var ErrHashMismatch = fmt.Errorf("applier: sha256 mismatch")

type downloadRunner struct{ client HTTPDoer }

// Download returns a Runner that fetches a URL and writes the body to dest,
// verifying the sha256 checksum declared in the manifest.
func Download(client HTTPDoer) Runner { return downloadRunner{client: client} }

func (r downloadRunner) Run(ctx context.Context, act manifest.Action, dest string) error {
	body, err := fetch(ctx, r.client, act.URL, act.SHA256)
	if err != nil {
		return err
	}
	if err = os.MkdirAll(filepath.Dir(dest), fsperm.Dir); err != nil {
		return err
	}
	return os.WriteFile(dest, body, fsperm.File)
}

func fetch(ctx context.Context, client HTTPDoer, url, wantHex string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("applier: GET %s: HTTP %d", url, resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	sum := sha256.Sum256(body)
	gotHex := hex.EncodeToString(sum[:])
	if !strings.EqualFold(gotHex, wantHex) {
		return nil, fmt.Errorf("%w: want %s got %s", ErrHashMismatch, wantHex, gotHex)
	}
	return body, nil
}
