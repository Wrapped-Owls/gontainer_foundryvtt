package backoff

import (
	"fmt"
	"time"
)

const MaxDelay = 960 * time.Second

const BaseDelay = 10 * time.Second

// HealthyUptime sits above Foundry's boot time so a startup crash never counts as
// a recovery.
const HealthyUptime = 60 * time.Second

// MaxConsecutiveFailures is where restarting the same process in the same container
// stops being a plausible fix and the environment itself needs recreating.
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

// Decision is OnFailure's result: log by Mode and Delay, wait out Delay unless it
// is zero or ModeKubernetes, then propagate ExitCode.
type Decision struct {
	Mode      Mode
	Delay     time.Duration
	ExitCode  int
	State     State
	StateFile string // empty when Mode != ModeBackoff
}

func (d Decision) IsExhausted() bool {
	return d.State.ConsecutiveFailures >= MaxConsecutiveFailures
}

type Manager struct {
	// CacheDir, when non-empty, is where backoff_state.json lives.
	CacheDir string
	// KubernetesBypass triggers the early return; NewFromEnv sets its default.
	KubernetesBypass bool
	// memFailures mirrors the persisted counter, and stands in for it entirely
	// when the state file is unreachable.
	memFailures int
}
