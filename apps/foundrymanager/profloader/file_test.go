package profloader

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/profile"
	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrykit/fsperm"
)

func TestFromFile_notExist(t *testing.T) {
	t.Parallel()

	profiles, active, err := FromFile("/nonexistent/profiles.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if profiles != nil {
		t.Fatalf("expected nil, got %v", profiles)
	}
	if active != "" {
		t.Fatalf("expected empty active, got %q", active)
	}
}

func TestFromFile_valid(t *testing.T) {
	t.Parallel()

	encoded, _ := json.Marshal(map[string]any{
		"active": "alice",
		"profiles": []map[string]any{
			{"name": "alice", "label": "Alice", "dataPath": "/data/alice"},
		},
	})
	path := filepath.Join(t.TempDir(), "profiles.json")
	if err := os.WriteFile(path, encoded, 0o600); err != nil {
		t.Fatal(err)
	}

	profiles, active, err := FromFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(profiles) != 1 {
		t.Fatalf("expected 1 profile, got %d", len(profiles))
	}
	if profiles[0].Name != "alice" {
		t.Errorf("expected alice, got %q", profiles[0].Name)
	}
	if active != "alice" {
		t.Errorf("expected active alice, got %q", active)
	}
}

func TestFromFile_malformed(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "bad.json")
	if err := os.WriteFile(path, []byte("{not json}"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := FromFile(path); err == nil {
		t.Error("expected error for malformed JSON")
	}
}

func TestWriteActive_createAndUpdate(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "profiles.json")

	if writeErr := WriteActive(path, "bob"); writeErr != nil {
		t.Fatalf("unexpected error: %v", writeErr)
	}
	_, active, err := FromFile(path)
	if err != nil {
		t.Fatalf("unexpected error reading back: %v", err)
	}
	if active != "bob" {
		t.Errorf("expected bob, got %q", active)
	}

	if writeErr := WriteActive(path, "alice"); writeErr != nil {
		t.Fatalf("unexpected error on update: %v", writeErr)
	}
	_, active, err = FromFile(path)
	if err != nil {
		t.Fatalf("unexpected error reading back: %v", err)
	}
	if active != "alice" {
		t.Errorf("expected alice after update, got %q", active)
	}
}

func TestWriteActive_preservesProfiles(t *testing.T) {
	t.Parallel()

	encoded, _ := json.Marshal(map[string]any{
		"profiles": []map[string]any{
			{"name": "alice", "dataPath": "/d/alice"},
		},
	})
	path := filepath.Join(t.TempDir(), "profiles.json")
	if err := os.WriteFile(path, encoded, 0o600); err != nil {
		t.Fatal(err)
	}

	if err := WriteActive(path, "alice"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	profiles, active, err := FromFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if active != "alice" {
		t.Errorf("expected alice, got %q", active)
	}
	if len(profiles) != 1 || profiles[0].Name != "alice" {
		t.Errorf("profiles not preserved: %+v", profiles)
	}
}

func TestFromFileTolerance(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		content      string
		writeFile    bool
		wantProfiles int
		wantActive   string
		wantErr      bool
	}{
		{
			name:      "a missing file is not an error",
			writeFile: false,
		},
		{
			name:      "an empty pre-created file is not an error",
			content:   "",
			writeFile: true,
		},
		{
			name:      "a whitespace-only file is not an error",
			content:   "\n  \t\n",
			writeFile: true,
		},
		{
			name:         "a populated file is read",
			content:      `{"active":"alice","profiles":[{"name":"alice"},{"name":"bob"}]}`,
			writeFile:    true,
			wantProfiles: 2,
			wantActive:   "alice",
		},
		{
			name:      "genuinely malformed json is still an error",
			content:   `{"profiles":[`,
			writeFile: true,
			wantErr:   true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			path := filepath.Join(t.TempDir(), "profiles.json")
			if testCase.writeFile {
				if err := os.WriteFile(path, []byte(testCase.content), 0o600); err != nil {
					t.Fatal(err)
				}
			}

			profiles, active, err := FromFile(path)
			if testCase.wantErr {
				if err == nil {
					t.Fatal("expected an error for malformed json")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(profiles) != testCase.wantProfiles {
				t.Fatalf("profiles = %d, want %d", len(profiles), testCase.wantProfiles)
			}
			if active != testCase.wantActive {
				t.Fatalf("active = %q, want %q", active, testCase.wantActive)
			}
		})
	}
}

func TestWriteProfiles_createAndUpdate(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "profiles.json")

	if err := WriteProfiles(path, []profile.Profile{{Name: "alice"}}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, _, err := FromFile(path)
	if err != nil {
		t.Fatalf("unexpected error reading back: %v", err)
	}
	if len(got) != 1 || got[0].Name != "alice" {
		t.Errorf("expected [alice], got %+v", got)
	}

	if err = WriteProfiles(path, []profile.Profile{{Name: "bob"}}); err != nil {
		t.Fatalf("unexpected error on update: %v", err)
	}
	got, _, err = FromFile(path)
	if err != nil {
		t.Fatalf("unexpected error reading back: %v", err)
	}
	if len(got) != 1 || got[0].Name != "bob" {
		t.Errorf("expected [bob] after update, got %+v", got)
	}
}

func TestWriteProfiles_preservesActive(t *testing.T) {
	t.Parallel()

	encoded, _ := json.Marshal(map[string]any{"active": "alice"})
	path := filepath.Join(t.TempDir(), "profiles.json")
	if err := os.WriteFile(path, encoded, 0o600); err != nil {
		t.Fatal(err)
	}

	if err := WriteProfiles(path, []profile.Profile{{Name: "alice"}}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	profiles, active, err := FromFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if active != "alice" {
		t.Errorf("expected alice, got %q", active)
	}
	if len(profiles) != 1 || profiles[0].Name != "alice" {
		t.Errorf("profiles not written: %+v", profiles)
	}
}

func TestWriteAtomic_replacesFileViaRename(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "profiles.json")

	if err := WriteActive(path, "alice"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	before, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}

	if err = WriteActive(path, "bob"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	after, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}

	if os.SameFile(before, after) {
		t.Error("expected the write to replace the file via rename, not truncate it in place")
	}
}

func TestWriteAtomic_fileModeIsSecret(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		write func(path string) error
	}{
		{
			name: "write profiles",
			write: func(path string) error {
				return WriteProfiles(path, []profile.Profile{{Name: "alice"}})
			},
		},
		{
			name:  "write active",
			write: func(path string) error { return WriteActive(path, "alice") },
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			path := filepath.Join(t.TempDir(), "profiles.json")
			if err := testCase.write(path); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			stat, err := os.Stat(path)
			if err != nil {
				t.Fatalf("stat: %v", err)
			}
			if stat.Mode().Perm() != fsperm.Secret {
				t.Errorf("perm = %v, want %v", stat.Mode().Perm(), fsperm.Secret)
			}
		})
	}
}

func TestMutateFile_corruptJSONPreservesFile(t *testing.T) {
	t.Parallel()

	writeProfiles := func(path string) error {
		return WriteProfiles(path, []profile.Profile{{Name: "alice"}})
	}
	writeActive := func(path string) error { return WriteActive(path, "alice") }
	corrupt := `{"active":"alice","profiles":[{"name":"alice","dataPath":"/data/alice"}]}GARBAGE`

	testCases := []struct {
		name      string
		write     func(path string) error
		writeFile bool
		content   string
		wantErr   bool
	}{
		{name: "WriteProfiles/corrupt json is rejected without touching the file", write: writeProfiles, writeFile: true, content: corrupt, wantErr: true},
		{name: "WriteProfiles/a missing file still writes fine", write: writeProfiles, writeFile: false},
		{name: "WriteProfiles/an empty file still writes fine", write: writeProfiles, writeFile: true, content: ""},
		{name: "WriteActive/corrupt json is rejected without touching the file", write: writeActive, writeFile: true, content: corrupt, wantErr: true},
		{name: "WriteActive/a missing file still writes fine", write: writeActive, writeFile: false},
		{name: "WriteActive/an empty file still writes fine", write: writeActive, writeFile: true, content: ""},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			path := filepath.Join(dir, "profiles.json")
			if testCase.writeFile {
				if err := os.WriteFile(path, []byte(testCase.content), 0o600); err != nil {
					t.Fatal(err)
				}
			}

			err := testCase.write(path)
			if !testCase.wantErr {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}

			if err == nil {
				t.Fatal("expected an error for corrupt json")
			}
			got, readErr := os.ReadFile(path)
			if readErr != nil {
				t.Fatalf("read back: %v", readErr)
			}
			if string(got) != testCase.content {
				t.Errorf("file mutated on failed write: got %q, want %q", got, testCase.content)
			}
		})
	}
}

func TestConcurrentWriteProfilesAndWriteActive(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "profiles.json")
	const rounds = 50

	for round := range rounds {
		wantProfiles := []profile.Profile{{Name: fmt.Sprintf("profile-%d", round)}}
		wantActive := fmt.Sprintf("active-%d", round)

		start := make(chan struct{})
		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			defer wg.Done()
			<-start
			if err := WriteProfiles(path, wantProfiles); err != nil {
				t.Errorf("round %d: WriteProfiles: %v", round, err)
			}
		}()
		go func() {
			defer wg.Done()
			<-start
			if err := WriteActive(path, wantActive); err != nil {
				t.Errorf("round %d: WriteActive: %v", round, err)
			}
		}()
		close(start)
		wg.Wait()

		gotProfiles, gotActive, err := FromFile(path)
		if err != nil {
			t.Fatalf("round %d: FromFile: %v", round, err)
		}
		if len(gotProfiles) != 1 || gotProfiles[0].Name != wantProfiles[0].Name {
			t.Fatalf("round %d: profiles lost, got %+v, want %+v", round, gotProfiles, wantProfiles)
		}
		if gotActive != wantActive {
			t.Fatalf("round %d: active lost, got %q, want %q", round, gotActive, wantActive)
		}
	}
}
