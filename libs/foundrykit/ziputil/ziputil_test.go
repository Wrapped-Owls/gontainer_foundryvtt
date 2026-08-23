package ziputil

import (
	"archive/zip"
	"bytes"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"testing"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrykit/fsperm"
)

const unsupportedZipMethod = 99

type unsupportedWriteCloser struct{ io.Writer }

func (unsupportedWriteCloser) Close() error { return nil }

var registerUnsupportedCompressor = sync.OnceFunc(func() {
	zip.RegisterCompressor(unsupportedZipMethod, func(w io.Writer) (io.WriteCloser, error) {
		return unsupportedWriteCloser{w}, nil
	})
})

func buildZip(t *testing.T, name, body string, mode fs.FileMode) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	header := &zip.FileHeader{Name: name, Method: zip.Deflate}
	header.SetMode(mode)
	entry, err := zw.CreateHeader(header)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = io.WriteString(entry, body); err != nil {
		t.Fatal(err)
	}
	if err = zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func buildUnsupportedMethodZip(t *testing.T, name, body string) []byte {
	t.Helper()
	registerUnsupportedCompressor()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	entry, err := zw.CreateHeader(&zip.FileHeader{Name: name, Method: unsupportedZipMethod})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = io.WriteString(entry, body); err != nil {
		t.Fatal(err)
	}
	if err = zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func buildBadChecksumZip(t *testing.T, name, body string) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	header := &zip.FileHeader{
		Name:               name,
		Method:             zip.Store,
		CRC32:              0xdeadbeef,
		UncompressedSize64: uint64(len(body)),
		CompressedSize64:   uint64(len(body)),
	}
	entry, err := zw.CreateRaw(header)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = io.WriteString(entry, body); err != nil {
		t.Fatal(err)
	}
	if err = zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func zipEntry(t *testing.T, zipBytes []byte, name string) *zip.File {
	t.Helper()
	zr, err := zip.NewReader(bytes.NewReader(zipBytes), int64(len(zipBytes)))
	if err != nil {
		t.Fatal(err)
	}
	for _, candidate := range zr.File {
		if candidate.Name == name {
			return candidate
		}
	}
	t.Fatalf("entry %q not found in fixture", name)
	return nil
}

func TestWriteEntry(t *testing.T) {
	oldUmask := syscall.Umask(0)
	defer syscall.Umask(oldUmask)

	wantFallbackPerm := fsperm.File

	testCases := []struct {
		name      string
		zipBytes  []byte
		entryName string
		destPath  func(t *testing.T) string
		wantErr   error
		wantPerm  *fs.FileMode
		wantData  string
	}{
		{
			name:      "extracts file content and preserves mode",
			zipBytes:  buildZip(t, "file.txt", "hello ziputil", 0o640),
			entryName: "file.txt",
			destPath: func(t *testing.T) string {
				return filepath.Join(t.TempDir(), "out.txt")
			},
			wantData: "hello ziputil",
			wantPerm: func() *fs.FileMode { m := fs.FileMode(0o640); return &m }(),
		},
		{
			name:      "zero mode falls back to fsperm.File",
			zipBytes:  buildZip(t, "file.txt", "fallback content", 0),
			entryName: "file.txt",
			destPath: func(t *testing.T) string {
				return filepath.Join(t.TempDir(), "out.txt")
			},
			wantData: "fallback content",
			wantPerm: &wantFallbackPerm,
		},
		{
			name:      "entry open error on unsupported compression method",
			zipBytes:  buildUnsupportedMethodZip(t, "bad.bin", "irrelevant"),
			entryName: "bad.bin",
			destPath: func(t *testing.T) string {
				return filepath.Join(t.TempDir(), "out.bin")
			},
			wantErr: zip.ErrAlgorithm,
		},
		{
			name:      "copy error on checksum mismatch",
			zipBytes:  buildBadChecksumZip(t, "corrupt.bin", "corrupt content"),
			entryName: "corrupt.bin",
			destPath: func(t *testing.T) string {
				return filepath.Join(t.TempDir(), "out.bin")
			},
			wantErr: zip.ErrChecksum,
		},
		{
			name:      "destination open error on missing directory",
			zipBytes:  buildZip(t, "file.txt", "content", 0o640),
			entryName: "file.txt",
			destPath: func(t *testing.T) string {
				return filepath.Join(t.TempDir(), "missing", "out.txt")
			},
			wantErr: fs.ErrNotExist,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			entry := zipEntry(t, testCase.zipBytes, testCase.entryName)
			dest := testCase.destPath(t)

			err := WriteEntry(entry, dest)
			if !errors.Is(err, testCase.wantErr) {
				t.Fatalf("err = %v, want %v", err, testCase.wantErr)
			}
			if testCase.wantErr != nil {
				return
			}

			got, err := os.ReadFile(dest)
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != testCase.wantData {
				t.Fatalf("data = %q, want %q", got, testCase.wantData)
			}

			if testCase.wantPerm != nil {
				stat, err := os.Stat(dest)
				if err != nil {
					t.Fatal(err)
				}
				if stat.Mode().Perm() != *testCase.wantPerm {
					t.Fatalf("perm = %v, want %v", stat.Mode().Perm(), *testCase.wantPerm)
				}
			}
		})
	}
}

func TestOpen(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		path    func(t *testing.T) string
		wantErr error
	}{
		{
			name: "opens an existing archive",
			path: func(t *testing.T) string {
				path := filepath.Join(t.TempDir(), "archive.zip")
				if err := os.WriteFile(path, buildZip(t, "a.txt", "content", 0o640), 0o644); err != nil {
					t.Fatal(err)
				}
				return path
			},
		},
		{
			name: "missing archive",
			path: func(t *testing.T) string {
				return filepath.Join(t.TempDir(), "missing.zip")
			},
			wantErr: fs.ErrNotExist,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rc, err := Open(testCase.path(t))
			if !errors.Is(err, testCase.wantErr) {
				t.Fatalf("err = %v, want %v", err, testCase.wantErr)
			}
			if rc != nil {
				defer func() { _ = rc.Close() }()
			}
		})
	}
}
