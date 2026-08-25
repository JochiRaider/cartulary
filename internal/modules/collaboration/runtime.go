package collaboration

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	privatestream "github.com/JochiRaider/cartulary/internal/modules/collaboration/internal/stream"
	"github.com/JochiRaider/cartulary/internal/modules/collaboration/protocol"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

// Options contains the borrowed platform capabilities needed to construct the
// Collaboration runtime. The runtime owns its private process components, but
// it never closes the supplied database or transport.
type Options struct {
	Postgres           postgres.DB
	AcceptSocket       protocol.AcceptSocket
	CheckBrowserOrigin protocol.CheckBrowserOrigin
	ServiceVersion     string
	Now                func() time.Time
	PublicationCatalog *PublicationCatalog
}

// Runtime is Collaboration's application facade. Concrete live-session and
// dispatch components remain private to the owner.
type Runtime struct {
	hub          *hub
	store        *privatestream.PostgresStream
	dispatcher   *privatestream.Dispatcher
	publications PublicationAppender
	options      Options
}

func NewRuntime(options Options) (*Runtime, error) {
	if options.Postgres == nil {
		return nil, errors.New("collaboration PostgreSQL dependency is required")
	}
	if options.AcceptSocket == nil {
		return nil, errors.New("collaboration WebSocket accept dependency is required")
	}
	if options.CheckBrowserOrigin == nil {
		return nil, errors.New("collaboration WebSocket Origin dependency is required")
	}
	if options.Now == nil {
		options.Now = func() time.Time { return time.Now().UTC() }
	}
	publications, err := NewPublicationAppender(options.PublicationCatalog)
	if err != nil {
		return nil, err
	}
	liveHub := newHub()
	liveHub.ConfigureTelemetry(options.ServiceVersion)
	store := privatestream.NewPostgresStream(options.Postgres, options.Now)
	return &Runtime{
		hub:          liveHub,
		store:        store,
		dispatcher:   privatestream.NewDispatcher(store, liveHub, options.Now),
		publications: publications,
		options:      options,
	}, nil
}

func (runtime *Runtime) Start(ctx context.Context) error {
	if runtime == nil || runtime.dispatcher == nil {
		return errors.New("collaboration runtime is not configured")
	}
	return runtime.dispatcher.Start(ctx)
}

func (runtime *Runtime) Close(ctx context.Context) error {
	if runtime == nil || runtime.dispatcher == nil {
		return nil
	}
	return runtime.dispatcher.Close(ctx)
}

func (runtime *Runtime) Publications() PublicationAppender {
	if runtime == nil {
		return nil
	}
	return runtime.publications
}

// RevokeSession is the narrow Auth-owned consumer capability.
func (runtime *Runtime) RevokeSession(sessionID uuid.UUID, reasonCode string) {
	if runtime == nil || runtime.hub == nil {
		return
	}
	runtime.hub.RevokeSession(sessionID, reasonCode)
}

// NotifyIncidentClosed is the narrow Incidents-owned terminal-effect
// capability. The effect identity is accepted for the owner contract; live
// fan-out remains idempotent process-local state.
func (runtime *Runtime) NotifyIncidentClosed(_ context.Context, _ uuid.UUID, incidentID uuid.UUID) {
	if runtime == nil || runtime.hub == nil {
		return
	}
	runtime.hub.TerminateIncident(incidentID, protocol.IncidentTerminalClosed)
}

// NotifyIncidentMembershipRevoked is the narrow Incidents-owned membership
// effect capability.
func (runtime *Runtime) NotifyIncidentMembershipRevoked(_ context.Context, _ uuid.UUID, incidentID uuid.UUID, userID uuid.UUID) {
	if runtime == nil || runtime.hub == nil {
		return
	}
	runtime.hub.RevokeIncidentAccess(incidentID, userID)
}

// IncidentEventObserver is a semantic observation capability used by
// application-level tests. It exposes neither hub identity nor lifecycle.
type IncidentEventObserver interface {
	SubscribeIncident(uuid.UUID, int) (<-chan protocol.Message, func())
}

type incidentEventObserver struct {
	hub *hub
}

func (observer incidentEventObserver) SubscribeIncident(incidentID uuid.UUID, buffer int) (<-chan protocol.Message, func()) {
	return observer.hub.SubscribeIncident(incidentID, buffer)
}

func (runtime *Runtime) IncidentEvents() IncidentEventObserver {
	if runtime == nil {
		return incidentEventObserver{}
	}
	return incidentEventObserver{hub: runtime.hub}
}
