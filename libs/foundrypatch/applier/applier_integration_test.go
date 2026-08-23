//go:build integration

package applier

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrypatch/ledger"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrypatch/manifest"
)

func TestApplier_FileReplace(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	patches := []manifest.Patch{{
		ID:       "core-fix",
		Versions: ">=1.0.0",
		Actions: []manifest.Action{{
			Type:    manifest.ActionFileReplace,
			Dest:    "resources/app/target.txt",
			Content: "patched",
		}},
	}}

	applier := &Applier{Root: root}
	if applyErr := applier.Apply(t.Context(), patches, nil); applyErr != nil {
		t.Fatalf("apply: %v", applyErr)
	}

	targetBody, readErr := os.ReadFile(filepath.Join(root, "resources/app/target.txt"))
	if readErr != nil {
		t.Fatalf("read target: %v", readErr)
	}
	if string(targetBody) != "patched" {
		t.Fatalf("target content = %q, want %q", targetBody, "patched")
	}
}

func TestApplier_RefusesEscapingTarget(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		dest string
	}{
		{name: "dot-dot escape", dest: "../../etc/passwd"},
		{name: "absolute path", dest: "/etc/passwd"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			applier := &Applier{Root: root}
			patches := []manifest.Patch{{
				ID:       "escape-attempt",
				Versions: ">=1.0.0",
				Actions: []manifest.Action{{
					Type:    manifest.ActionFileReplace,
					Dest:    testCase.dest,
					Content: "pwned",
				}},
			}}

			if applyErr := applier.Apply(t.Context(), patches, nil); applyErr == nil {
				t.Fatal("want an error refusing the escaping destination")
			}

			dirEntries, readDirErr := os.ReadDir(root)
			if readDirErr != nil {
				t.Fatalf("read root: %v", readDirErr)
			}
			if len(dirEntries) != 0 {
				t.Fatalf("root has unexpected entries after rejected patch: %v", dirEntries)
			}
		})
	}
}

func TestApplier_LedgerPreventsDoubleApply(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	patchLedger := &ledger.Ledger{}

	var appliedCount int
	applier := &Applier{
		Root:   root,
		Ledger: patchLedger,
		OnApplied: func(entry ledger.Entry) {
			appliedCount++
			patchLedger.Upsert(entry)
		},
	}

	patches := []manifest.Patch{{
		ID:       "idempotent-fix",
		Versions: ">=1.0.0",
		Actions: []manifest.Action{{
			Type:    manifest.ActionFileReplace,
			Dest:    "marker.txt",
			Content: "v1",
		}},
	}}

	if applyErr := applier.Apply(t.Context(), patches, nil); applyErr != nil {
		t.Fatalf("first apply: %v", applyErr)
	}
	if appliedCount != 1 {
		t.Fatalf("applied count after first apply = %d, want 1", appliedCount)
	}

	if applyErr := applier.Apply(t.Context(), patches, nil); applyErr != nil {
		t.Fatalf("second apply: %v", applyErr)
	}
	if appliedCount != 1 {
		t.Fatalf("applied count after second apply = %d, want 1 (no double-apply)", appliedCount)
	}

	markerBody, readErr := os.ReadFile(filepath.Join(root, "marker.txt"))
	if readErr != nil {
		t.Fatalf("read marker: %v", readErr)
	}
	if string(markerBody) != "v1" {
		t.Fatalf("marker content = %q, want %q", markerBody, "v1")
	}
}
