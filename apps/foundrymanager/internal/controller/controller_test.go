package controller

import (
	"context"
	"errors"
	"testing"
)

func TestRequestSwitch_cancelsContext(t *testing.T) {
	ctrl := New()
	ctx, cancel := context.WithCancelCause(context.Background())
	defer cancel(nil)
	ctrl.SetCancel(cancel)

	ctrl.RequestSwitch("alice")

	select {
	case name := <-ctrl.SwitchCh:
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
	ctrl := New()

	ctrl.RequestSwitch("alice")
	ctrl.RequestSwitch("bob")

	name := <-ctrl.SwitchCh
	if name != "bob" {
		t.Errorf("expected bob (latest), got %q", name)
	}
}

func TestActive(t *testing.T) {
	ctrl := New()
	if ctrl.Active() != "" {
		t.Error("expected empty initial active")
	}
	ctrl.SetActive("alice")
	if ctrl.Active() != "alice" {
		t.Errorf("expected alice, got %q", ctrl.Active())
	}
}
