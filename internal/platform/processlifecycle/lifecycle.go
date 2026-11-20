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

// Admission states, one consistent snapshot of why the admission gate is open
// or closed. The HTTP gate maps each closed state to its service_unavailable
// reason and /readyz status (Core 01 §3.3.6.2, Core 04 REQ-04-078); the first
// applicable of fatal, stopping, paused and pending wins.
const (
	AdmissionStateOpen     = "open"
	AdmissionStatePending  = "pending"
	AdmissionStatePaused   = "paused"
	AdmissionStateStopping = "stopping"
	AdmissionStateFatal    = "fatal"
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
	fatalDone      chan struct{}
	publicationSet bool
	// paused records that uncertain lease ownership closed admission and the
	// original session has not yet proved continuous ownership.
	paused bool
	// stopping records an ordinary stop request (Core 04 REQ-04-146a).
	stopping bool
	// ownership is the live lease-ownership proof the gate opens under (Core
	// 04 REQ-04-145). It is evaluated under mu, so an open decision never
	// races a lease that became uncertain first.
	ownership func() bool
	// opened closes at the gate's first opening, the public publication
	// instant that serving-gated components wait for (EXT-REQ-214).
	opened       chan struct{}
	openedClosed bool
}

func New() *Controller {
	return &Controller{
		state:       StateStarting,
		fatalEvents: make(chan FatalSignal, 1),
		fatalDone:   make(chan struct{}),
		opened:      make(chan struct{}),
	}
}

// RequireOwnership installs the live proof that every application lease is
// held. Publication and restoration then open admission only while it holds.
// The first proof installed is kept.
func (c *Controller) RequireOwnership(owned func() bool) {
	if c == nil || owned == nil {
		return
	}
	c.mu.Lock()
	if c.ownership == nil {
		c.ownership = owned
	}
	c.mu.Unlock()
}

// Publish completes startup publication. The gate opens only while lease
// ownership holds; publication over uncertain ownership leaves admission
// paused until the original session proves ownership (Core 04 REQ-04-145).
func (c *Controller) Publish() error {
	if c == nil {
		return errors.New("extension_publication_failed")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.state != StateStarting || c.fatal != nil || c.publicationSet || c.stopping {
		return errors.New("extension_publication_failed")
	}
	c.publicationSet = true
	c.state = StateRunning
	c.openLocked()
	return nil
}

// openLocked opens admission unless uncertain ownership pauses it, and marks
// the first opening.
func (c *Controller) openLocked() bool {
	if c.paused || (c.ownership != nil && !c.ownership()) {
		c.paused = true
		c.admissionOpen = false
		return false
	}
	c.admissionOpen = true
	if !c.openedClosed {
		close(c.opened)
		c.openedClosed = true
	}
	return true
}

// CloseAdmission pauses admission while lease ownership is uncertain (Core 04
// REQ-04-145). It stays closed until RestoreAdmission or a fatal.
func (c *Controller) CloseAdmission() {
	if c == nil {
		return
	}
	c.mu.Lock()
	c.admissionOpen = false
	c.paused = true
	c.mu.Unlock()
}

// RestoreAdmission records the original session's proof of continuous
// ownership and reopens admission that uncertainty paused, when every lease
// is held. A process that is not running, is stopping or fatal, or never
// published stays closed.
func (c *Controller) RestoreAdmission() bool {
	if c == nil {
		return false
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.paused = false
	if c.state != StateRunning || c.fatal != nil || !c.publicationSet || c.stopping {
		return false
	}
	return c.openLocked()
}

// BeginStop enters ordinary shutdown at the stop request (Core 04
// REQ-04-146a): admission and readiness close before any drain and never
// reopen. A serving process quiesces; a starting one terminates. A fatal
// already selected keeps precedence, and a repeated stop changes nothing.
func (c *Controller) BeginStop() {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.fatal != nil || c.stopping {
		return
	}
	c.stopping = true
	c.admissionOpen = false
	switch c.state {
	case StateRunning:
		c.state = StateQuiescing
	case StateStarting:
		c.state = StateTerminating
	}
}

// AdmissionState reports the admission gate's state as one snapshot.
func (c *Controller) AdmissionState() string {
	if c == nil {
		return AdmissionStatePending
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	switch {
	case c.fatal != nil:
		return AdmissionStateFatal
	case c.stopping:
		return AdmissionStateStopping
	case c.admissionOpen:
		return AdmissionStateOpen
	case c.publicationSet && c.state == StateRunning:
		return AdmissionStatePaused
	default:
		return AdmissionStatePending
	}
}

// AdmissionPaused reports whether admission that startup publication
// completed is closed by uncertain lease ownership (Core 04 REQ-04-145).
// Before publication, admission is pending rather than paused.
func (c *Controller) AdmissionPaused() bool {
	return c.AdmissionState() == AdmissionStatePaused
}

// Opened closes when admission first opens.
func (c *Controller) Opened() <-chan struct{} {
	if c == nil {
		return nil
	}
	return c.opened
}

func (c *Controller) Fatal(reasonCode string) bool {
	if c == nil || reasonCode == "" {
		return false
	}
	c.mu.Lock()
	if c.fatal != nil {
		if reasonCode == "operation_cancellation_grace_exhausted" && c.state != StateExited {
			c.state = StateTerminating
		}
		c.mu.Unlock()
		return false
	}
	signal := FatalSignal{ReasonCode: reasonCode, ExitCode: 70}
	c.fatal = &signal
	close(c.fatalDone)
	c.admissionOpen = false
	// A fatal during an ordinary shutdown's drain keeps draining under the
	// same deadline (Core 04 REQ-04-146a).
	if (c.state == StateRunning || c.state == StateQuiescing) && reasonCode != "operation_cancellation_grace_exhausted" {
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

// FatalDone broadcasts the first fatal transition without consuming the
// process runner's single diagnostic signal.
func (c *Controller) FatalDone() <-chan struct{} {
	if c == nil {
		return nil
	}
	return c.fatalDone
}

// MarkTerminating closes admission for good once components start closing.
// Without a fatal it is ordinary shutdown, even when no stop request preceded
// it.
func (c *Controller) MarkTerminating() {
	if c == nil {
		return
	}
	c.mu.Lock()
	c.admissionOpen = false
	if c.fatal == nil {
		c.stopping = true
	}
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
	if c.fatal == nil {
		c.stopping = true
	}
	c.state = StateExited
	c.mu.Unlock()
}
