package processlifecycle

import (
	"errors"
	"sync"
)

type State string

const (
	StateStarting    State = "starting"
	StateRunning     State = "running"
	StateQuiescing   State = "quiescing"
	StateTerminating State = "terminating"
	StateExited      State = "exited"
)

type FatalSignal struct {
	ReasonCode string
	ExitCode   int
}

type Controller struct {
	mu             sync.RWMutex
	state          State
	admissionOpen  bool
	fatal          *FatalSignal
	fatalEvents    chan FatalSignal
	publicationSet bool
}

func New() *Controller {
	return &Controller{state: StateStarting, fatalEvents: make(chan FatalSignal, 1)}
}

func (c *Controller) Publish() error {
	if c == nil {
		return errors.New("extension_publication_failed")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.state != StateStarting || c.fatal != nil || c.publicationSet {
		return errors.New("extension_publication_failed")
	}
	c.publicationSet = true
	c.admissionOpen = true
	c.state = StateRunning
	return nil
}

func (c *Controller) CloseAdmission() {
	if c == nil {
		return
	}
	c.mu.Lock()
	c.admissionOpen = false
	c.mu.Unlock()
}

func (c *Controller) RestoreAdmission() bool {
	if c == nil {
		return false
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.state != StateRunning || c.fatal != nil || !c.publicationSet {
		return false
	}
	c.admissionOpen = true
	return true
}

func (c *Controller) Fatal(reasonCode string) bool {
	if c == nil || reasonCode == "" {
		return false
	}
	c.mu.Lock()
	if c.fatal != nil {
		c.mu.Unlock()
		return false
	}
	signal := FatalSignal{ReasonCode: reasonCode, ExitCode: 70}
	c.fatal = &signal
	c.admissionOpen = false
	if c.state == StateRunning {
		c.state = StateQuiescing
	} else {
		c.state = StateTerminating
	}
	c.mu.Unlock()
	c.fatalEvents <- signal
	return true
}

func (c *Controller) AdmissionOpen() bool {
	if c == nil {
		return false
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.admissionOpen
}

func (c *Controller) State() State {
	if c == nil {
		return StateExited
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.state
}

func (c *Controller) FatalReason() string {
	if c == nil {
		return ""
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.fatal == nil {
		return ""
	}
	return c.fatal.ReasonCode
}

func (c *Controller) FatalEvents() <-chan FatalSignal {
	if c == nil {
		return nil
	}
	return c.fatalEvents
}

func (c *Controller) MarkTerminating() {
	if c == nil {
		return
	}
	c.mu.Lock()
	c.admissionOpen = false
	if c.state != StateExited {
		c.state = StateTerminating
	}
	c.mu.Unlock()
}

func (c *Controller) MarkExited() {
	if c == nil {
		return
	}
	c.mu.Lock()
	c.admissionOpen = false
	c.state = StateExited
	c.mu.Unlock()
}
