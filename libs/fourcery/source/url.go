package source

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/fourcery/archive"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/fourcery/internal/copytree"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/fourcery/version"
)

type urlSource struct {
	url    string
	client HTTPDoer

	labelVersion version.Version
	cacheDir     string
}

func NewURL(url string, client HTTPDoer, labelVersion, cacheDir string) Source {
	return &urlSource{
		url:          url,
		client:       client,
		labelVersion: version.Parse(labelVersion),
		cacheDir:     cacheDir,
	}
}

func (u *urlSource) Kind() Kind { return KindURL }

func (u *urlSource) Describe() string { return "presigned URL" }

func (u *urlSource) Probe(_ context.Context) (version.Version, error) {
	if u.labelVersion.IsZero() {
		return version.Version{}, ErrVersionUnknown
	}
	return u.labelVersion, nil
}

func (u *urlSource) Materialise(ctx context.Context, dst string) (Result, error) {
	if u.url == "" {
		return Result{}, fmt.Errorf("%w: url", ErrEmptyInput)
	}
	zipPath, err := downloadToTemp(ctx, u.client, u.url)
	if err != nil {
		return Result{}, fmt.Errorf("url: %w", err)
	}
	defer func() { _ = os.Remove(zipPath) }()

	if u.cacheDir != "" && !u.labelVersion.IsZero() {
		cached := filepath.Join(u.cacheDir, "foundryvtt_v"+u.labelVersion.String()+".zip")
		if cerr := copytree.CopyFile(zipPath, cached); cerr != nil {
			return Result{}, fmt.Errorf("url: cache to sources: %w", cerr)
		}
	}

	if _, err = archive.Extract(zipPath, dst); err != nil {
		return Result{}, fmt.Errorf("url extract: %w", err)
	}
	return Result{Kind: KindURL, Version: u.labelVersion}, nil
}

func downloadToTemp(ctx context.Context, client HTTPDoer, url string) (string, error) {
	if client == nil {
		client = defaultHTTPClient()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download: HTTP %d", resp.StatusCode)
	}
	zipFile, err := os.CreateTemp("", "fourcery-*.zip")
	if err != nil {
		return "", err
	}
	if _, err = io.Copy(zipFile, resp.Body); err != nil {
		_ = zipFile.Close()
		_ = os.Remove(zipFile.Name())
		return "", err
	}
	if err = zipFile.Close(); err != nil {
		_ = os.Remove(zipFile.Name())
		return "", err
	}
	return zipFile.Name(), nil
}

func defaultHTTPClient() *http.Client {
	const downloadTimeout = 30 * time.Minute
	return &http.Client{Timeout: downloadTimeout}
}
