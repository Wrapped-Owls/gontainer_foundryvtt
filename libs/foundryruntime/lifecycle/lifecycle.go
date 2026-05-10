package lifecycle

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type InstalledInfo struct {
	Present bool
	Version string
}

type pkgJSON struct {
	Version string `json:"version"`
}

func DetectInstalled(installRoot string) (InstalledInfo, error) {
	mainPath := filepath.Join(installRoot, "resources", "app", "main.mjs")
	if _, err := os.Stat(mainPath); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return InstalledInfo{}, nil
		}
		return InstalledInfo{}, fmt.Errorf("lifecycle: stat %s: %w", mainPath, err)
	}
	pkg := filepath.Join(installRoot, "resources", "app", "package.json")
	b, err := os.ReadFile(pkg)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return InstalledInfo{Present: true}, nil
		}
		return InstalledInfo{Present: true}, fmt.Errorf("lifecycle: read %s: %w", pkg, err)
	}
	var p pkgJSON
	if err = json.Unmarshal(b, &p); err != nil {
		return InstalledInfo{Present: true}, fmt.Errorf("lifecycle: parse %s: %w", pkg, err)
	}
	return InstalledInfo{Present: true, Version: strings.TrimSpace(p.Version)}, nil
}
