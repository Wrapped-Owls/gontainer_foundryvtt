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

type Kind int

const (
	KindUnknown Kind = iota
	KindLinux
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

var (
	ErrUnknownKind = errors.New("archive: unknown release kind")
	ErrNotZip      = errors.New("archive: not a zip file")
	ErrUnsafePath  = errors.New("archive: zip entry escapes destination")
)

var magicZip = []byte{'P', 'K', 0x03, 0x04}

func IsZip(path string) (bool, error) {
	const zipMagicLen = 4

	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer func() { _ = f.Close() }()
	var head [zipMagicLen]byte
	n, err := io.ReadFull(f, head[:])
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) {
		return false, err
	}
	return n == zipMagicLen && string(head[:]) == string(magicZip), nil
}

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

func Extract(zipPath, baseDir string) (Kind, error) {
	const dirPerm fs.FileMode = 0o755
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
		if strings.Contains(f.Name, "..") {
			return kind, fmt.Errorf("%w: %s", ErrUnsafePath, f.Name)
		}
		rel := filepath.FromSlash(prefix + f.Name)
		dest := filepath.Join(cleanBase, rel)
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
	const filePerm fs.FileMode = 0o644
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
