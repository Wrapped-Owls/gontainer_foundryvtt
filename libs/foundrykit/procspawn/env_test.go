package procspawn

import (
	"reflect"
	"testing"
)

func TestFilterEnvDefaultPasslist(t *testing.T) {
	in := []string{
		"HOME=/root",
		"PATH=/usr/bin",
		"NODE_OPTIONS=--max-old-space-size=4096",
		"NODE_ENV=production",
		"TZ=UTC",
		"FOUNDRY_PASSWORD=secret",
		"NOTHING",
	}
	want := []string{
		"HOME=/root",
		"NODE_OPTIONS=--max-old-space-size=4096",
		"NODE_ENV=production",
		"TZ=UTC",
	}
	got := FilterEnv(in, DefaultPasslist)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("FilterEnv = %v, want %v", got, want)
	}
}

func TestFilterEnvCustomMatchers(t *testing.T) {
	matchers := []Matcher{
		ExactMatch("HOME"),
		PrefixMatch("MY_"),
		SuffixMatch("_KEY"),
	}
	in := []string{
		"HOME=/root",
		"MY_VAR=foo",
		"SECRET_KEY=bar",
		"OTHER=ignored",
	}
	want := []string{"HOME=/root", "MY_VAR=foo", "SECRET_KEY=bar"}
	got := FilterEnv(in, matchers)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("FilterEnv = %v, want %v", got, want)
	}
}

func TestFilterEnvNoMatch(t *testing.T) {
	in := []string{"FOO=bar", "BAZ=qux"}
	got := FilterEnv(in, []Matcher{ExactMatch("NONE")})
	if len(got) != 0 {
		t.Errorf("expected empty, got %v", got)
	}
}

func TestFilterEnvSkipsMalformed(t *testing.T) {
	in := []string{"NOEQUALS", "GOOD=value"}
	got := FilterEnv(in, []Matcher{ExactMatch("GOOD")})
	if len(got) != 1 || got[0] != "GOOD=value" {
		t.Errorf("expected [GOOD=value], got %v", got)
	}
}
