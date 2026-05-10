package procspawn

import (
	"os"
	"syscall"

	"github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrykit/confloader"
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

func LoadFromEnv(c *Config) error {
	return confloader.BindEnv()
}
