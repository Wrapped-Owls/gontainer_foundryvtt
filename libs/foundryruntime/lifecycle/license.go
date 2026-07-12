package lifecycle

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrykit/fsperm"
)

const licenseName = "license.json"

type licenseFile struct {
	Version string `json:"version"`
}

func SyncLicense(dataPath, targetVersion, cacheDir string) error {
	if cacheDir == "" {
		return nil
	}
	if err := harvestLicense(dataPath, cacheDir); err != nil {
		return err
	}
	return seedLicense(dataPath, targetVersion, cacheDir) // host+key bound; mismatch re-prompts
}

func harvestLicense(dataPath, cacheDir string) error {
	src := filepath.Join(ConfigDir(dataPath), licenseName)
	license, err := os.ReadFile(src) //nolint:gosec // path derived from operator config
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("read license %s: %w", src, err)
	}
	var lf licenseFile
	if err = json.Unmarshal(license, &lf); err != nil {
		return fmt.Errorf("parse license %s: %w", src, err)
	}
	dir, ok := versionCacheDir(cacheDir, lf.Version)
	if !ok {
		return nil
	}
	if err = os.MkdirAll(dir, fsperm.Dir); err != nil {
		return fmt.Errorf("mkdir %s: %w", dir, err)
	}
	return writeFileIfChanged(filepath.Join(dir, licenseName), license)
}

func seedLicense(dataPath, targetVersion, cacheDir string) error {
	dir, ok := versionCacheDir(cacheDir, targetVersion)
	if !ok {
		return nil
	}
	src := filepath.Join(dir, licenseName)
	license, err := os.ReadFile(src) //nolint:gosec // cache dir traversal-guarded
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("read license cache %s: %w", src, err)
	}
	confDir := ConfigDir(dataPath)
	if err = os.MkdirAll(confDir, fsperm.Dir); err != nil {
		return fmt.Errorf("mkdir %s: %w", confDir, err)
	}
	return writeFileIfChanged(filepath.Join(confDir, licenseName), license)
}

func writeFileIfChanged(dest string, content []byte) error {
	existing, err := os.ReadFile(dest) //nolint:gosec // guarded path
	if err == nil && bytes.Equal(existing, content) {
		return nil
	}
	if err = os.WriteFile(dest, content, fsperm.Secret); err != nil { //nolint:gosec // guarded path
		return fmt.Errorf("write %s: %w", dest, err)
	}
	return nil
}

func versionCacheDir(cacheDir, version string) (string, bool) {
	version = strings.TrimSpace(version)
	if version == "" || version == "." || version == ".." ||
		version != filepath.Base(version) || strings.ContainsAny(version, `/\`) {
		return "", false // reject empty or path-escaping version
	}
	return filepath.Join(cacheDir, version), true
}
