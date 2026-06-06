package controller

import (
	"context"
	"errors"
	"sync"
)

var ErrProfileSwitch = errors.New("foundrymanager: profile switch requested")

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
