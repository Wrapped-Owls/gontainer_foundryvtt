package profloader

import (
	"fmt"
	"os"
	"sync"
	"testing"

	"github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager/profile"
)

const (
	seededAliceActive = `{"active":"alice","profiles":[{"name":"alice"}]}`
	malformed         = seededAliceActive + "GARBAGE"
)

type writerCase struct {
	name       string
	seed       string
	isSeeded   bool
	wantNames  []string
	wantActive string
	wantErr    bool
}

func runWriterCases(t *testing.T, write func(w *Writer) error, testCases []writerCase) {
	t.Helper()
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			path := profilesPath(t)
			if testCase.isSeeded {
				seedFile(t, path, testCase.seed)
			}

			err := write(NewWriter(path))
			if (err != nil) != testCase.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, testCase.wantErr)
			}
			if !testCase.wantErr {
				assertStored(t, path, testCase.wantNames, testCase.wantActive)
				return
			}
			stored, readErr := os.ReadFile(path)
			if readErr != nil {
				t.Fatalf("read back: %v", readErr)
			}
			if string(stored) != testCase.seed {
				t.Errorf("file mutated on failed write: %q", stored)
			}
		})
	}
}

func TestWriterWriteActive(t *testing.T) {
	t.Parallel()

	runWriterCases(t, func(w *Writer) error { return w.WriteActive(profBob) }, []writerCase{
		{name: "creates a missing file", wantNames: []string{}, wantActive: profBob},
		{name: "fills an empty file", isSeeded: true, wantNames: []string{}, wantActive: profBob},
		{
			name:       "keeps the profiles",
			seed:       seededAliceActive,
			isSeeded:   true,
			wantNames:  []string{profAlice},
			wantActive: profBob,
		},
		{name: "refuses malformed json", seed: malformed, isSeeded: true, wantErr: true},
	})
}

func TestWriterWriteProfiles(t *testing.T) {
	t.Parallel()

	write := func(w *Writer) error { return w.WriteProfiles([]profile.Profile{{Name: profBob}}) }
	runWriterCases(t, write, []writerCase{
		{name: "creates a missing file", wantNames: []string{profBob}},
		{name: "fills an empty file", isSeeded: true, wantNames: []string{profBob}},
		{
			name:       "keeps the active name",
			seed:       seededAliceActive,
			isSeeded:   true,
			wantNames:  []string{profBob},
			wantActive: profAlice,
		},
		{name: "refuses malformed json", seed: malformed, isSeeded: true, wantErr: true},
	})
}

func TestWriterSerializesConcurrentWrites(t *testing.T) {
	t.Parallel()

	path := profilesPath(t)
	writer := NewWriter(path)
	for round := range 50 {
		wantName := fmt.Sprintf("profile-%d", round)
		wantActive := fmt.Sprintf("active-%d", round)

		start := make(chan struct{})
		var wg sync.WaitGroup
		wg.Go(func() {
			<-start
			if err := writer.WriteProfiles([]profile.Profile{{Name: wantName}}); err != nil {
				t.Errorf("round %d: WriteProfiles: %v", round, err)
			}
		})
		wg.Go(func() {
			<-start
			if err := writer.WriteActive(wantActive); err != nil {
				t.Errorf("round %d: WriteActive: %v", round, err)
			}
		})
		close(start)
		wg.Wait()

		assertStored(t, path, []string{wantName}, wantActive)
	}
}
