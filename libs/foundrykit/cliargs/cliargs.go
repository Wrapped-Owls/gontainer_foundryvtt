package cliargs

// SplitSubcommand returns the leading subcommand and the args after it. No args,
// or a leading flag, yields fallback with args untouched.
func SplitSubcommand(args []string, fallback string) (string, []string) {
	if len(args) > 0 && !startsWithFlag(args[0]) {
		return args[0], args[1:]
	}
	return fallback, args
}

func startsWithFlag(s string) bool { return len(s) > 0 && s[0] == '-' }
