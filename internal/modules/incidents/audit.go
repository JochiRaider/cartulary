package incidents

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/platform/administrativeaudit"
)

type auditEvent struct {
	actorUserID  *uuid.UUID
	targetUserID *uuid.UUID
	incidentID   *uuid.UUID
	kind         auditEventKind
	source       auditEventSource
	reasonCode   *string
	clientTxnID  *string
	requestID    *string
	beforeJSON   any
	afterJSON    any
	roles        auditRoleFacts
	occurredAt   time.Time
}

type auditEventKind uint8

const (
	auditIncidentCreated auditEventKind = iota + 1
	auditIncidentUpdated
	auditIncidentClosed
	auditIncidentReopened
	auditMembershipCreated
	auditMembershipUpdated
	auditMembershipDeleted
)

func (kind auditEventKind) rawValue() string {
	switch kind {
	case auditIncidentCreated:
		return "incident_created"
	case auditIncidentUpdated:
		return "incident_updated"
	case auditIncidentClosed:
		return "incident_close"
	case auditIncidentReopened:
		return "incident_reopen"
	case auditMembershipCreated:
		return "incident_membership_created"
	case auditMembershipUpdated:
		return "incident_membership_updated"
	case auditMembershipDeleted:
		return "incident_membership_deleted"
	default:
		return ""
	}
}

func (kind auditEventKind) membershipAction() (string, bool) {
	switch kind {
	case auditMembershipCreated:
		return administrativeaudit.ActionMembershipCreated, true
	case auditMembershipUpdated:
		return administrativeaudit.ActionMembershipRoleChanged, true
	case auditMembershipDeleted:
		return administrativeaudit.ActionMembershipDeleted, true
	default:
		return "", false
	}
}

type auditEventSource uint8

const (
	auditSourceAPI auditEventSource = iota + 1
	auditSourceSystem
)

func (source auditEventSource) publicValue() string {
	switch source {
	case auditSourceAPI:
		return administrativeaudit.SourceAPI
	case auditSourceSystem:
		return administrativeaudit.SourceSystem
	default:
		return ""
	}
}

type auditRoleFacts struct {
	before *string
	after  *string
}

func auditRole(role string) *string {
	return &role
}

func insertAuditEvent(ctx context.Context, tx pgx.Tx, event auditEvent) (uuid.UUID, error) {
	eventKind := event.kind.rawValue()
	publicSource := event.source.publicValue()
	if eventKind == "" || publicSource == "" || event.occurredAt.IsZero() {
		return uuid.Nil, errors.New("insert incident audit event: closed facts are incomplete")
	}
	occurredAt := event.occurredAt.UTC()
	raw := administrativeaudit.RawEvent{
		ActorUserID:  event.actorUserID,
		TargetUserID: event.targetUserID,
		IncidentID:   event.incidentID,
		EventSource:  "incidents",
		EventKind:    eventKind,
		ReasonCode:   event.reasonCode,
		ClientTxnID:  event.clientTxnID,
		RequestID:    event.requestID,
		Before:       event.beforeJSON,
		After:        event.afterJSON,
		OccurredAt:   occurredAt,
	}
	actionCode, projected := event.kind.membershipAction()
	if !projected {
		if event.roles.before != nil || event.roles.after != nil {
			return uuid.Nil, errors.New("insert incident audit event: non-membership event has role facts")
		}
		eventID, err := administrativeaudit.AppendRawTx(ctx, tx, raw)
		if err != nil {
			return uuid.Nil, fmt.Errorf("insert incident audit event: %w", err)
		}
		return eventID, nil
	}
	if event.incidentID == nil || event.actorUserID == nil || event.targetUserID == nil {
		return uuid.Nil, errors.New("insert incident membership audit event: projection identifiers are incomplete")
	}
	changes, err := membershipAuditChanges(event.kind, event.roles)
	if err != nil {
		return uuid.Nil, err
	}
	targetID := event.targetUserID.String()
	eventID, err := administrativeaudit.AppendTx(ctx, tx, raw, administrativeaudit.Event{
		ScopeKind:   administrativeaudit.ScopeIncident,
		ScopeID:     event.incidentID,
		OccurredAt:  occurredAt,
		ActorKind:   administrativeaudit.ActorUser,
		ActorUserID: event.actorUserID,
		Source:      publicSource,
		ActionCode:  actionCode,
		TargetKind:  administrativeaudit.TargetIncidentMembership,
		TargetID:    &targetID,
		Changes:     changes,
		ReasonCode:  event.reasonCode,
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("insert incident audit event: %w", err)
	}
	return eventID, nil
}

func membershipAuditChanges(kind auditEventKind, roles auditRoleFacts) ([]administrativeaudit.Change, error) {
	if roles.before != nil && !slices.Contains(membershipRoles, *roles.before) {
		return nil, errors.New("insert incident membership audit event: invalid before role")
	}
	if roles.after != nil && !slices.Contains(membershipRoles, *roles.after) {
		return nil, errors.New("insert incident membership audit event: invalid after role")
	}
	switch kind {
	case auditMembershipCreated:
		if roles.before != nil || roles.after == nil {
			return nil, errors.New("insert incident membership-created audit event: invalid role facts")
		}
		return []administrativeaudit.Change{administrativeaudit.Visible("role", nil, *roles.after)}, nil
	case auditMembershipUpdated:
		if roles.before == nil || roles.after == nil {
			return nil, errors.New("insert incident membership-updated audit event: invalid role facts")
		}
		return []administrativeaudit.Change{administrativeaudit.Visible("role", *roles.before, *roles.after)}, nil
	case auditMembershipDeleted:
		if roles.before == nil || roles.after != nil {
			return nil, errors.New("insert incident membership-deleted audit event: invalid role facts")
		}
		return []administrativeaudit.Change{administrativeaudit.Visible("role", *roles.before, nil)}, nil
	default:
		return nil, errors.New("insert incident membership audit event: unmapped event kind")
	}
}
