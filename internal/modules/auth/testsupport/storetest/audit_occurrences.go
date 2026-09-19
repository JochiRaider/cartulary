package storetest

import (
	"context"
	"errors"

	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/google/uuid"
)

// AuditOccurrenceSelector selects committed source-owner audit records. An exact
// operation correlation is mandatory; callers cannot accidentally count a whole
// incident or another actor's activity as evidence for their operation.
type AuditOccurrenceSelector struct {
	ActorID, IncidentID                           uuid.UUID
	EventCode, ResourceID, ClientTxnID, RequestID string
}

func CountAuditOccurrences(ctx context.Context, db postgres.DB, s AuditOccurrenceSelector) (int, error) {
	if s.ActorID == uuid.Nil || s.IncidentID == uuid.Nil || s.EventCode == "" || (s.ClientTxnID == "" && s.RequestID == "") {
		return 0, errors.New("audit occurrence selector requires actor, incident, event and operation correlation")
	}
	var count int
	err := db.QueryRow(ctx, `SELECT count(*) FROM deployment_admin_audit_events WHERE actor_user_id=$1 AND incident_id=$2 AND event_kind=$3 AND ($4='' OR client_txn_id=$4) AND ($5='' OR request_id=$5) AND ($6='' OR after_json->>'network_flow.audit_resource_id'=$6)`, s.ActorID, s.IncidentID, s.EventCode, s.ClientTxnID, s.RequestID, s.ResourceID).Scan(&count)
	return count, err
}
