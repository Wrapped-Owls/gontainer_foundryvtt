// Package controller coordinates profile switching between the dashboard HTTP
// handler and the manager's process loop.
package controller

import (
	"context"
	"errors"
	"sync"
)

// ErrProfileSwitch is injected as the cause when a profile switch cancels the
// current session context. Callers can use errors.Is to distinguish a switch
// from a normal shutdown.
var ErrProfileSwitch = errors.New("foundrymanager: profile switch requested")

// SwitchController coordinates the handoff between the dashboard HTTP handler
// and the manager's process loop. All methods are safe for concurrent use.
type SwitchController struct {
	mu       sync.Mutex
	cancelFn context.CancelCauseFunc
	current  string
	SwitchCh chan string
}

// New returns an initialised SwitchController.
func New() *SwitchController {
	return &SwitchController{SwitchCh: make(chan string, 1)}
}

// SetCancel stores the cancel function for the current profile session.
func (c *SwitchController) SetCancel(fn context.CancelCauseFunc) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cancelFn = fn
}

// SetActive records the name of the currently active profile.
func (c *SwitchController) SetActive(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.current = name
}

// Active returns the name of the currently active profile.
func (c *SwitchController) Active() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.current
}

// RequestSwitch queues a profile switch and cancels the current session.
// If a previous switch is still pending it is replaced.
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
	}
}
