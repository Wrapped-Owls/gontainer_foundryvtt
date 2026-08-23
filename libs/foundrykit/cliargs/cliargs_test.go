package cliargs

import (
	"slices"
	"testing"
)

func TestSplitSubcommand(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		args     []string
		want     string
		wantArgs []string
	}{
		{name: "no arguments falls back", want: "run"},
		{name: "an empty argument is not the fallback", args: []string{""}, want: ""},
		{
			name:     "a leading flag falls back and keeps the flag",
			args:     []string{"--verbose"},
			want:     "run",
			wantArgs: []string{"--verbose"},
		},
		{
			name:     "a subcommand is split from its args",
			args:     []string{"healthcheck", "--foo"},
			want:     "healthcheck",
			wantArgs: []string{"--foo"},
		},
		{name: "an unknown subcommand is returned as is", args: []string{"frobnicate"}, want: "frobnicate"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			sub, rest := SplitSubcommand(testCase.args, "run")
			if sub != testCase.want {
				t.Fatalf("sub = %q, want %q", sub, testCase.want)
			}
			if !slices.Equal(rest, testCase.wantArgs) {
				t.Fatalf("args = %v, want %v", rest, testCase.wantArgs)
			}
		})
	}
}
