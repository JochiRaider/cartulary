package incidents

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/platform/administrativeaudit"
)

type auditEvent struct {
	ActorUserID  *uuid.UUID
	TargetUserID *uuid.UUID
	IncidentID   *uuid.UUID
	EventSource  string
	EventKind    string
	ReasonCode   *string
	ClientTxnID  *string
	RequestID    *string
	BeforeJSON   any
	AfterJSON    any
	PublicSource string
}

func insertAuditEvent(ctx context.Context, tx pgx.Tx, event auditEvent) (uuid.UUID, error) {
	occurredAt := time.Now().UTC()
	raw := administrativeaudit.RawEvent{
		ActorUserID:  event.ActorUserID,
		TargetUserID: event.TargetUserID,
		IncidentID:   event.IncidentID,
		EventSource:  event.EventSource,
		EventKind:    event.EventKind,
		ReasonCode:   event.ReasonCode,
		ClientTxnID:  event.ClientTxnID,
		RequestID:    event.RequestID,
		Before:       event.BeforeJSON,
		After:        event.AfterJSON,
		OccurredAt:   occurredAt,
	}
	actionCode, changes, projected := membershipAuditProjection(event)
	if !projected {
		eventID, err := administrativeaudit.AppendRawTx(ctx, tx, raw)
		if err != nil {
			return uuid.Nil, fmt.Errorf("insert incident audit event: %w", err)
		}
		return eventID, nil
	}
	if event.IncidentID == nil || event.ActorUserID == nil || event.TargetUserID == nil {
		return uuid.Nil, errors.New("insert incident membership audit event: projection identifiers are incomplete")
	}
	source := event.PublicSource
	if source == "" {
		source = administrativeaudit.SourceAPI
	}
	targetID := event.TargetUserID.String()
	eventID, err := administrativeaudit.AppendTx(ctx, tx, raw, administrativeaudit.Event{
		ScopeKind:   administrativeaudit.ScopeIncident,
		ScopeID:     event.IncidentID,
		OccurredAt:  occurredAt,
		ActorKind:   administrativeaudit.ActorUser,
		ActorUserID: event.ActorUserID,
		Source:      source,
		ActionCode:  actionCode,
		TargetKind:  administrativeaudit.TargetIncidentMembership,
		TargetID:    &targetID,
		Changes:     changes,
		ReasonCode:  event.ReasonCode,
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("insert incident audit event: %w", err)
	}
	return eventID, nil
}

func membershipAuditProjection(event auditEvent) (string, []administrativeaudit.Change, bool) {
	beforeRole := membershipRole(event.BeforeJSON)
	afterRole := membershipRole(event.AfterJSON)
	switch event.EventKind {
	case "incident_membership_created":
		return administrativeaudit.ActionMembershipCreated, []administrativeaudit.Change{
			administrativeaudit.Visible("role", nil, afterRole),
		}, true
	case "incident_membership_updated":
		return administrativeaudit.ActionMembershipRoleChanged, []administrativeaudit.Change{
			administrativeaudit.Visible("role", beforeRole, afterRole),
		}, true
	case "incident_membership_deleted":
		return administrativeaudit.ActionMembershipDeleted, []administrativeaudit.Change{
			administrativeaudit.Visible("role", beforeRole, nil),
		}, true
	default:
		return "", nil, false
	}
}

func membershipRole(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		return typed["role"]
	case map[string]string:
		return typed["role"]
	default:
		payload, err := json.Marshal(value)
		if err != nil {
			return nil
		}
		var resource struct {
			Role any `json:"role"`
		}
		if err := json.Unmarshal(payload, &resource); err != nil {
			return nil
		}
		return resource.Role
	}
}
