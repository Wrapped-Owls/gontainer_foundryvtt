//go:build integration

package activate

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/profile"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundryruntime/jsruntime"
)

const (
	envInstallRoot   = "FOUNDRY_INSTALL_ROOT"
	envDataPath      = "FOUNDRY_DATA_PATH"
	envSourcesDir    = "FOUNDRY_SOURCES_DIR"
	envFoundryVer    = "FOUNDRY_VERSION"
	envJSRuntimeKind = "FOUNDRY_JS_RUNTIME"
	envPatchManifest = "FOUNDRY_PATCH_MANIFEST"
	envLicenseCache  = "FOUNDRY_LICENSE_CACHE"
	envProfilesFile  = "FOUNDRY_PROFILES_FILE"
	envSecretFile    = "FOUNDRY_SECRET_FILE"
	envConfFile      = "CONF_FILE"

	foundryVersion14 = "14.361.2"
	foundryVersion13 = "13.351.0"

	binBun    = "bun"
	binNode22 = "node22"
	binNode24 = "node24"
)

func writeStubBinary(t *testing.T, path string) {
	t.Helper()
	if err := os.WriteFile(path, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatal(err)
	}
}

func writeFoundryCandidate(t *testing.T, dir, version string) {
	t.Helper()
	appDir := filepath.Join(dir, "resources", "app")
	if err := os.MkdirAll(appDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(appDir, "main.mjs"), []byte("//"), 0o644); err != nil {
		t.Fatal(err)
	}
	pkg, err := json.Marshal(struct {
		Version string `json:"version"`
	}{Version: version})
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(appDir, "package.json"), pkg, 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestPrepare_NoNetwork(t *testing.T) {
	root := t.TempDir()
	binDir := t.TempDir()
	writeStubBinary(t, filepath.Join(binDir, binBun))
	writeStubBinary(t, filepath.Join(binDir, binNode22))
	writeStubBinary(t, filepath.Join(binDir, binNode24))

	candidate14 := filepath.Join(root, "foundryvtt_v"+foundryVersion14)
	writeFoundryCandidate(t, candidate14, foundryVersion14)

	dataPath := filepath.Join(root, "data")

	t.Setenv("PATH", binDir)
	t.Setenv(envInstallRoot, root)
	t.Setenv(envDataPath, dataPath)
	t.Setenv(envSourcesDir, filepath.Join(root, "sources"))
	t.Setenv(envFoundryVer, foundryVersion14)
	t.Setenv(envJSRuntimeKind, string(jsruntime.Node))
	t.Setenv(envPatchManifest, filepath.Join(root, "manifest.yaml"))
	t.Setenv(envLicenseCache, filepath.Join(root, "licenses"))
	t.Setenv(envProfilesFile, filepath.Join(root, "profiles.json"))
	t.Setenv(envSecretFile, filepath.Join(root, "secrets.json"))
	t.Setenv(envConfFile, filepath.Join(root, "foundryctl.json"))

	state, err := Prepare(t.Context(), slog.New(slog.DiscardHandler))
	if err != nil {
		t.Fatalf("prepare: %v", err)
	}

	if state.Install.Root != candidate14 {
		t.Fatalf("install root = %q, want %q", state.Install.Root, candidate14)
	}
	if state.Install.Version.String() != foundryVersion14 {
		t.Fatalf("install version = %q, want %q", state.Install.Version.String(), foundryVersion14)
	}
	if state.App.Paths.DataPath != dataPath {
		t.Fatalf("data path = %q, want %q", state.App.Paths.DataPath, dataPath)
	}
	wantRuntimePath14 := filepath.Join(binDir, binNode24)
	if state.JSRuntime.Kind != jsruntime.Node || state.JSRuntime.Path != wantRuntimePath14 {
		t.Fatalf("js runtime = %+v, want node at %q", state.JSRuntime, wantRuntimePath14)
	}

	candidate13 := filepath.Join(root, "foundryvtt_v"+foundryVersion13)
	writeFoundryCandidate(t, candidate13, foundryVersion13)

	profileState, err := PrepareProfile(
		t.Context(), slog.New(slog.DiscardHandler), state, profile.Profile{Version: foundryVersion13},
	)
	if err != nil {
		t.Fatalf("prepare profile: %v", err)
	}
	if profileState.Install.Root != candidate13 {
		t.Fatalf("profile install root = %q, want %q", profileState.Install.Root, candidate13)
	}
	if profileState.Install.Version.String() != foundryVersion13 {
		t.Fatalf(
			"profile install version = %q, want %q",
			profileState.Install.Version.String(), foundryVersion13,
		)
	}
	wantRuntimePath13 := filepath.Join(binDir, binNode22)
	if profileState.JSRuntime.Kind != jsruntime.Node || profileState.JSRuntime.Path != wantRuntimePath13 {
		t.Fatalf("profile js runtime = %+v, want node at %q", profileState.JSRuntime, wantRuntimePath13)
	}
}
