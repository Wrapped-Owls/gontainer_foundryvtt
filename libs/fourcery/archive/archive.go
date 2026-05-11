// Package archive detects the kind of FoundryVTT release archive (Linux
// vs Node.js) and extracts it to disk using stdlib archive/zip and
// content sniffing — no external tools required.
package archive

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const (
	dirPerm  fs.FileMode = 0o755
	filePerm fs.FileMode = 0o644
)

// Kind classifies a Foundry release archive.
type Kind int

const (
	// KindUnknown is returned for anything we cannot positively identify.
	KindUnknown Kind = iota
	// KindLinux is the desktop Linux release (Electron). Contains the
	// resources/ tree at the top level and main.mjs at
	// resources/app/main.mjs.
	KindLinux
	// KindNode is the headless Node.js release. The archive contents are
	// the `app` tree itself; main.mjs lives at the archive root.
	KindNode
)

func (k Kind) String() string {
	switch k {
	case KindLinux:
		return "linux"
	case KindNode:
		return "node"
	default:
		return "unknown"
	}
}

// Errors returned by this package.
var (
	ErrUnknownKind = errors.New("archive: unknown release kind")
	ErrNotZip      = errors.New("archive: not a zip file")
	ErrUnsafePath  = errors.New("archive: zip entry escapes destination")
)

// magicZip is the standard PKZip header.
var magicZip = []byte{'P', 'K', 0x03, 0x04}

// IsZip returns true if path is recognisably a zip archive.
func IsZip(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer func() { _ = f.Close() }()
	var head [4]byte
	n, err := io.ReadFull(f, head[:])
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) {
		return false, err
	}
	return n == 4 && string(head[:]) == string(magicZip), nil
}

// Detect classifies the zip at path. It does NOT extract; it only reads
// the central directory.
func Detect(path string) (Kind, error) {
	zr, err := zip.OpenReader(path)
	if err != nil {
		return KindUnknown, fmt.Errorf("%w: %v", ErrNotZip, err)
	}
	defer func() { _ = zr.Close() }()
	hasLinux, hasNode := false, false
	for _, f := range zr.File {
		switch f.Name {
		case "resources/app/main.mjs":
			hasLinux = true
		case "main.mjs":
			hasNode = true
		}
		if hasLinux {
			break
		}
	}
	switch {
	case hasLinux:
		return KindLinux, nil
	case hasNode:
		return KindNode, nil
	default:
		return KindUnknown, ErrUnknownKind
	}
}

// Extract unpacks the release at zipPath into baseDir, normalising both
// archive layouts so the result is always:
//
//	<baseDir>/resources/app/main.mjs
//	<baseDir>/resources/app/...
//
// For a Linux release the "resources/" prefix is preserved verbatim.
// For a Node release the entries are placed under "resources/app/"
// transparently.
func Extract(zipPath, baseDir string) (Kind, error) {
	kind, err := Detect(zipPath)
	if err != nil {
		return kind, err
	}
	zr, err := zip.OpenReader(zipPath)
	if err != nil {
		return kind, err
	}
	defer func() { _ = zr.Close() }()

	prefix := ""
	if kind == KindNode {
		prefix = "resources/app/"
	}

	cleanBase, err := filepath.Abs(baseDir)
	if err != nil {
		return kind, err
	}

	for _, f := range zr.File {
		// Reject path components that would escape the destination before
		// filepath.Join collapses them.
		if strings.Contains(f.Name, "..") {
			return kind, fmt.Errorf("%w: %s", ErrUnsafePath, f.Name)
		}
		rel := filepath.FromSlash(prefix + f.Name)
		dest := filepath.Join(cleanBase, rel)
		// Belt-and-braces: make sure the resolved path stays under base.
		if !strings.HasPrefix(dest+string(os.PathSeparator), cleanBase+string(os.PathSeparator)) &&
			dest != cleanBase {
			return kind, fmt.Errorf("%w: %s", ErrUnsafePath, f.Name)
		}
		if f.FileInfo().IsDir() {
			if err = os.MkdirAll(dest, dirPerm); err != nil {
				return kind, err
			}
			continue
		}
		if err = os.MkdirAll(filepath.Dir(dest), dirPerm); err != nil {
			return kind, err
		}
		if err = writeZipEntry(f, dest); err != nil {
			return kind, err
		}
	}
	return kind, nil
}

func writeZipEntry(f *zip.File, dest string) error {
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer func() { _ = rc.Close() }()
	mode := f.Mode().Perm()
	if mode == 0 {
		mode = filePerm
	}
	out, err := os.OpenFile(dest, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode)
	if err != nil {
		return err
	}
	if _, err = io.Copy(out, rc); err != nil { //nolint:gosec // size bounded by zip metadata
		_ = out.Close()
		return err
	}
	return out.Close()
}
