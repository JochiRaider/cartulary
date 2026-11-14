package source

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/internal/envelopefacts"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/internal/linkfacts"
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

func LoadDecisionMachineStateForUpdateTx(ctx context.Context, tx pgx.Tx, facts linkfacts.Capability, recordID uuid.UUID) (policy.DecisionMachineState, error) {
	var state policy.DecisionMachineState
	if err := tx.QueryRow(ctx, `
SELECT record_id, incident_id, status, owner_user_id::text, decided_at
  FROM decisions
 WHERE record_id = $1
 FOR UPDATE
`, recordID).Scan(&state.RecordID, &state.IncidentID, &state.Status, &state.OwnerUserID, &state.DecidedAt); err != nil {
		return policy.DecisionMachineState{}, &Error{Operation: "load decision machine state", Err: err}
	}
	return withActiveDecisionRelationsTx(ctx, tx, facts, state)
}

// LoadDecisionRestoreStateTx uses the coordinator-protected envelope and a
// non-waiting typed source lock. Deleted peers retain their intrinsic state.
func LoadDecisionRestoreStateTx(ctx context.Context, tx pgx.Tx, facts linkfacts.Capability, recordID, incidentID uuid.UUID) (policy.DecisionMachineState, bool, error) {
	var state policy.DecisionMachineState
	var deleted bool
	err := tx.QueryRow(ctx, `SELECT d.record_id,d.incident_id,d.status,d.owner_user_id::text,d.decided_at,r.deleted_at IS NOT NULL
FROM decisions d JOIN records r ON r.record_id=d.record_id AND r.incident_id=d.incident_id
WHERE d.record_id=$1 AND d.incident_id=$2 AND r.record_type='decision'
FOR UPDATE OF d NOWAIT`, recordID, incidentID).Scan(&state.RecordID, &state.IncidentID, &state.Status, &state.OwnerUserID, &state.DecidedAt, &deleted)
	if err != nil {
		return policy.DecisionMachineState{}, false, &Error{Operation: "load Decision restore state", Err: err}
	}
	state, err = withActiveDecisionRelationsTx(ctx, tx, facts, state)
	return state, deleted, err
}

func withActiveDecisionRelationsTx(ctx context.Context, tx pgx.Tx, facts linkfacts.Capability, state policy.DecisionMachineState) (policy.DecisionMachineState, error) {
	links, err := loadRecordLinkFactsTx(ctx, tx, facts, state.IncidentID, state.RecordID)
	if err != nil {
		return policy.DecisionMachineState{}, err
	}
	for _, fact := range links {
		if fact.LinkType != "supersedes" {
			continue
		}
		if fact.DestinationRecordID == state.RecordID {
			state.IncomingSupersedes++
			state.IncomingSupersederID = minimumUUIDText(state.IncomingSupersederID, fact.SourceRecordID)
		}
		if fact.SourceRecordID == state.RecordID {
			state.OutgoingSupersedes++
			state.OutgoingTargetID = minimumUUIDText(state.OutgoingTargetID, fact.DestinationRecordID)
		}
	}
	return state, nil
}

// PortableDecisionDeletedTx reads lifecycle only through the typed source's
// same-incident envelope. Earlier portable phases still validate every source value.
func PortableDecisionDeletedTx(ctx context.Context, tx pgx.Tx, recordID, incidentID uuid.UUID) (bool, error) {
	var deleted bool
	if err := tx.QueryRow(ctx, `SELECT r.deleted_at IS NOT NULL
FROM decisions d JOIN records r ON r.record_id=d.record_id AND r.incident_id=d.incident_id
WHERE d.record_id=$1 AND d.incident_id=$2 AND r.record_type='decision'`, recordID, incidentID).Scan(&deleted); err != nil {
		return false, &Error{Operation: "load portable Decision lifecycle", Err: err}
	}
	return deleted, nil
}

func SupersessionRelationsValidTx(ctx context.Context, tx pgx.Tx, facts linkfacts.Capability, recordID, incidentID uuid.UUID) (bool, error) {
	links, err := loadRecordLinkFactsTx(ctx, tx, facts, incidentID, recordID)
	if err != nil {
		return false, err
	}
	for _, fact := range links {
		if fact.LinkType != "supersedes" || (fact.SourceRecordID != recordID && fact.DestinationRecordID != recordID) {
			continue
		}
		valid, err := decisionSupersessionEndpointsValidTx(ctx, tx, fact, incidentID)
		if err != nil {
			return false, err
		}
		if !valid {
			return false, nil
		}
	}
	return true, nil
}

// PortableTaskReferenceTargetValidTx keeps deleted Task source portable without
// admitting dangling or untyped targets. Live Tasks still require active targets.
func PortableTaskReferenceTargetValidTx(ctx context.Context, tx pgx.Tx, taskID, targetID, incidentID uuid.UUID, recordType string) (bool, error) {
	var query string
	switch recordType {
	case "party":
		query = `SELECT EXISTS (
            SELECT 1 FROM task_requests t
            JOIN records task_record ON task_record.record_id=t.record_id AND task_record.incident_id=t.incident_id
            JOIN records target ON target.record_id=$1 AND target.incident_id=t.incident_id
            JOIN parties p ON p.record_id=target.record_id AND p.incident_id=target.incident_id
            WHERE t.record_id=$3 AND t.incident_id=$2 AND task_record.record_type='task_request'
              AND target.record_type='party' AND (task_record.deleted_at IS NOT NULL OR target.deleted_at IS NULL)
        )`
	case "decision":
		query = `SELECT EXISTS (
            SELECT 1 FROM task_requests t
            JOIN records task_record ON task_record.record_id=t.record_id AND task_record.incident_id=t.incident_id
            JOIN records target ON target.record_id=$1 AND target.incident_id=t.incident_id
            JOIN decisions d ON d.record_id=target.record_id AND d.incident_id=target.incident_id
            WHERE t.record_id=$3 AND t.incident_id=$2 AND task_record.record_type='task_request'
              AND target.record_type='decision' AND (task_record.deleted_at IS NOT NULL OR target.deleted_at IS NULL)
        )`
	default:
		return false, nil
	}
	var valid bool
	if err := tx.QueryRow(ctx, query, targetID, incidentID, taskID).Scan(&valid); err != nil {
		return false, &Error{Operation: "load portable Task reference fact", Err: err}
	}
	return valid, nil
}

// TaskDecisionFieldLinkValidTx compares the nullable convenience scalar with
// retained non-deleted owner-field bindings, independently of endpoint visibility.
func TaskDecisionFieldLinkValidTx(ctx context.Context, tx pgx.Tx, facts linkfacts.FieldReferenceReader, taskID uuid.UUID, decisionID *uuid.UUID, incidentID uuid.UUID) (bool, error) {
	if facts == nil {
		return false, &Error{Operation: "load Task Decision bindings", Err: errors.New("retained field references are required")}
	}
	bindings, err := facts.LoadFieldReferencesTx(ctx, tx, incidentID, taskID, policy.TaskDecisionRecordField)
	if err != nil {
		return false, &Error{Operation: "load Task Decision bindings", Err: err}
	}
	return len(bindings) == 0 || (len(bindings) == 1 && decisionID != nil && bindings[0].LinkType == "references_record" && bindings[0].DestinationRecordID == *decisionID), nil
}

func OwnedLinksValidTx(ctx context.Context, tx pgx.Tx, facts linkfacts.Capability, envelopes envelopefacts.Reader, sourceRecordID, incidentID uuid.UUID, recordType string) (bool, error) {
	links, err := loadRecordLinkFactsTx(ctx, tx, facts, incidentID, sourceRecordID)
	if err != nil {
		return false, err
	}
	for _, fact := range links {
		expectedType, owned := ownedLinkType(fact)
		if !owned || fact.SourceRecordID != sourceRecordID {
			continue
		}
		if fact.LinkType != expectedType {
			return false, nil
		}
		valid, err := ownedLinkEndpointsValidTx(ctx, tx, envelopes, fact, incidentID, recordType)
		if err != nil {
			return false, err
		}
		if !valid {
			return false, nil
		}
	}
	return true, nil
}

func loadRecordLinkFactsTx(ctx context.Context, tx pgx.Tx, facts linkfacts.Capability, incidentID, recordID uuid.UUID) ([]linkfacts.Fact, error) {
	if facts == nil {
		return nil, &Error{Operation: "load Links facts", Err: errors.New("links fact capability is required")}
	}
	values, err := facts.LoadRecordLinkFactsTx(ctx, tx, incidentID, recordID)
	if err != nil {
		return nil, &Error{Operation: "load Links facts", Err: err}
	}
	if values == nil {
		return []linkfacts.Fact{}, nil
	}
	return values, nil
}

func minimumUUIDText(current sql.NullString, candidate uuid.UUID) sql.NullString {
	text := candidate.String()
	if !current.Valid || text < current.String {
		return sql.NullString{String: text, Valid: true}
	}
	return current
}

func decisionSupersessionEndpointsValidTx(ctx context.Context, tx pgx.Tx, fact linkfacts.Fact, incidentID uuid.UUID) (bool, error) {
	var sourceIncidentID, targetIncidentID uuid.UUID
	var sourceType, targetType, sourceStatus, targetStatus string
	var sourceActive, targetActive bool
	err := tx.QueryRow(ctx, `
SELECT source_record.incident_id, source_record.record_type, source_record.deleted_at IS NULL,
       target_record.incident_id, target_record.record_type, target_record.deleted_at IS NULL,
       source_decision.status, target_decision.status
  FROM records source_record
  JOIN records target_record ON target_record.record_id = $2
  JOIN decisions source_decision ON source_decision.record_id = source_record.record_id
  JOIN decisions target_decision ON target_decision.record_id = target_record.record_id
 WHERE source_record.record_id = $1
`, fact.SourceRecordID, fact.DestinationRecordID).Scan(
		&sourceIncidentID, &sourceType, &sourceActive,
		&targetIncidentID, &targetType, &targetActive,
		&sourceStatus, &targetStatus,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, &Error{Operation: "load supersession relation facts", Err: err}
	}
	return sourceIncidentID == incidentID && targetIncidentID == incidentID &&
		sourceType == "decision" && targetType == "decision" && sourceActive && targetActive &&
		(sourceStatus == "approved" || sourceStatus == "executed") &&
		(targetStatus == "superseded" || targetStatus == "executed"), nil
}

func ownedLinkType(fact linkfacts.Fact) (string, bool) {
	if !fact.HasFieldKey {
		return "", false
	}
	switch fact.FieldKey {
	case "task.linked_record_ids", "decision.affected_record_ids":
		return "references_record", true
	case "decision.support_refs":
		return "supported_by", true
	default:
		return "", false
	}
}

func ownedLinkEndpointsValidTx(ctx context.Context, tx pgx.Tx, envelopes envelopefacts.Reader, fact linkfacts.Fact, incidentID uuid.UUID, recordType string) (bool, error) {
	source, found, err := envelopes.LoadEnvelopeFactTx(ctx, tx, fact.SourceRecordID)
	if err != nil {
		return false, &Error{Operation: fmt.Sprintf("load %s owned-link source fact", recordType), Err: err}
	}
	if !found || source.IncidentID != incidentID || source.RecordType != recordType || source.Deleted {
		return false, nil
	}
	target, found, err := envelopes.LoadEnvelopeFactTx(ctx, tx, fact.DestinationRecordID)
	if err != nil {
		return false, &Error{Operation: fmt.Sprintf("load %s owned-link target fact", recordType), Err: err}
	}
	return found && target.IncidentID == incidentID && !target.Deleted, nil
}
