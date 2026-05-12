package archive

import (
	"archive/zip"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrykit/fsperm"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrykit/ziputil"
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
		if strings.Contains(f.Name, "..") { // reject before filepath.Join collapses ".."
			return kind, fmt.Errorf("%w: %s", ErrUnsafePath, f.Name)
		}
		rel := filepath.FromSlash(prefix + f.Name)
		dest := filepath.Join(cleanBase, rel)
		if !strings.HasPrefix(dest+string(os.PathSeparator), cleanBase+string(os.PathSeparator)) &&
			dest != cleanBase {
			return kind, fmt.Errorf("%w: %s", ErrUnsafePath, f.Name)
		}
		if f.FileInfo().IsDir() {
			if err = os.MkdirAll(dest, fsperm.Dir); err != nil {
				return kind, err
			}
			continue
		}
		if err = os.MkdirAll(filepath.Dir(dest), fsperm.Dir); err != nil {
			return kind, err
		}
		if err = ziputil.WriteEntry(f, dest); err != nil {
			return kind, err
		}
	}
	return kind, nil
}
