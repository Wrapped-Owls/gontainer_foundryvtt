package procspawn

import (
	"os"
	"syscall"
)

var defaultSignals = []os.Signal{syscall.SIGTERM, syscall.SIGINT}

type Config struct {
	Passlist       []Matcher
	ForwardSignals []os.Signal
}

func Default() Config {
	return Config{
		Passlist:       DefaultPasslist,
		ForwardSignals: defaultSignals,
	}
}
