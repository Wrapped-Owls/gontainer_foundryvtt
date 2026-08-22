package backoff

import (
	"fmt"
	"time"
)

const MaxDelay = 960 * time.Second

const BaseDelay = 10 * time.Second

const HealthyUptime = 60 * time.Second

const MaxConsecutiveFailures = 10

const stateFile = "backoff_state.json"

type State struct {
	ConsecutiveFailures int    `json:"consecutive_failures"`
	LastFailureTS       string `json:"last_failure_timestamp"`
}

type Mode int

const (
	ModeKubernetes Mode = iota
	ModeNoCache
	ModeBackoff
)

func (m Mode) String() string {
	switch m {
	case ModeKubernetes:
		return "kubernetes"
	case ModeNoCache:
		return "no-cache"
	case ModeBackoff:
		return "backoff"
	}
	return fmt.Sprintf("mode(%d)", int(m))
}

type Decision struct {
	Mode      Mode
	Delay     time.Duration
	ExitCode  int
	State     State
	StateFile string
}

func (d Decision) IsExhausted() bool {
	return d.State.ConsecutiveFailures >= MaxConsecutiveFailures
}

type Tracker struct {
	CacheDir         string
	KubernetesBypass bool
	memFailures      int
}
