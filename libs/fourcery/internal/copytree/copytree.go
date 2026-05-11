package copytree

import (
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

// ErrUnsafeLink is returned when a symlink target escapes the source
// root or contains "..".
var ErrUnsafeLink = errors.New("copytree: symlink escapes source root")

// Copy recursively copies the contents of src into dst. dst must
// already exist. Regular files are copied with their existing mode;
// directories are created with 0o755; symlinks are preserved only when
// their target stays inside src.
func Copy(src, dst string) error {
	cleanSrc, err := filepath.Abs(src)
	if err != nil {
		return fmt.Errorf("copytree: abs src: %w", err)
	}
	cleanDst, err := filepath.Abs(dst)
	if err != nil {
		return fmt.Errorf("copytree: abs dst: %w", err)
	}
	return filepath.WalkDir(cleanSrc, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, rerr := filepath.Rel(cleanSrc, path)
		if rerr != nil {
			return rerr
		}
		if rel == "." {
			return nil
		}
		target := filepath.Join(cleanDst, rel)
		info, ierr := d.Info()
		if ierr != nil {
			return ierr
		}
		switch {
		case d.IsDir():
			return os.MkdirAll(target, dirPerm)
		case info.Mode()&fs.ModeSymlink != 0:
			return copySymlink(path, target, cleanSrc)
		case info.Mode().IsRegular():
			return copyFile(path, target, info.Mode().Perm())
		default:
			return fmt.Errorf("copytree: unsupported file %s (mode %v)", path, info.Mode())
		}
	})
}

func copyFile(src, dst string, mode fs.FileMode) error {
	if mode == 0 {
		mode = filePerm
	}
	if err := os.MkdirAll(filepath.Dir(dst), dirPerm); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_EXCL|os.O_WRONLY, mode)
	if err != nil {
		return err
	}
	if _, err = io.Copy(out, in); err != nil {
		_ = out.Close()
		_ = os.Remove(dst)
		return err
	}
	return out.Close()
}

func copySymlink(src, dst, srcRoot string) error {
	tgt, err := os.Readlink(src)
	if err != nil {
		return err
	}
	if strings.Contains(tgt, "..") {
		return fmt.Errorf("%w: %s -> %s", ErrUnsafeLink, src, tgt)
	}
	resolved := tgt
	if !filepath.IsAbs(resolved) {
		resolved = filepath.Join(filepath.Dir(src), resolved)
	}
	resolvedAbs, err := filepath.Abs(resolved)
	if err != nil {
		return err
	}
	if !strings.HasPrefix(
		resolvedAbs+string(os.PathSeparator),
		srcRoot+string(os.PathSeparator),
	) && resolvedAbs != srcRoot {
		return fmt.Errorf("%w: %s -> %s", ErrUnsafeLink, src, tgt)
	}
	if err := os.MkdirAll(filepath.Dir(dst), dirPerm); err != nil {
		return err
	}
	return os.Symlink(tgt, dst)
}
