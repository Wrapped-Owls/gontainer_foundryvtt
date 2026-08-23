package procspawn

import (
	"syscall"
	"testing"
)

func TestExitCodeFromWaitStatus(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		ws       syscall.WaitStatus
		wantCode int
		wantOK   bool
	}{
		{"exited zero", syscall.WaitStatus(0), 0, true},
		{"exited nonzero", syscall.WaitStatus(42 << 8), 42, true},
		{"signaled term", syscall.WaitStatus(syscall.SIGTERM), 128 + int(syscall.SIGTERM), true},
		{"signaled kill", syscall.WaitStatus(syscall.SIGKILL), 128 + int(syscall.SIGKILL), true},
		{"stopped", syscall.WaitStatus(0x7F), -1, false},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			gotCode, gotOK := exitCodeFromWaitStatus(testCase.ws)
			if gotOK != testCase.wantOK {
				t.Fatalf("ok = %v, want %v", gotOK, testCase.wantOK)
			}
			if gotCode != testCase.wantCode {
				t.Fatalf("code = %d, want %d", gotCode, testCase.wantCode)
			}
		})
	}
}
