package colorlog

import "testing"

func TestLevelFromEnvVariants(t *testing.T) {
	testCases := []struct {
		name, ll, verbose string
		want              Level
	}{
		{"default", "", "", LevelInfo},
		{"verbose-flips-to-debug", "", "1", LevelDebug},
		{"explicit-warn", "warn", "1", LevelWarn},
		{"explicit-error", "ERROR", "", LevelError},
		{"unknown-falls-back", "bogus", "", LevelInfo},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Setenv("CONTAINER_LOG_LEVEL", testCase.ll)
			t.Setenv("CONTAINER_VERBOSE", testCase.verbose)
			if got := LevelFromEnv(); got != testCase.want {
				t.Errorf("LevelFromEnv() = %v, want %v", got, testCase.want)
			}
		})
	}
}
