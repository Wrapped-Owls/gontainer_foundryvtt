// Package lifecycle holds the install/upgrade decision tree and the
// orchestration helpers that the apps/foundryctl PID 1 binary stitches
// together at run time.
//
// Responsibilities are intentionally split into small, side-effect-free
// helpers so they can be table-tested without fakes:
//
//   - DetectInstalled     — inspect a Foundry install root, report the
//     version present (if any) by reading resources/app/package.json.
//   - DecideInstall       — given a desired version + current state,
//     return what action the controller must take.
//   - WriteOptions        — render config.Options to <dataPath>/Config/
//     options.json, idempotently (no write if bytes unchanged).
//   - WriteAdminPassword  — render hashed admin.txt (or remove it when
//     no admin key is configured).
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

const (
	dirPerm    fs.FileMode = 0o755
	filePerm   fs.FileMode = 0o644
	secretPerm fs.FileMode = 0o600
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
