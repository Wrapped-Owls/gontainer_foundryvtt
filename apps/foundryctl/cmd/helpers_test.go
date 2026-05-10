package cmd

import "testing"

func TestOrDefault(t *testing.T) {
	if got := orDefault("a", "b"); got != "a" {
		t.Errorf("got %q", got)
	}
	if got := orDefault("", "b"); got != "b" {
		t.Errorf("got %q", got)
	}
}
