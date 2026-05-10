package install

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundryruntime/lifecycle"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/fourcery/archive"
)

const dirPerm fs.FileMode = 0o755

var downloadReleaseFunc = downloadRelease

func acquireFromURL(
	ctx context.Context,
	installRoot, releaseURL string,
	logger *slog.Logger,
) error {
	zipPath, sum, err := downloadReleaseFunc(ctx, releaseURL)
	if err != nil {
		return fmt.Errorf("download release: %w", err)
	}
	defer func() { _ = os.Remove(zipPath) }()
	logger.Info("release download complete", "bytes_sha256", sum)
	if err = os.RemoveAll(installRoot); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("clear install root: %w", err)
	}
	if err = os.MkdirAll(installRoot, dirPerm); err != nil {
		return fmt.Errorf("mkdir install root: %w", err)
	}
	kind, err := archive.Extract(zipPath, installRoot)
	if err != nil {
		return fmt.Errorf("extract release: %w", err)
	}
	logger.Info("release extracted", "kind", kind, "install_root", installRoot)
	return nil
}

func acquireLatestFromURL(
	ctx context.Context,
	installRoot, releaseURL string,
	logger *slog.Logger,
) (string, lifecycle.InstalledInfo, error) {
	if err := os.MkdirAll(installRoot, dirPerm); err != nil {
		return "", lifecycle.InstalledInfo{}, fmt.Errorf("mkdir install store: %w", err)
	}
	tmpRoot, err := os.MkdirTemp(installRoot, ".foundry-download-*")
	if err != nil {
		return "", lifecycle.InstalledInfo{}, fmt.Errorf("create temp install root: %w", err)
	}
	defer func() { _ = os.RemoveAll(tmpRoot) }()
	if err = acquireFromURL(ctx, tmpRoot, releaseURL, logger); err != nil {
		return "", lifecycle.InstalledInfo{}, err
	}
	installed, err := lifecycle.DetectInstalled(tmpRoot)
	if err != nil {
		return "", lifecycle.InstalledInfo{}, fmt.Errorf("detect downloaded install: %w", err)
	}
	finalRoot := filepath.Join(installRoot, normalizeVersionDir(installed.Version))
	if err = os.RemoveAll(finalRoot); err != nil && !os.IsNotExist(err) {
		return "", lifecycle.InstalledInfo{}, fmt.Errorf(
			"remove existing install %s: %w",
			finalRoot,
			err,
		)
	}
	if err = os.Rename(tmpRoot, finalRoot); err != nil {
		return "", lifecycle.InstalledInfo{}, fmt.Errorf("move install into place: %w", err)
	}
	return finalRoot, installed, nil
}

func downloadRelease(ctx context.Context, releaseURL string) (zipPath, sum string, err error) {
	var req *http.Request
	req, err = http.NewRequestWithContext(ctx, http.MethodGet, releaseURL, nil)
	if err != nil {
		return "", "", err
	}
	const downloadTimeout = 30 * time.Minute
	client := &http.Client{Timeout: downloadTimeout}
	var resp *http.Response
	resp, err = client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("download: HTTP %d", resp.StatusCode)
	}
	var zipFile *os.File
	zipFile, err = os.CreateTemp("", "foundry-*.zip")
	if err != nil {
		return "", "", err
	}
	h := sha256.New()
	if _, err = io.Copy(io.MultiWriter(zipFile, h), resp.Body); err != nil {
		_ = zipFile.Close()
		_ = os.Remove(zipFile.Name())
		return "", "", err
	}
	if err = zipFile.Close(); err != nil {
		_ = os.Remove(zipFile.Name())
		return "", "", err
	}
	return zipFile.Name(), hex.EncodeToString(h.Sum(nil)), nil
}
