package procspawn

import (
	"os"
	"slices"
	"syscall"
	"testing"
)

func TestSpecWithDefaults(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		spec Spec
	}{
		{"zero value", Spec{Path: "/bin/true"}},
		{"env already set", Spec{Path: "/bin/true", Env: []string{"FOO=bar"}}},
		{
			"forward signals already set",
			Spec{Path: "/bin/true", ForwardSignals: []os.Signal{syscall.SIGHUP}},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			want := testCase.spec
			got := testCase.spec.withDefaults()

			if got.Env == nil {
				t.Error("Env should be non-nil after withDefaults")
			}
			if want.Env != nil && !slices.Equal(got.Env, want.Env) {
				t.Errorf("Env should be preserved when already set, got %v want %v", got.Env, want.Env)
			}
			if len(got.ForwardSignals) == 0 {
				t.Error("ForwardSignals should be non-empty after withDefaults")
			}
			if want.ForwardSignals != nil && !slices.Equal(got.ForwardSignals, want.ForwardSignals) {
				t.Errorf(
					"ForwardSignals should be preserved when already set, got %v want %v",
					got.ForwardSignals,
					want.ForwardSignals,
				)
			}
			if got.Stdin != os.Stdin {
				t.Error("Stdin should default to os.Stdin")
			}
			if got.Stdout != os.Stdout {
				t.Error("Stdout should default to os.Stdout")
			}
			if got.Stderr != os.Stderr {
				t.Error("Stderr should default to os.Stderr")
			}
		})
	}
}

func TestSpecWithDefaultsPreservesCustomStreams(t *testing.T) {
	t.Parallel()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	t.Cleanup(func() {
		_ = r.Close()
		_ = w.Close()
	})

	spec := Spec{Path: "/bin/true", Stdin: r, Stdout: w, Stderr: w}
	got := spec.withDefaults()

	if got.Stdin != r {
		t.Error("Stdin should be preserved when already set")
	}
	if got.Stdout != w {
		t.Error("Stdout should be preserved when already set")
	}
	if got.Stderr != w {
		t.Error("Stderr should be preserved when already set")
	}
}
