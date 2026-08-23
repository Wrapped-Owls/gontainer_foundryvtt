package controller

import (
	"context"
	"errors"
	"sync"
)

// ErrProfileSwitch is the cancel cause for a switch; check it with errors.Is.
var ErrProfileSwitch = errors.New("foundrymanager: profile switch requested")

// ErrRestart marks a restart's cancel cause, distinct from a switch or a shutdown.
var ErrRestart = errors.New("foundrymanager: restart requested")

// SwitchController coordinates the handoff between the dashboard HTTP handler
// and the manager's process loop. All methods are safe for concurrent use.
type SwitchController struct {
	mu       sync.Mutex
	cancelFn context.CancelCauseFunc
	current  string
	SwitchCh chan string
}

func New() *SwitchController {
	return &SwitchController{SwitchCh: make(chan string, 1)}
}

func (c *SwitchController) SetCancel(fn context.CancelCauseFunc) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cancelFn = fn
}

func (c *SwitchController) SetActive(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.current = name
}

func (c *SwitchController) Active() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.current
}

// RequestRestart reports whether a live session was actually cancelled. A stale or
// absent cancel means the request went nowhere, and the caller must not claim it did.
func (c *SwitchController) RequestRestart() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cancelFn == nil {
		return false
	}
	c.cancelFn(ErrRestart)
	c.cancelFn = nil
	return true
}

// RequestSwitch replaces any pending switch and cancels the current session.
// cancelFn is cleared after firing so a later RequestRestart cannot fire it again.
func (c *SwitchController) RequestSwitch(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	select {
	case <-c.SwitchCh:
	default:
	}
	c.SwitchCh <- name
	if c.cancelFn != nil {
		c.cancelFn(ErrProfileSwitch)
		c.cancelFn = nil
	}
}
