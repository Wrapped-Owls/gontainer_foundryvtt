package probe

import (
	"archive/zip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// ErrNoVersion is returned when neither the filename nor an embedded
// package.json reveals a version.
var ErrNoVersion = errors.New("probe: no version found")

// pkgJSON is the minimal subset of package.json we read; mirrors the
// shape consumed by libs/foundryruntime/lifecycle.
type pkgJSON struct {
	Version string `json:"version"`
}

// filenamePattern accepts foundryvtt-v14.361.2.zip / foundryvtt_v14.361
// / foundryvtt-14.361.0 (with or without trailing .zip), case-insensitive.
var filenamePattern = regexp.MustCompile(
	`(?i)^foundryvtt[_\-]?v?(\d+\.\d+(?:\.\d+)?)(?:\.zip)?$`,
)

// Filename returns the version embedded in name (a single path
// component, not a path). Returns ErrNoVersion when the name does not
// match the canonical pattern.
func Filename(name string) (string, error) {
	base := filepath.Base(strings.TrimSpace(name))
	m := filenamePattern.FindStringSubmatch(base)
	if len(m) < 2 {
		return "", ErrNoVersion
	}
	return m[1], nil
}

// Folder reads <root>/resources/app/package.json and returns its
// version. Returns ErrNoVersion when the file is absent or the version
// field is empty.
func Folder(root string) (string, error) {
	path := filepath.Join(root, "resources", "app", "package.json")
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return "", ErrNoVersion
		}
		return "", fmt.Errorf("probe: read %s: %w", path, err)
	}
	v, err := parseVersion(b)
	if err != nil {
		return "", fmt.Errorf("probe: parse %s: %w", path, err)
	}
	if v == "" {
		return "", ErrNoVersion
	}
	return v, nil
}

// Zip opens path, locates a package.json entry, and returns its
// version. Linux-layout archives ship the file at
// resources/app/package.json; Node-layout archives ship it at the root
// as package.json. Returns ErrNoVersion when neither entry is present
// or the version field is empty.
func Zip(path string) (string, error) {
	zr, err := zip.OpenReader(path)
	if err != nil {
		return "", fmt.Errorf("probe: open zip %s: %w", path, err)
	}
	defer func() { _ = zr.Close() }()
	var entry *zip.File
	for _, f := range zr.File {
		switch f.Name {
		case "resources/app/package.json", "package.json":
			entry = f
		}
		if entry != nil {
			break
		}
	}
	if entry == nil {
		return "", ErrNoVersion
	}
	rc, err := entry.Open()
	if err != nil {
		return "", fmt.Errorf("probe: open zip entry %s: %w", entry.Name, err)
	}
	defer func() { _ = rc.Close() }()
	b, err := io.ReadAll(rc)
	if err != nil {
		return "", fmt.Errorf("probe: read zip entry %s: %w", entry.Name, err)
	}
	v, err := parseVersion(b)
	if err != nil {
		return "", fmt.Errorf("probe: parse zip entry %s: %w", entry.Name, err)
	}
	if v == "" {
		return "", ErrNoVersion
	}
	return v, nil
}

func parseVersion(b []byte) (string, error) {
	var p pkgJSON
	if err := json.Unmarshal(b, &p); err != nil {
		return "", err
	}
	return strings.TrimSpace(p.Version), nil
}
