package colorlog

import "testing"

func TestLevelFromEnvVariants(t *testing.T) {
	cases := []struct {
		name, ll, verbose string
		want              Level
	}{
		{"default", "", "", LevelInfo},
		{"verbose-flips-to-debug", "", "1", LevelDebug},
		{"explicit-warn", "warn", "1", LevelWarn},
		{"explicit-error", "ERROR", "", LevelError},
		{"unknown-falls-back", "bogus", "", LevelInfo},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("CONTAINER_LOG_LEVEL", tc.ll)
			t.Setenv("CONTAINER_VERBOSE", tc.verbose)
			if got := LevelFromEnv(); got != tc.want {
				t.Errorf("LevelFromEnv() = %v, want %v", got, tc.want)
			}
		})
	}
}
