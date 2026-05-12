package procspawn

import (
	"os"
	"syscall"
)

// defaultSignals is the set of OS signals forwarded to the child by default.
var defaultSignals = []os.Signal{syscall.SIGTERM, syscall.SIGINT}

// Config configures the process spawner.
type Config struct {
	// Passlist is the set of env-key matchers forwarded to the child.
	// Default: DefaultPasslist.
	Passlist []Matcher
	// ForwardSignals is the set of signals forwarded to the child's process group.
	// Default: SIGTERM + SIGINT.
	ForwardSignals []os.Signal
}

// Default returns a Config with the standard passlist and signal set.
func Default() Config {
	return Config{
		Passlist:       DefaultPasslist,
		ForwardSignals: defaultSignals,
	}
}
