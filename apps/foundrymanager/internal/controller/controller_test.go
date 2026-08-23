package controller

import (
	"context"
	"errors"
	"testing"
)

func TestRequestSwitch_cancelsContext(t *testing.T) {
	t.Parallel()

	c := New()
	ctx, cancel := context.WithCancelCause(context.Background())
	defer cancel(nil)
	c.SetCancel(cancel)

	c.RequestSwitch("alice")

	select {
	case name := <-c.SwitchCh:
		if name != "alice" {
			t.Errorf("expected alice, got %q", name)
		}
	default:
		t.Fatal("expected name in SwitchCh")
	}
	if ctx.Err() == nil {
		t.Fatal("expected context cancelled")
	}
	if !errors.Is(context.Cause(ctx), ErrProfileSwitch) {
		t.Errorf("unexpected cause: %v", context.Cause(ctx))
	}
}

func TestRequestSwitch_replacesPending(t *testing.T) {
	t.Parallel()

	c := New()

	c.RequestSwitch("alice")
	c.RequestSwitch("bob")

	name := <-c.SwitchCh
	if name != "bob" {
		t.Errorf("expected bob (latest), got %q", name)
	}
}

func TestActive(t *testing.T) {
	t.Parallel()

	c := New()
	if c.Active() != "" {
		t.Error("expected empty initial active")
	}
	c.SetActive("alice")
	if c.Active() != "alice" {
		t.Errorf("expected alice, got %q", c.Active())
	}
}

func TestRequestRestartCancelsWithoutQueueingASwitch(t *testing.T) {
	t.Parallel()

	c := New()
	_, cancel := context.WithCancelCause(context.Background())
	defer cancel(nil)
	var cause error
	c.SetCancel(func(err error) { cause = err; cancel(err) })

	c.RequestRestart()

	if !errors.Is(cause, ErrRestart) {
		t.Fatalf("cancel cause = %v, want %v", cause, ErrRestart)
	}
	select {
	case name := <-c.SwitchCh:
		t.Fatalf("restart queued a switch to %q", name)
	default:
	}
}

func TestRequestRestartWithoutASessionIsANoop(t *testing.T) {
	t.Parallel()

	if New().RequestRestart() {
		t.Fatal("RequestRestart reported a cancelled session when none was running")
	}
}

func TestRequestRestartAfterSwitch(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		reRegister  bool
		wantRestart bool
	}{
		{
			name:        "restart right after a switch is stale and reports no live session",
			reRegister:  false,
			wantRestart: false,
		},
		{
			name:        "restart after a new session registers its cancel is live again",
			reRegister:  true,
			wantRestart: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			c := New()
			_, switchCancel := context.WithCancelCause(context.Background())
			defer switchCancel(nil)
			var switchCause error
			c.SetCancel(func(err error) { switchCause = err; switchCancel(err) })

			c.RequestSwitch("alice")
			if !errors.Is(switchCause, ErrProfileSwitch) {
				t.Fatalf("switch cancel cause = %v, want %v", switchCause, ErrProfileSwitch)
			}
			select {
			case name := <-c.SwitchCh:
				if name != "alice" {
					t.Fatalf("expected alice queued, got %q", name)
				}
			default:
				t.Fatal("expected the switch to queue a name")
			}

			var restartCause error
			if testCase.reRegister {
				_, restartCancel := context.WithCancelCause(context.Background())
				defer restartCancel(nil)
				c.SetCancel(func(err error) { restartCause = err; restartCancel(err) })
			}

			got := c.RequestRestart()
			if got != testCase.wantRestart {
				t.Fatalf("RequestRestart() = %v, want %v", got, testCase.wantRestart)
			}
			if !testCase.reRegister {
				if errors.Is(switchCause, ErrRestart) {
					t.Fatal("restart fired the switch's stale cancel a second time")
				}
				return
			}
			if !errors.Is(restartCause, ErrRestart) {
				t.Fatalf("restart cancel cause = %v, want %v", restartCause, ErrRestart)
			}
		})
	}
}
