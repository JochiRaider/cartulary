package collaboration

import (
	"context"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration/protocol"
)

type socketTermination struct {
	reasonCode     string
	incidentClosed bool
}

func (e *socketTermination) Error() string { return "collaboration socket authority terminated" }

// Establishment writes do not run the live select loop. Check the same registered
// authority signals before each write, and latch a consumed terminal signal.
type socketAuthority struct {
	sessionRevocations  <-chan string
	incidentRevocations <-chan string
	incidentTerminals   <-chan string
	sessionExpiresAt    time.Time
	now                 func() time.Time
	terminal            *socketTermination
}

func (a *socketAuthority) check() error {
	if a.terminal != nil {
		return a.terminal
	}
	select {
	case reason := <-a.sessionRevocations:
		a.terminal = &socketTermination{reasonCode: reason}
	default:
	}
	if a.terminal == nil {
		select {
		case reason := <-a.incidentRevocations:
			a.terminal = &socketTermination{reasonCode: reason}
		default:
		}
	}
	if a.terminal == nil {
		select {
		case reason := <-a.incidentTerminals:
			a.terminal = &socketTermination{reasonCode: reason, incidentClosed: true}
		default:
		}
	}
	if a.terminal == nil && !a.sessionExpiresAt.After(a.now()) {
		a.terminal = &socketTermination{reasonCode: "session_expired"}
	}
	if a.terminal != nil {
		return a.terminal
	}
	return nil
}

type establishmentSocket struct {
	protocol.Socket
	authority *socketAuthority
}

func (s establishmentSocket) Write(ctx context.Context, kind protocol.MessageKind, payload []byte) error {
	if err := s.authority.check(); err != nil {
		return err
	}
	return s.Socket.Write(ctx, kind, payload)
}
