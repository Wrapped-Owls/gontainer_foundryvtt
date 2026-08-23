package version

import "testing"

func TestVersionParse(t *testing.T) {
	t.Parallel()

	testCases := []struct {
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
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			v := Parse(testCase.input)
			if v.String() != testCase.wantStr {
				t.Errorf("String() = %q, want %q", v.String(), testCase.wantStr)
			}
			if v.IsZero() != testCase.wantZero {
				t.Errorf("IsZero() = %v, want %v", v.IsZero(), testCase.wantZero)
			}
			hasParsed := v.parsed != nil
			if hasParsed != testCase.wantParsed {
				t.Errorf("parsed non-nil = %v, want %v", hasParsed, testCase.wantParsed)
			}
		})
	}
}

func TestVersionHasPatch(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		input string
		want  bool
	}{
		{"14.361.2", true},
		{"14.361.0", true},
		{"14.361", false},
		{"14", false},
		{"", false},
	}
	for _, testCase := range testCases {
		t.Run(testCase.input, func(t *testing.T) {
			t.Parallel()
			if got := Parse(testCase.input).HasPatch(); got != testCase.want {
				t.Errorf("Parse(%q).HasPatch() = %v, want %v", testCase.input, got, testCase.want)
			}
		})
	}
}

func TestVersionCompare(t *testing.T) {
	t.Parallel()

	testCases := []struct {
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
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			got := Parse(testCase.a).Compare(Parse(testCase.b))
			if got != testCase.want {
				t.Errorf("Parse(%q).Compare(Parse(%q)) = %d, want %d", testCase.a, testCase.b, got, testCase.want)
			}
		})
	}
}

func TestVersionMatches(t *testing.T) {
	t.Parallel()

	testCases := []struct {
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
	for _, testCase := range testCases {
		t.Run(testCase.actual+"/"+testCase.desired, func(t *testing.T) {
			t.Parallel()
			got := Parse(testCase.actual).Matches(Parse(testCase.desired))
			if got != testCase.want {
				t.Errorf("Parse(%q).Matches(Parse(%q)) = %v, want %v",
					testCase.actual, testCase.desired, got, testCase.want)
			}
		})
	}
}

func TestVersionDirName(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		input string
		want  string
	}{
		{"14.361.2", "foundryvtt_v14.361.2"},
		{"14.361", "foundryvtt_v14.361.0"},
		{"nightly", "foundryvtt_vnightly"},
	}
	for _, testCase := range testCases {
		t.Run(testCase.input, func(t *testing.T) {
			t.Parallel()
			if got := Parse(testCase.input).DirName(); got != testCase.want {
				t.Errorf("Parse(%q).DirName() = %q, want %q", testCase.input, got, testCase.want)
			}
		})
	}
}

func TestVersionMajor(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		input string
		want  uint64
	}{
		{"a full semver", "14.361.2", 14},
		{"a major.minor constraint", "13.351", 13},
		{"an unparseable value is neutral", "latest", 0},
		{"the zero value is neutral", "", 0},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			if got := Parse(testCase.input).Major(); got != testCase.want {
				t.Fatalf("Major() = %d, want %d", got, testCase.want)
			}
		})
	}
}
