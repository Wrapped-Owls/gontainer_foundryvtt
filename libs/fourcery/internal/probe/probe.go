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

var ErrNoVersion = errors.New("probe: no version found")

type pkgJSON struct { // mirrors the shape consumed by libs/foundryruntime/lifecycle
	Version string `json:"version"`
}

var filenamePattern = regexp.MustCompile(
	`(?i)^foundryvtt[_\-]?v?(\d+\.\d+(?:\.\d+)?)(?:\.zip)?$`,
)

func Filename(name string) (string, error) {
	base := filepath.Base(strings.TrimSpace(name))
	m := filenamePattern.FindStringSubmatch(base)
	if len(m) < 2 {
		return "", ErrNoVersion
	}
	return m[1], nil
}

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
