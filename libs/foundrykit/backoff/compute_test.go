package backoff

import (
	"testing"
	"testing/quick"
	"time"
)

func TestComputeDelaySchedule(t *testing.T) {
	cases := []struct {
		n    int
		want time.Duration
	}{
		{1, 0},
		{2, 10 * time.Second},
		{3, 20 * time.Second},
		{4, 40 * time.Second},
		{5, 80 * time.Second},
		{6, 160 * time.Second},
		{7, 320 * time.Second},
		{8, 640 * time.Second},
		{9, 960 * time.Second},
		{10, 960 * time.Second},
		{1000, 960 * time.Second},
	}
	for _, tc := range cases {
		if got := computeDelay(tc.n); got != tc.want {
			t.Errorf("computeDelay(%d) = %v, want %v", tc.n, got, tc.want)
		}
	}
}

// TestComputeDelayProperties checks that the schedule is monotonic and capped.
func TestComputeDelayProperties(t *testing.T) {
	monotonic := func(a, b uint8) bool {
		na, nb := int(a)+1, int(b)+1
		if na > nb {
			na, nb = nb, na
		}
		return computeDelay(na) <= computeDelay(nb)
	}
	if err := quick.Check(monotonic, nil); err != nil {
		t.Errorf("not monotonic: %v", err)
	}
	capped := func(n uint16) bool { return computeDelay(int(n)) <= MaxDelay }
	if err := quick.Check(capped, nil); err != nil {
		t.Errorf("not capped: %v", err)
	}
	zeroFirst := func() bool { return computeDelay(1) == 0 }
	if err := quick.Check(zeroFirst, nil); err != nil {
		t.Errorf("first failure not zero: %v", err)
	}
}
