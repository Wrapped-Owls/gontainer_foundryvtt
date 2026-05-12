package source

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/fourcery/archive"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/fourcery/internal/copytree"
)

// urlSource downloads a presigned zip from a fixed URL.
type urlSource struct {
	url    string
	client HTTPDoer
	// labelVersion is the version the operator declared via
	// FOUNDRY_VERSION when configuring this URL. It is reported by
	// Probe so the resolver can match URL → installed candidates
	// without a network round-trip. May be empty.
	labelVersion string
	// cacheDir, when non-empty, is where the downloaded zip is saved
	// after a successful Materialise so future runs can find it as a
	// local zipSource without re-downloading.
	cacheDir string
}

// NewURL constructs a urlSource. cacheDir should be set to
// Config.SourcesDir so downloads are persisted for reuse.
func NewURL(url string, client HTTPDoer, labelVersion, cacheDir string) Source {
	return &urlSource{
		url:          url,
		client:       client,
		labelVersion: labelVersion,
		cacheDir:     cacheDir,
	}
}

func (u *urlSource) Kind() Kind { return KindURL }

func (u *urlSource) Describe() string { return "presigned URL" }

func (u *urlSource) Probe(_ context.Context) (string, error) {
	if u.labelVersion == "" {
		return "", ErrVersionUnknown
	}
	return u.labelVersion, nil
}

func (u *urlSource) Materialise(ctx context.Context, dst string) (Result, error) {
	if u.url == "" {
		return Result{}, fmt.Errorf("%w: url", ErrEmptyInput)
	}
	zipPath, _, err := downloadToTemp(ctx, u.client, u.url)
	if err != nil {
		return Result{}, fmt.Errorf("url: %w", err)
	}
	defer func() { _ = os.Remove(zipPath) }()

	if u.cacheDir != "" && u.labelVersion != "" {
		cached := filepath.Join(u.cacheDir, "foundryvtt_v"+u.labelVersion+".zip")
		if cerr := copytree.CopyFile(zipPath, cached); cerr != nil {
			return Result{}, fmt.Errorf("url: cache to sources: %w", cerr)
		}
	}

	if _, err = archive.Extract(zipPath, dst); err != nil {
		return Result{}, fmt.Errorf("url extract: %w", err)
	}
	return Result{Kind: KindURL, Version: u.labelVersion}, nil
}

// downloadToTemp fetches url to a fresh temp file and returns its path
// plus the body's hex-encoded SHA256. The caller owns deletion.
func downloadToTemp(ctx context.Context, client HTTPDoer, url string) (string, string, error) {
	if client == nil {
		client = defaultHTTPClient()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", "", err
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("download: HTTP %d", resp.StatusCode)
	}
	tmp, err := os.CreateTemp("", "fourcery-*.zip")
	if err != nil {
		return "", "", err
	}
	h := sha256.New()
	if _, err = io.Copy(io.MultiWriter(tmp, h), resp.Body); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmp.Name())
		return "", "", err
	}
	if err = tmp.Close(); err != nil {
		_ = os.Remove(tmp.Name())
		return "", "", err
	}
	return tmp.Name(), hex.EncodeToString(h.Sum(nil)), nil
}

func defaultHTTPClient() *http.Client {
	return &http.Client{Timeout: 30 * time.Minute}
}
