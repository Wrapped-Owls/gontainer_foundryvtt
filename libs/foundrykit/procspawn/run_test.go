package procspawn

import (
	"context"
	"syscall"
	"testing"
	"time"
)

func TestRunSuccess(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	code, err := Run(ctx, Spec{Path: "/bin/true"})
	if err != nil || code != 0 {
		t.Fatalf("true: code=%d err=%v", code, err)
	}
}

func TestRunFailure(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	code, err := Run(ctx, Spec{Path: "/bin/false"})
	if err != nil || code != 1 {
		t.Fatalf("false: code=%d err=%v", code, err)
	}
}

func TestRunContextCancellationSignalsChild(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()
	start := time.Now()
	code, err := Run(ctx, Spec{
		Path: "/bin/sh", Args: []string{"-c", "trap 'exit 0' TERM; sleep 30"},
	})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if code != 0 && code != 128+int(syscall.SIGTERM) {
		t.Fatalf("unexpected exit code %d", code)
	}
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Fatalf("child took too long to exit: %v", elapsed)
	}
}

func TestRunMissingBinary(t *testing.T) {
	_, err := Run(context.Background(), Spec{Path: "/no/such/binary"})
	if err == nil {
		t.Fatal("expected error for missing binary")
	}
}
