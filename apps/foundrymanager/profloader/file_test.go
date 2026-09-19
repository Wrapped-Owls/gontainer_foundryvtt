package profloader

import (
	"slices"
	"testing"
)

func TestFromFile(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		content    string
		isSeeded   bool
		wantNames  []string
		wantActive string
		wantErr    bool
	}{
		{name: "a missing file is not an error", wantNames: []string{}},
		{name: "an empty pre-created file is not an error", isSeeded: true, wantNames: []string{}},
		{
			name:      "a whitespace-only file is not an error",
			content:   "\n  \t\n",
			isSeeded:  true,
			wantNames: []string{},
		},
		{
			name:       "a populated file is read",
			content:    `{"active":"alice","profiles":[{"name":"alice","dataPath":"/data/alice"},{"name":"bob"}]}`,
			isSeeded:   true,
			wantNames:  []string{profAlice, profBob},
			wantActive: profAlice,
		},
		{
			name:     "malformed json is an error",
			content:  `{"profiles":[`,
			isSeeded: true,
			wantErr:  true,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			path := profilesPath(t)
			if testCase.isSeeded {
				seedFile(t, path, testCase.content)
			}

			profiles, active, err := FromFile(path)
			if (err != nil) != testCase.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, testCase.wantErr)
			}
			if testCase.wantErr {
				return
			}
			if names := profileNames(profiles); !slices.Equal(names, testCase.wantNames) {
				t.Errorf("profiles = %v, want %v", names, testCase.wantNames)
			}
			if active != testCase.wantActive {
				t.Errorf("active = %q, want %q", active, testCase.wantActive)
			}
		})
	}
}
