package action

import (
	"archive/zip"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrypatch/internal/testzip"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrypatch/manifest"
)

func TestZipOverlayRunner_ExtractsFiles(t *testing.T) {
	overlay := testzip.MakeZip(t, map[string]string{
		"subdir/patch.txt": "patched",
	})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(overlay)
	}))
	defer srv.Close()

	dest := t.TempDir()
	act := manifest.Action{
		Type:   manifest.ActionZipOverlay,
		URL:    srv.URL,
		SHA256: bodySum(overlay),
	}
	if err := ZipOverlay(http.DefaultClient).Run(context.Background(), act, dest); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Join(dest, "subdir", "patch.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "patched" {
		t.Errorf("got %q, want %q", got, "patched")
	}
}

func TestZipOverlayRunner_HashMismatch(t *testing.T) {
	overlay := testzip.MakeZip(t, map[string]string{"f.txt": "x"})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(overlay)
	}))
	defer srv.Close()

	act := manifest.Action{Type: manifest.ActionZipOverlay, URL: srv.URL, SHA256: "bad"}
	if err := ZipOverlay(
		http.DefaultClient,
	).Run(context.Background(), act, t.TempDir()); err == nil {
		t.Fatal("expected hash mismatch error")
	}
}

func TestZipOverlayRunner_RejectsMalformedArchive(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name             string
		archive          func(t *testing.T) []byte
		wantErr          error
		wantErrSubstring string
	}{
		{
			name: "zip slip entry is refused",
			archive: func(t *testing.T) []byte {
				return testzip.MakeZip(t, map[string]string{"../../etc/passwd": "pwned"})
			},
			wantErrSubstring: "escapes dest",
		},
		{
			name: "corrupt archive fails to open",
			archive: func(t *testing.T) []byte {
				return []byte("not a zip archive")
			},
			wantErr: zip.ErrFormat,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			overlay := testCase.archive(t)
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				_, _ = w.Write(overlay)
			}))
			defer srv.Close()

			act := manifest.Action{
				Type:   manifest.ActionZipOverlay,
				URL:    srv.URL,
				SHA256: bodySum(overlay),
			}
			err := ZipOverlay(http.DefaultClient).Run(context.Background(), act, t.TempDir())
			if err == nil {
				t.Fatal("expected error")
			}
			if testCase.wantErr != nil && !errors.Is(err, testCase.wantErr) {
				t.Fatalf("err = %v, want %v", err, testCase.wantErr)
			}
			if testCase.wantErrSubstring != "" && !strings.Contains(err.Error(), testCase.wantErrSubstring) {
				t.Fatalf("err = %v, want substring %q", err, testCase.wantErrSubstring)
			}
		})
	}
}
