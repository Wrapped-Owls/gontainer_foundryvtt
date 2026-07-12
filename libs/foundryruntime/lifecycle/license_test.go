package lifecycle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeLicense(t *testing.T, dataPath, version string) {
	t.Helper()
	dir := ConfigDir(dataPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	body := `{"host":"h","license":"KEY","version":"` + version + `","signature":"sig-` + version + `"}`
	if err := os.WriteFile(filepath.Join(dir, licenseName), []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
}

func readLicense(t *testing.T, dataPath string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(ConfigDir(dataPath), licenseName))
	if err != nil {
		t.Fatalf("read license: %v", err)
	}
	return string(data)
}

func TestSyncLicense_restoresCachedVersionOnSwitch(t *testing.T) {
	data := t.TempDir()
	cache := t.TempDir()

	// Accept v13, harvest it into the cache.
	writeLicense(t, data, "13")
	if err := SyncLicense(data, "13", cache); err != nil {
		t.Fatalf("sync v13: %v", err)
	}
	// Switch to v14 in the same data path: v13 is harvested, no cached v14 yet.
	writeLicense(t, data, "14")
	if err := SyncLicense(data, "14", cache); err != nil {
		t.Fatalf("sync v14: %v", err)
	}
	// Switch back to v13: the cached v13 license must be restored.
	if err := SyncLicense(data, "13", cache); err != nil {
		t.Fatalf("sync back to v13: %v", err)
	}
	if got := readLicense(t, data); !strings.Contains(got, `"version":"13"`) {
		t.Errorf("expected restored v13 license, got %q", got)
	}
}

func TestSyncLicense_noCacheDirIsNoop(t *testing.T) {
	data := t.TempDir()
	writeLicense(t, data, "13")
	if err := SyncLicense(data, "13", ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSyncLicense_missingLicenseIsNoop(t *testing.T) {
	data := t.TempDir()
	cache := t.TempDir()
	if err := SyncLicense(data, "13", cache); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := os.Stat(filepath.Join(ConfigDir(data), licenseName)); !os.IsNotExist(err) {
		t.Errorf("expected no license written, err=%v", err)
	}
}

func TestVersionCacheDir_rejectsTraversal(t *testing.T) {
	testCases := []struct {
		name    string
		version string
		wantOK  bool
	}{
		{"plain", "14.361.0", true},
		{"empty", "", false},
		{"dotdot", "..", false},
		{"slash", "../etc", false},
		{"backslash", `..\etc`, false},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if _, ok := versionCacheDir("/cache", testCase.version); ok != testCase.wantOK {
				t.Errorf(
					"versionCacheDir(%q) ok=%v, want %v",
					testCase.version,
					ok,
					testCase.wantOK,
				)
			}
		})
	}
}
