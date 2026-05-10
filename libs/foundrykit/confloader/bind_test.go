package confloader_test

import (
	"errors"
	"testing"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrykit/confloader"
)

func TestBindFieldUnset(t *testing.T) {
	var s string
	binder := confloader.BindField(&s, "CONFLOADER_TEST_UNSET_VAR_XYZ", nil)
	if err := binder(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s != "" {
		t.Errorf("field should remain empty when var unset, got %q", s)
	}
}

func TestBindFieldSet(t *testing.T) {
	t.Setenv("CONFLOADER_TEST_HOST", "myhost")
	var s string
	binder := confloader.BindField(&s, "CONFLOADER_TEST_HOST", nil)
	if err := binder(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s != "myhost" {
		t.Errorf("got %q, want myhost", s)
	}
}

func TestBindFieldParseError(t *testing.T) {
	t.Setenv("CONFLOADER_TEST_INT", "notanint")
	var n int
	binder := confloader.BindField(&n, "CONFLOADER_TEST_INT", func(v string) (int, error) {
		return 0, errors.New("parse error")
	})
	if err := binder(); err == nil {
		t.Fatal("expected parse error")
	}
}

func TestBindFieldPresentTreatsEmptyAsValid(t *testing.T) {
	t.Setenv("CONFLOADER_TEST_EMPTY", "")
	s := "original"
	binder := confloader.BindFieldPresent(&s, "CONFLOADER_TEST_EMPTY", nil)
	if err := binder(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s != "" {
		t.Errorf("explicit empty should set field to empty, got %q", s)
	}
}

func TestBindFieldPresentAbsent(t *testing.T) {
	// Unset var: field should not change.
	s := "keep"
	binder := confloader.BindFieldPresent(&s, "CONFLOADER_TEST_ABSENT_XYZ", nil)
	if err := binder(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s != "keep" {
		t.Errorf("absent var should not change field, got %q", s)
	}
}

func TestBindRequiredMissing(t *testing.T) {
	var s string
	err := confloader.BindEnv(confloader.BindRequired(&s, "CONFLOADER_REQUIRED_MISSING_XYZ", nil))
	if err == nil {
		t.Fatal("expected error for missing required var")
	}
}

func TestBindRequiredPresent(t *testing.T) {
	t.Setenv("CONFLOADER_REQUIRED_PRESENT", "hello")
	var s string
	err := confloader.BindEnv(confloader.BindRequired(&s, "CONFLOADER_REQUIRED_PRESENT", nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s != "hello" {
		t.Errorf("got %q, want hello", s)
	}
}

func TestBindEnvRunsAll(t *testing.T) {
	count := 0
	err := confloader.BindEnv(
		func() error { count++; return nil },
		func() error { count++; return nil },
	)
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Errorf("expected 2 binders called, got %d", count)
	}
}
