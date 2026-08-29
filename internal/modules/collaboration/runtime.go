package collaboration

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	privatestream "github.com/JochiRaider/cartulary/internal/modules/collaboration/internal/stream"
	"github.com/JochiRaider/cartulary/internal/modules/collaboration/protocol"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/platform/telemetry"
)

// Options contains the borrowed platform capabilities needed to construct the
// Collaboration runtime. The runtime owns its private process components, but
// it never closes the supplied database or transport.
type Options struct {
	Postgres                   postgres.DB
	AcceptSocket               protocol.AcceptSocket
	CheckBrowserOrigin         protocol.CheckBrowserOrigin
	ServiceVersion             string
	Now                        func() time.Time
	PublicationCatalog         *PublicationCatalog
	OnUnexpectedDispatcherLoss func()
}

type routeDependencies struct {
	acceptSocket   protocol.AcceptSocket
	checkOrigin    protocol.CheckBrowserOrigin
	serviceVersion string
}

// Runtime is Collaboration's application facade. Concrete live-session and
// dispatch components remain private to the owner.
type Runtime struct {
	hub                        *hub
	store                      *privatestream.PostgresStream
	dispatcher                 *privatestream.Dispatcher
	recordChanges              RecordChangedAppender
	jobProgress                JobProgressAppender
	extensionResourceChanges   ExtensionResourceChangedAppender
	routes                     routeDependencies
	onUnexpectedDispatcherLoss func()
	dispatcherLossOnce         sync.Once
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
	if options.OnUnexpectedDispatcherLoss == nil {
		return nil, errors.New("collaboration dispatcher loss callback is required")
	}
	if options.Now == nil {
		options.Now = func() time.Time { return time.Now().UTC() }
	}
	serviceVersion := strings.TrimSpace(options.ServiceVersion)
	if serviceVersion == "" {
		serviceVersion = telemetry.VersionUnknown
	}
	recordChanges, err := NewRecordChangedAppender(options.PublicationCatalog)
	if err != nil {
		return nil, err
	}
	liveHub := newHub(serviceVersion)
	store, err := privatestream.NewPostgresStream(options.Postgres)
	if err != nil {
		return nil, err
	}
	runtime := &Runtime{
		hub:                        liveHub,
		store:                      store,
		recordChanges:              recordChanges,
		jobProgress:                NewJobProgressAppender(),
		extensionResourceChanges:   NewExtensionResourceChangedAppender(),
		onUnexpectedDispatcherLoss: options.OnUnexpectedDispatcherLoss,
		routes: routeDependencies{
			acceptSocket:   options.AcceptSocket,
			checkOrigin:    options.CheckBrowserOrigin,
			serviceVersion: serviceVersion,
		},
	}
	runtime.dispatcher, err = privatestream.NewDispatcher(
		store,
		liveHub,
		options.Now,
		serviceVersion,
		runtime.reportUnexpectedDispatcherLoss,
	)
	if err != nil {
		return nil, err
	}
	return runtime, nil
}

func (runtime *Runtime) reportUnexpectedDispatcherLoss() {
	if runtime == nil || runtime.onUnexpectedDispatcherLoss == nil {
		return
	}
	runtime.dispatcherLossOnce.Do(runtime.onUnexpectedDispatcherLoss)
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

func (runtime *Runtime) RecordChanges() RecordChangedAppender {
	if runtime == nil {
		return nil
	}
	return runtime.recordChanges
}

func (runtime *Runtime) JobProgress() JobProgressAppender {
	if runtime == nil {
		return nil
	}
	return runtime.jobProgress
}

func (runtime *Runtime) ExtensionResourceChanges() ExtensionResourceChangedAppender {
	if runtime == nil {
		return nil
	}
	return runtime.extensionResourceChanges
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
