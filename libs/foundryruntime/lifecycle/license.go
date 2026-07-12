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

// licenseFile is the subset of Config/license.json used to key the cache by the
// Foundry version the signature was issued for.
type licenseFile struct {
	Version string `json:"version"`
}

// SyncLicense harvests the license currently in the data path into a per-version
// cache, then seeds the data path with the cached license for targetVersion when
// one exists. This lets a Foundry version switch reuse a license already accepted
// for that version instead of prompting for it again.
//
// The signature in license.json is bound to host+key+version, so a cached file is
// valid only while the host and license key stay the same; a mismatch simply
// makes Foundry prompt again, which is the pre-existing behaviour.
func SyncLicense(dataPath, targetVersion, cacheDir string) error {
	if cacheDir == "" {
		return nil
	}
	if err := harvestLicense(dataPath, cacheDir); err != nil {
		return err
	}
	return seedLicense(dataPath, targetVersion, cacheDir)
}

// harvestLicense copies the data path's current license into cache/<version>.
func harvestLicense(dataPath, cacheDir string) error {
	src := filepath.Join(ConfigDir(dataPath), licenseName)
	data, err := os.ReadFile(src)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("read license %s: %w", src, err)
	}
	var lf licenseFile
	if err = json.Unmarshal(data, &lf); err != nil {
		return fmt.Errorf("parse license %s: %w", src, err)
	}
	dir, ok := versionCacheDir(cacheDir, lf.Version)
	if !ok {
		return nil
	}
	if err = os.MkdirAll(dir, fsperm.Dir); err != nil {
		return fmt.Errorf("mkdir %s: %w", dir, err)
	}
	dest := filepath.Join(dir, licenseName)
	if existing, readErr := os.ReadFile(dest); readErr == nil && bytes.Equal(existing, data) {
		return nil
	}
	if err = os.WriteFile(dest, data, fsperm.Secret); err != nil {
		return fmt.Errorf("write license cache %s: %w", dest, err)
	}
	return nil
}

// seedLicense restores cache/<targetVersion> into the data path when present.
func seedLicense(dataPath, targetVersion, cacheDir string) error {
	dir, ok := versionCacheDir(cacheDir, targetVersion)
	if !ok {
		return nil
	}
	src := filepath.Join(dir, licenseName)
	data, err := os.ReadFile(src)
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
	dest := filepath.Join(confDir, licenseName)
	if existing, readErr := os.ReadFile(dest); readErr == nil && bytes.Equal(existing, data) {
		return nil
	}
	if err = os.WriteFile(dest, data, fsperm.Secret); err != nil {
		return fmt.Errorf("write license %s: %w", dest, err)
	}
	return nil
}

// versionCacheDir builds the cache subdirectory for a version, rejecting values
// that are empty or would escape the cache root via path separators.
func versionCacheDir(cacheDir, version string) (string, bool) {
	version = strings.TrimSpace(version)
	if version == "" || version == "." || version == ".." ||
		version != filepath.Base(version) || strings.ContainsAny(version, `/\`) {
		return "", false
	}
	return filepath.Join(cacheDir, version), true
}
