package applier

import (
	"context"
	"errors"
	"testing"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrypatch/manifest"
)

func TestApplyRejectsEscape(t *testing.T) {
	a := &Applier{Root: t.TempDir()}
	err := a.Apply(context.Background(), []manifest.Patch{{
		ID: "p", Versions: ">=1",
		Actions: []manifest.Action{{Type: manifest.ActionFileReplace, Dest: "../../etc/x"}},
	}}, nil)
	if err == nil {
		t.Fatal("expected error for escape path")
	}
}

func TestSafeJoinAbsolutePathRejects(t *testing.T) {
	a := &Applier{Root: t.TempDir()}
	_, err := a.safeJoin("/etc/passwd")
	if err == nil {
		t.Fatal("absolute path should be rejected")
	}
}

func TestSafeJoinDotDotRejects(t *testing.T) {
	a := &Applier{Root: t.TempDir()}
	_, err := a.safeJoin("../escape")
	if err == nil {
		t.Fatal("path with .. should be rejected")
	}
}

func TestSafeJoinValidPath(t *testing.T) {
	root := t.TempDir()
	a := &Applier{Root: root}
	got, err := a.safeJoin("sub/dir/file.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == "" {
		t.Error("got empty path")
	}
}

func TestRunActionUnknownType(t *testing.T) {
	a := &Applier{Root: t.TempDir(), HTTPClient: nil}
	act := manifest.Action{Type: "bogus-action", Dest: "file.txt"}
	err := a.runAction(context.Background(), act)
	if err == nil || !errors.Is(err, manifest.ErrUnknownAction) {
		t.Fatalf("expected ErrUnknownAction, got %v", err)
	}
}
