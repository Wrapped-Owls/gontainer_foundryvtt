package cmd

import "testing"

func TestVersionDoesNotPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Version panicked: %v", r)
		}
	}()
	Version()
}
