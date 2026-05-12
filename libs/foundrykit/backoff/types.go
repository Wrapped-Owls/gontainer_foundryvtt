// Package backoff implements persistent exponential backoff for the Foundry
// controller. It:
//
//   - persists backoff state in <CacheDir>/backoff_state.json
//   - delay = min(10 * 2^(n-2), 960) seconds where n is the 1-based
//     consecutive failure count, with n==1 yielding zero delay
//   - Kubernetes bypass: when KUBERNETES_SERVICE_HOST is non-empty, all
//     OnFailure() calls return immediately so CrashLoopBackOff handles
//     restart throttling
//   - "no cache" mode: when CacheDir cannot be created or is empty, the
//     controller is expected to sleep indefinitely (the Decision.Mode
//     reflects that so the caller can hook a context-cancellable sleep)
//
// The package itself does not perform sleeping. The caller (the runtime
// controller) decides how to wait so that SIGTERM / context cancellation
// can interrupt the wait promptly. A helper Sleep() that blocks until
// ctx is done or the delay elapses is provided for convenience.
package backoff

import (
	"fmt"
	"math/rand/v2"
	"time"
)

// MaxDelay caps the computed delay at 960 seconds.
const MaxDelay = 960 * time.Second

// BaseDelay is the multiplier used in the exponential schedule.
const BaseDelay = 10 * time.Second

// stateFile is the basename written inside CacheDir.
const stateFile = "backoff_state.json"

// State is the on-disk schema for persisted backoff state.
type State struct {
	ConsecutiveFailures int    `json:"consecutive_failures"`
	LastFailureTS       string `json:"last_failure_timestamp"`
}

// Mode classifies how OnFailure decided to handle the failure.
type Mode int

const (
	// ModeKubernetes — bypassed, no wait required.
	ModeKubernetes Mode = iota
	// ModeNoCache — indefinite sleep required (no persistent state available).
	ModeNoCache
	// ModeBackoff — wait for Delay then exit with the original code.
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

// Decision is the result of OnFailure. The caller is expected to:
//
//   - log according to Mode + Delay
//   - sleep (or wait on ctx) for Delay (skip if zero or ModeKubernetes)
//   - propagate ExitCode
//
// State is the post-write counter (useful for logging / tests).
type Decision struct {
	Mode      Mode
	Delay     time.Duration
	ExitCode  int
	State     State
	StateFile string // empty when Mode != ModeBackoff
}

// Manager is the stateful entry-point. Construct with New().
type Manager struct {
	// CacheDir, when non-empty, is where backoff_state.json lives.
	CacheDir string
	// KubernetesBypass enables the early-return branch. Default driven
	// by NewFromEnv(); manual users set it explicitly.
	KubernetesBypass bool
	// Now allows tests to inject a deterministic clock.
	Now func() time.Time
	// Rand allows tests to inject a deterministic jitter source. Currently
	// unused — included so the schedule can grow jitter later without an
	// API break.
	Rand *rand.Rand
}
