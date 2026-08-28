package source

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

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
	links, err := loadRecordLinkFactsTx(ctx, tx, facts, state.IncidentID, recordID)
	if err != nil {
		return policy.DecisionMachineState{}, err
	}
	for _, fact := range links {
		if fact.LinkType != "supersedes" {
			continue
		}
		if fact.DestinationRecordID == recordID {
			state.IncomingSupersedes++
			state.IncomingSupersederID = minimumUUIDText(state.IncomingSupersederID, fact.SourceRecordID)
		}
		if fact.SourceRecordID == recordID {
			state.OutgoingSupersedes++
			state.OutgoingTargetID = minimumUUIDText(state.OutgoingTargetID, fact.DestinationRecordID)
		}
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

func TaskDecisionFieldLinkValidTx(ctx context.Context, tx pgx.Tx, facts linkfacts.Capability, taskID, decisionID, incidentID uuid.UUID) (bool, error) {
	links, err := loadRecordLinkFactsTx(ctx, tx, facts, incidentID, taskID)
	if err != nil {
		return false, err
	}
	count := 0
	var linkedTarget uuid.UUID
	for _, fact := range links {
		if fact.SourceRecordID == taskID && fact.HasFieldKey && fact.FieldKey == policy.TaskDecisionRecordField && fact.LinkType == "references_record" {
			count++
			linkedTarget = fact.DestinationRecordID
		}
	}
	return count == 0 || (count == 1 && linkedTarget == decisionID), nil
}

func OwnedLinksValidTx(ctx context.Context, tx pgx.Tx, facts linkfacts.Capability, sourceRecordID, incidentID uuid.UUID, recordType string) (bool, error) {
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
		valid, err := ownedLinkEndpointsValidTx(ctx, tx, fact, incidentID, recordType)
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

func ownedLinkEndpointsValidTx(ctx context.Context, tx pgx.Tx, fact linkfacts.Fact, incidentID uuid.UUID, recordType string) (bool, error) {
	var sourceIncidentID, targetIncidentID uuid.UUID
	var sourceType string
	var sourceActive, targetActive bool
	err := tx.QueryRow(ctx, `
SELECT source_record.incident_id, source_record.record_type, source_record.deleted_at IS NULL,
       target_record.incident_id, target_record.deleted_at IS NULL
  FROM records source_record
  JOIN records target_record ON target_record.record_id = $2
 WHERE source_record.record_id = $1
`, fact.SourceRecordID, fact.DestinationRecordID).Scan(
		&sourceIncidentID, &sourceType, &sourceActive, &targetIncidentID, &targetActive,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, &Error{Operation: fmt.Sprintf("load %s owned-link facts", recordType), Err: err}
	}
	return sourceIncidentID == incidentID && targetIncidentID == incidentID &&
		sourceType == recordType && sourceActive && targetActive, nil
}
