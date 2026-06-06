package profloader

import (
	"testing"
)

func TestFromEnv_empty(t *testing.T) {
	profiles, err := FromEnv("FOUNDRY_PROFILE_X_UNIQUE")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(profiles) != 0 {
		t.Errorf("expected 0, got %d", len(profiles))
	}
}

func TestFromEnv_numericKeys(t *testing.T) {
	t.Setenv("TEST_PROF_0_NAME", "alice")
	t.Setenv("TEST_PROF_0_DATA_PATH", "/data/alice")
	t.Setenv("TEST_PROF_1_NAME", "bob")
	t.Setenv("TEST_PROF_1_DATA_PATH", "/data/bob")

	profiles, err := FromEnv("TEST_PROF")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(profiles) != 2 {
		t.Fatalf("expected 2, got %d", len(profiles))
	}
	if profiles[0].Name != "alice" || profiles[0].DataPath != "/data/alice" {
		t.Errorf("unexpected profile[0]: %+v", profiles[0])
	}
	if profiles[1].Name != "bob" {
		t.Errorf("unexpected profile[1]: %+v", profiles[1])
	}
}

func TestFromEnv_namedKeys(t *testing.T) {
	t.Setenv("TEST_NAMED_CHARLIE_NAME", "charlie")
	t.Setenv("TEST_NAMED_CHARLIE_DATA_PATH", "/data/charlie")

	profiles, err := FromEnv("TEST_NAMED")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(profiles) != 1 {
		t.Fatalf("expected 1, got %d", len(profiles))
	}
	if profiles[0].Name != "charlie" || profiles[0].DataPath != "/data/charlie" {
		t.Errorf("unexpected profile: %+v", profiles[0])
	}
}

func TestFromEnv_mixedKeys(t *testing.T) {
	t.Setenv("TEST_MIX_0_NAME", "zero")
	t.Setenv("TEST_MIX_ALPHA_NAME", "alpha")
	t.Setenv("TEST_MIX_2_NAME", "two")

	profiles, err := FromEnv("TEST_MIX")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(profiles) != 3 {
		t.Fatalf("expected 3, got %d", len(profiles))
	}
	// numerics first (0, 2), then string (ALPHA)
	if profiles[0].Name != "zero" {
		t.Errorf("expected zero first, got %q", profiles[0].Name)
	}
	if profiles[1].Name != "two" {
		t.Errorf("expected two second, got %q", profiles[1].Name)
	}
	if profiles[2].Name != "alpha" {
		t.Errorf("expected alpha third, got %q", profiles[2].Name)
	}
}

func TestFromEnv_partialFields(t *testing.T) {
	t.Setenv("TEST_PARTIAL_0_NAME", "alice")
	t.Setenv("TEST_PARTIAL_0_VERSION", "14.0.0")

	profiles, err := FromEnv("TEST_PARTIAL")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(profiles) != 1 {
		t.Fatalf("expected 1, got %d", len(profiles))
	}
	if profiles[0].Name != "alice" || profiles[0].Version != "14.0.0" {
		t.Errorf("unexpected profile: %+v", profiles[0])
	}
	if profiles[0].DataPath != "" {
		t.Errorf("expected empty DataPath, got %q", profiles[0].DataPath)
	}
}
