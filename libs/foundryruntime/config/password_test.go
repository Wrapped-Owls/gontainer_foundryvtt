package config

import "testing"

func TestHashAdminKeyBasic(t *testing.T) {
	got, err := HashAdminKey("hunter2", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != pbkdfKeyLen*2 {
		t.Fatalf("expected %d hex chars, got %d", pbkdfKeyLen*2, len(got))
	}
	// Deterministic given same inputs.
	got2, _ := HashAdminKey("hunter2", "")
	if got != got2 {
		t.Error("hash should be deterministic")
	}
}

func TestHashAdminKeyWithSalt(t *testing.T) {
	a, _ := HashAdminKey("hunter2", "")
	b, _ := HashAdminKey("hunter2", "custom-salt")
	if a == b {
		t.Error("different salt should produce different hash")
	}
}

func TestHashAdminKeyEmptyRejects(t *testing.T) {
	if _, err := HashAdminKey("", ""); err == nil {
		t.Error("empty key should error")
	}
	if _, err := HashAdminKey("   ", ""); err == nil {
		t.Error("whitespace-only key should error")
	}
}

func TestHashAdminKeyTrimsWhitespace(t *testing.T) {
	a, _ := HashAdminKey("  abc  ", "")
	b, _ := HashAdminKey("abc", "")
	if a != b {
		t.Error("leading/trailing whitespace should be trimmed before hashing")
	}
}
