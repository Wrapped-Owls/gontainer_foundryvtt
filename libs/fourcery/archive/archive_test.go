package archive

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/fourcery/internal/testzip"
)

func TestDetectLinuxRelease(t *testing.T) {
	zp := testzip.MakeZip(t, map[string]string{
		"resources/app/main.mjs":     "// linux entry",
		"resources/app/package.json": `{"release":{"generation":14,"build":361}}`,
	})
	k, err := Detect(zp)
	if err != nil || k != KindLinux {
		t.Fatalf("got kind=%v err=%v", k, err)
	}
}

func TestDetectNodeRelease(t *testing.T) {
	zp := testzip.MakeZip(t, map[string]string{
		"main.mjs":     "// node entry",
		"package.json": `{"release":{"generation":14,"build":361}}`,
	})
	k, err := Detect(zp)
	if err != nil || k != KindNode {
		t.Fatalf("got kind=%v err=%v", k, err)
	}
}

func TestDetectUnknown(t *testing.T) {
	zp := testzip.MakeZip(t, map[string]string{"random.txt": "x"})
	if _, err := Detect(zp); !errors.Is(err, ErrUnknownKind) {
		t.Fatalf("expected ErrUnknownKind, got %v", err)
	}
}

func TestExtractNormalisesNodeRelease(t *testing.T) {
	zp := testzip.MakeZip(t, map[string]string{
		"main.mjs":     "console.log('hi')",
		"public/x.css": "body{}",
	})
	dest := t.TempDir()
	k, err := Extract(zp, dest)
	if err != nil || k != KindNode {
		t.Fatalf("k=%v err=%v", k, err)
	}
	main := filepath.Join(dest, "resources", "app", "main.mjs")
	body, err := os.ReadFile(main)
	if err != nil {
		t.Fatalf("missing main.mjs: %v", err)
	}
	if !bytes.Contains(body, []byte("console.log")) {
		t.Errorf("wrong content")
	}
}

func TestExtractPreservesLinuxLayout(t *testing.T) {
	zp := testzip.MakeZip(t, map[string]string{
		"resources/app/main.mjs": "linux",
		"resources/lib/foo.so":   "elf",
	})
	dest := t.TempDir()
	if _, err := Extract(zp, dest); err != nil {
		t.Fatal(err)
	}
	for _, p := range []string{"resources/app/main.mjs", "resources/lib/foo.so"} {
		if _, err := os.Stat(filepath.Join(dest, filepath.FromSlash(p))); err != nil {
			t.Errorf("missing %s: %v", p, err)
		}
	}
}

func TestExtractRejectsZipSlip(t *testing.T) {
	zp := testzip.MakeZip(t, map[string]string{
		"main.mjs":         "ok", // makes it KindNode
		"../../etc/passwd": "bad",
	})
	dest := t.TempDir()
	if _, err := Extract(zp, dest); !errors.Is(err, ErrUnsafePath) {
		// Note: the prefix-prepending in Node mode would actually produce
		// resources/app/../../etc/passwd, which still escapes baseDir.
		if err == nil || !strings.Contains(err.Error(), "escapes") {
			t.Fatalf("expected unsafe-path error, got %v", err)
		}
	}
}
