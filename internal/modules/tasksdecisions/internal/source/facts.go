package source

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/internal/policy"
)

func LoadTaskLifecycleStateTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (policy.TaskLifecycleState, error) {
	var state policy.TaskLifecycleState
	err := tx.QueryRow(ctx, `
SELECT status, NULLIF(blocked_reason, ''), completed_at, owner_user_id::text, created_at
  FROM task_requests
 WHERE record_id = $1
`, recordID).Scan(&state.Status, &state.BlockedReason, &state.CompletedAt, &state.OwnerUserID, &state.CreatedAt)
	if err != nil {
		return policy.TaskLifecycleState{}, &Error{Operation: "load task lifecycle state", Err: err}
	}
	return state, nil
}

func LoadDecisionMachineStateForUpdateTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (policy.DecisionMachineState, error) {
	var state policy.DecisionMachineState
	if err := tx.QueryRow(ctx, `
SELECT record_id, incident_id, status, owner_user_id::text, decided_at
  FROM decisions
 WHERE record_id = $1
 FOR UPDATE
`, recordID).Scan(&state.RecordID, &state.IncidentID, &state.Status, &state.OwnerUserID, &state.DecidedAt); err != nil {
		return policy.DecisionMachineState{}, &Error{Operation: "load decision machine state", Err: err}
	}
	if err := tx.QueryRow(ctx, `
SELECT COUNT(*), MIN(src_record_id::text)
  FROM active_record_links_v1
 WHERE incident_id = $1
   AND dst_record_id = $2
   AND link_type = 'supersedes'
`, state.IncidentID, recordID).Scan(&state.IncomingSupersedes, &state.IncomingSupersederID); err != nil {
		return policy.DecisionMachineState{}, &Error{Operation: "load decision incoming supersedes", Err: err}
	}
	if err := tx.QueryRow(ctx, `
SELECT COUNT(*), MIN(dst_record_id::text)
  FROM active_record_links_v1
 WHERE incident_id = $1
   AND src_record_id = $2
   AND link_type = 'supersedes'
`, state.IncidentID, recordID).Scan(&state.OutgoingSupersedes, &state.OutgoingTargetID); err != nil {
		return policy.DecisionMachineState{}, &Error{Operation: "load decision outgoing supersedes", Err: err}
	}
	return state, nil
}

func EnvelopeValidTx(ctx context.Context, tx pgx.Tx, recordID, incidentID uuid.UUID, recordType string) (bool, error) {
	var actualIncidentID uuid.UUID
	var actualType string
	var active bool
	err := tx.QueryRow(ctx, `
SELECT incident_id, record_type, deleted_at IS NULL
  FROM records
 WHERE record_id = $1
`, recordID).Scan(&actualIncidentID, &actualType, &active)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, &Error{Operation: "load portable envelope fact", Err: err}
	}
	return actualIncidentID == incidentID && actualType == recordType && active, nil
}

func SupersessionRelationsValidTx(ctx context.Context, tx pgx.Tx, recordID, incidentID uuid.UUID) (bool, error) {
	var invalid bool
	err := tx.QueryRow(ctx, `
SELECT EXISTS (
    SELECT 1
      FROM record_links link
      LEFT JOIN records source_record ON source_record.record_id = link.src_record_id
      LEFT JOIN records target_record ON target_record.record_id = link.dst_record_id
      LEFT JOIN decisions source_decision ON source_decision.record_id = link.src_record_id
      LEFT JOIN decisions target_decision ON target_decision.record_id = link.dst_record_id
     WHERE link.link_type = 'supersedes'
       AND link.deleted_at IS NULL
       AND (link.src_record_id = $1 OR link.dst_record_id = $1)
       AND (
           link.incident_id <> $2
           OR source_record.record_id IS NULL
           OR target_record.record_id IS NULL
           OR source_decision.record_id IS NULL
           OR target_decision.record_id IS NULL
           OR source_record.incident_id <> $2
           OR target_record.incident_id <> $2
           OR source_record.deleted_at IS NOT NULL
           OR target_record.deleted_at IS NOT NULL
           OR source_record.record_type <> 'decision'
           OR target_record.record_type <> 'decision'
           OR source_decision.status NOT IN ('approved', 'executed')
           OR target_decision.status NOT IN ('superseded', 'executed')
       )
)
`, recordID, incidentID).Scan(&invalid)
	if err != nil {
		return false, &Error{Operation: "load supersession relation facts", Err: err}
	}
	return !invalid, nil
}

func TargetValidTx(ctx context.Context, tx pgx.Tx, recordID, incidentID uuid.UUID, recordType string) (bool, error) {
	var valid bool
	err := tx.QueryRow(ctx, `
SELECT EXISTS (
    SELECT 1 FROM records
     WHERE record_id = $1
       AND incident_id = $2
       AND record_type = $3
       AND deleted_at IS NULL
)
`, recordID, incidentID, recordType).Scan(&valid)
	if err != nil {
		return false, &Error{Operation: "load target reference fact", Err: err}
	}
	return valid, nil
}

func TaskDecisionFieldLinkValidTx(ctx context.Context, tx pgx.Tx, taskID, decisionID, incidentID uuid.UUID) (bool, error) {
	var count int
	var linkedTarget sql.NullString
	if err := tx.QueryRow(ctx, `
SELECT count(*), min(dst_record_id::text)
  FROM active_record_links_v1
 WHERE incident_id = $1
   AND src_record_id = $2
   AND field_key = 'task.decision_record_id'
   AND link_type = 'references_record'
`, incidentID, taskID).Scan(&count, &linkedTarget); err != nil {
		return false, &Error{Operation: "load task decision link fact", Err: err}
	}
	if count == 0 {
		return true, nil
	}
	parsed, err := uuid.Parse(linkedTarget.String)
	return count == 1 && linkedTarget.Valid && err == nil && parsed == decisionID, nil
}

func OwnedLinksValidTx(ctx context.Context, tx pgx.Tx, sourceRecordID, incidentID uuid.UUID, recordType string) (bool, error) {
	var invalid bool
	err := tx.QueryRow(ctx, `
SELECT EXISTS (
    SELECT 1
      FROM record_links link
      LEFT JOIN records source_record ON source_record.record_id = link.src_record_id
      LEFT JOIN records target_record ON target_record.record_id = link.dst_record_id
     WHERE link.src_record_id = $1
       AND link.deleted_at IS NULL
       AND link.field_key IN (
           'task.linked_record_ids',
           'decision.support_refs',
           'decision.affected_record_ids'
       )
       AND (
           (link.field_key = 'task.linked_record_ids' AND link.link_type <> 'references_record')
           OR (link.field_key = 'decision.support_refs' AND link.link_type <> 'supported_by')
           OR (link.field_key = 'decision.affected_record_ids' AND link.link_type <> 'references_record')
           OR link.incident_id <> $2
           OR source_record.record_id IS NULL
           OR target_record.record_id IS NULL
           OR source_record.incident_id <> $2
           OR source_record.record_type <> $3
           OR source_record.deleted_at IS NOT NULL
           OR target_record.incident_id <> $2
           OR target_record.deleted_at IS NOT NULL
       )
)
`, sourceRecordID, incidentID, recordType).Scan(&invalid)
	if err != nil {
		return false, &Error{Operation: fmt.Sprintf("load %s owned-link facts", recordType), Err: err}
	}
	return !invalid, nil
}
