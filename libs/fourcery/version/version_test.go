package version

import "testing"

func TestVersionParse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		input      string
		wantStr    string
		wantZero   bool
		wantParsed bool
	}{
		{"semver canonicalised", "14.361.2", "14.361.2", false, true},
		{"semver with v prefix", "v14.361.2", "14.361.2", false, true},
		{"partial semver", "14.361", "14.361.0", false, true},
		{"non-semver preserved", "nightly", "nightly", false, false},
		{"empty is zero", "", "", true, false},
		{"whitespace trimmed", "  14.361.2  ", "14.361.2", false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			v := Parse(tt.input)
			if v.String() != tt.wantStr {
				t.Errorf("String() = %q, want %q", v.String(), tt.wantStr)
			}
			if v.IsZero() != tt.wantZero {
				t.Errorf("IsZero() = %v, want %v", v.IsZero(), tt.wantZero)
			}
			hasParsed := v.parsed != nil
			if hasParsed != tt.wantParsed {
				t.Errorf("parsed non-nil = %v, want %v", hasParsed, tt.wantParsed)
			}
		})
	}
}

func TestVersionHasPatch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		want  bool
	}{
		{"14.361.2", true},
		{"14.361.0", true},
		{"14.361", false},
		{"14", false},
		{"", false},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			if got := Parse(tt.input).HasPatch(); got != tt.want {
				t.Errorf("Parse(%q).HasPatch() = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestVersionCompare(t *testing.T) {
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
			got := Parse(tt.a).Compare(Parse(tt.b))
			if got != tt.want {
				t.Errorf("Parse(%q).Compare(Parse(%q)) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestVersionMatches(t *testing.T) {
	t.Parallel()

	tests := []struct {
		actual, desired string
		want            bool
	}{
		{"14.361.2", "14.361.2", true},
		{"14.361.2", "14.361", true},
		{"14.361.0", "14.361.2", false},
		{"14.361.2", "14.362.0", false},
		{"14.361.0", "14.361", true},
		{"v14.361.0", "14.361", true},
		{"", "", true},
		{"14.361", "", false},
		{"nightly", "nightly", true},
		{"nightly", "14.361", false},
	}
	for _, tt := range tests {
		t.Run(tt.actual+"/"+tt.desired, func(t *testing.T) {
			t.Parallel()
			got := Parse(tt.actual).Matches(Parse(tt.desired))
			if got != tt.want {
				t.Errorf("Parse(%q).Matches(Parse(%q)) = %v, want %v",
					tt.actual, tt.desired, got, tt.want)
			}
		})
	}
}

func TestVersionDirName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		want  string
	}{
		{"14.361.2", "foundryvtt_v14.361.2"},
		{"14.361", "foundryvtt_v14.361.0"},
		{"nightly", "foundryvtt_vnightly"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			if got := Parse(tt.input).DirName(); got != tt.want {
				t.Errorf("Parse(%q).DirName() = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
