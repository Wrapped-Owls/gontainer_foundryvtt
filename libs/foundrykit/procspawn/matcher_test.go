package procspawn

import "testing"

func TestExactMatch(t *testing.T) {
	m := ExactMatch("HOME")
	if !m.Match("HOME") {
		t.Error("ExactMatch should match HOME")
	}
	if m.Match("HOME2") {
		t.Error("ExactMatch should not match HOME2")
	}
	if m.Match("") {
		t.Error("ExactMatch should not match empty string")
	}
}

func TestPrefixMatch(t *testing.T) {
	m := PrefixMatch("NODE_")
	if !m.Match("NODE_ENV") {
		t.Error("PrefixMatch should match NODE_ENV")
	}
	if !m.Match("NODE_OPTIONS") {
		t.Error("PrefixMatch should match NODE_OPTIONS")
	}
	if m.Match("NODE") {
		t.Error("PrefixMatch should not match NODE (missing _)")
	}
	if m.Match("MY_NODE_VAR") {
		t.Error("PrefixMatch should not match MY_NODE_VAR")
	}
}

func TestSuffixMatch(t *testing.T) {
	m := SuffixMatch("_KEY")
	if !m.Match("SECRET_KEY") {
		t.Error("SuffixMatch should match SECRET_KEY")
	}
	if !m.Match("API_KEY") {
		t.Error("SuffixMatch should match API_KEY")
	}
	if m.Match("KEY_OTHER") {
		t.Error("SuffixMatch should not match KEY_OTHER")
	}
	if m.Match("KEY") {
		t.Error("SuffixMatch should not match KEY (no prefix)")
	}
}

func TestDefaultPasslistComposition(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		key  string
		want bool
	}{
		{name: "home reaches the child", key: "HOME", want: true},
		{name: "node options reach the child", key: "NODE_OPTIONS", want: true},
		{name: "timezone reaches the child", key: "TZ", want: true},
		{name: "the loader path reaches the child", key: "LD_LIBRARY_PATH", want: true},
		{name: "secrets never reach the child", key: "FOUNDRY_ADMIN_KEY", want: false},
		{name: "the session never reaches the child", key: "FOUNDRY_SESSION", want: false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := false
			for _, m := range DefaultPasslist {
				if m.Match(testCase.key) {
					got = true
				}
			}
			if got != testCase.want {
				t.Errorf("%s passes = %v, want %v", testCase.key, got, testCase.want)
			}
		})
	}
}
