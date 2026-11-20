package collaboration

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration/protocol"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

type socketTermination struct {
	reasonCode     string
	incidentClosed bool
}

func (e *socketTermination) Error() string { return "collaboration socket authority terminated" }

// errSessionUnverified reports that the socket could not re-read its session at
// the session's expiry point. That is no proof of expiry, so the connection
// ends without revoking the session or reporting a revocation.
var errSessionUnverified = errors.New("collaboration socket could not re-read its session")

// Establishment writes do not run the live select loop. Check the same registered
// authority signals before each write, and latch a consumed terminal signal.
type socketAuthority struct {
	sessionRevocations  <-chan string
	incidentRevocations <-chan string
	incidentTerminals   <-chan string
	sessionExpiresAt    time.Time
	// rereadSession returns the session's current expiry through the central
	// authenticator, which never slides it. An authenticator that finds the
	// session expired revokes it and publishes session_expired to the
	// session's sockets.
	rereadSession func() (time.Time, *httpapi.APIError)
	now           func() time.Time
	terminal      *socketTermination
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
	if a.terminal == nil {
		if err := a.checkSessionExpiry(); err != nil {
			return err
		}
	}
	if a.terminal != nil {
		return a.terminal
	}
	return nil
}

// checkSessionExpiry ends the socket only for a session that has really
// expired (Core 01 REQ-01-277, Core 04 REQ-04-009). Before the expiry the
// socket last saw, the session is live. At that point the socket re-reads the
// session, because qualifying HTTP activity may have slid it: a live session
// re-arms the expiry. A session the authenticator found expired ends with the
// reason it published to this socket, session_expired; any other refused
// session ends as session_revoked. A session that cannot be read returns
// errSessionUnverified.
func (a *socketAuthority) checkSessionExpiry() error {
	if a.terminal != nil {
		return a.terminal
	}
	if a.sessionExpiresAt.After(a.now()) {
		return nil
	}
	expiresAt, apiErr := a.rereadSession()
	if apiErr == nil {
		a.sessionExpiresAt = expiresAt
		return nil
	}
	if apiErr.Status != http.StatusUnauthorized {
		return errSessionUnverified
	}
	reason := "session_revoked"
	select {
	case published := <-a.sessionRevocations:
		reason = published
	default:
	}
	a.terminal = &socketTermination{reasonCode: reason}
	return a.terminal
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
