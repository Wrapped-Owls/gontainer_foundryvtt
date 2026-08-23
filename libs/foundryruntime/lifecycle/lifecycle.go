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

// InstalledInfo describes whatever Foundry release is currently present
// at the install root. Both fields zero when nothing is installed.
type InstalledInfo struct {
	// Present is true iff resources/app/main.mjs exists.
	Present bool
	// Version is the value read from resources/app/package.json.
	// Empty when Present == false or package.json is unreadable.
	Version string
}

// pkgJSON is the minimal subset of package.json we read.
type pkgJSON struct {
	Version string `json:"version"`
}

// DetectInstalled inspects installRoot. The function is read-only and
// returns no error for a missing install (Present=false instead).
// Errors are returned only for unexpected I/O or malformed JSON.
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
