package backoff

import "testing"

func TestDecisionIsExhausted(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		failures int
		want     bool
	}{
		{"a fresh failure is recoverable", 1, false},
		{"one short of the budget still restarts", MaxConsecutiveFailures - 1, false},
		{"at the budget the environment is the problem", MaxConsecutiveFailures, true},
		{"past the budget stays exhausted", MaxConsecutiveFailures + 5, true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			dec := Decision{State: State{ConsecutiveFailures: testCase.failures}}
			if got := dec.IsExhausted(); got != testCase.want {
				t.Fatalf("IsExhausted() = %v, want %v", got, testCase.want)
			}
		})
	}
}
