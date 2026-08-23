package cliargs

func SplitSubcommand(args []string, fallback string) (string, []string) {
	if len(args) > 0 && !startsWithFlag(args[0]) {
		return args[0], args[1:]
	}
	return fallback, args
}

func startsWithFlag(s string) bool { return len(s) > 0 && s[0] == '-' }
