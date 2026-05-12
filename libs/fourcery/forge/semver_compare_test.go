package forge

import "testing"

func TestCompareSemver(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		a, b string
		want int
	}{
		{"equal", "14.361.2", "14.361.2", 0},
		{"a greater minor", "14.361.0", "14.360.0", 1},
		{"b greater minor", "14.360.0", "14.361.0", -1},
		{"a greater patch", "14.361.2", "14.361.0", 1},
		{"b greater patch", "14.361.0", "14.361.2", -1},
		{"a non-semver", "not-semver", "14.361.2", -1},
		{"b non-semver", "14.361.2", "not-semver", 1},
		{"both non-semver", "foo", "bar", 0},
		{"both empty", "", "", 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := compareSemver(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("compareSemver(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}
