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
	if len(DefaultPasslist) == 0 {
		t.Fatal("DefaultPasslist should not be empty")
	}
	// HOME must always pass
	pass := false
	for _, m := range DefaultPasslist {
		if m.Match("HOME") {
			pass = true
		}
	}
	if !pass {
		t.Error("DefaultPasslist should pass HOME")
	}
}
