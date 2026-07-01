package links

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type CollectionActionPayload struct {
	Actions []CollectionAction
}

type CollectionAction struct {
	Op             string
	RawText        string
	LinkedRecordID *uuid.UUID
	PartyID        *uuid.UUID
	ItemRef        string
	RiskRefText    string
	NormalizedText string
}

type CollectionValidationError struct {
	Field      string
	ReasonCode string
}

func (e *CollectionValidationError) Error() string {
	return "links: invalid collection action"
}

func (s *Store) ValidateCollectionPayloadsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, collections map[string]CollectionActionPayload) error {
	for fieldKey, payload := range collections {
		if err := s.ValidateCollectionPayloadTx(ctx, tx, incidentID, fieldKey, payload); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) ValidateCollectionPayloadTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, fieldKey string, payload CollectionActionPayload) error {
	for _, action := range payload.Actions {
		switch {
		case action.LinkedRecordID != nil:
			if err := validateCollectionTargetRecordTx(ctx, tx, incidentID, *action.LinkedRecordID, expectedCollectionTargetType(fieldKey), fieldKey); err != nil {
				return err
			}
		case action.PartyID != nil:
			if err := validateCollectionTargetRecordTx(ctx, tx, incidentID, *action.PartyID, "party", fieldKey); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Store) ApplyCollectionPayloadsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordID uuid.UUID, actorID uuid.UUID, collections map[string]CollectionActionPayload, now time.Time) (bool, error) {
	changed := false
	for fieldKey, payload := range collections {
		applied, err := s.ApplyCollectionPayloadTx(ctx, tx, incidentID, recordID, actorID, fieldKey, payload, now)
		if err != nil {
			return false, err
		}
		changed = changed || applied
	}
	return changed, nil
}

func (s *Store) ApplyCollectionPayloadTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordID uuid.UUID, actorID uuid.UUID, fieldKey string, payload CollectionActionPayload, now time.Time) (bool, error) {
	changed := false
	for _, action := range payload.Actions {
		switch action.Op {
		case "add_record_ref":
			applied, err := s.UpsertFieldReferenceTx(ctx, tx, incidentID, recordID, *action.LinkedRecordID, fieldKey, actorID, now)
			if err != nil {
				return false, err
			}
			changed = changed || applied
		case "remove_record_ref":
			dst, err := uuidFromItemRef(action.ItemRef, "record_ref:")
			if err != nil {
				return false, collectionValidationError(fieldKey)
			}
			applied, err := s.TombstoneFieldReferenceTx(ctx, tx, incidentID, recordID, dst, fieldKey, expectedCollectionTargetType(fieldKey), actorID, now)
			if err != nil {
				if errors.Is(err, ErrFieldReferenceNotFound) {
					return false, collectionValidationError(fieldKey)
				}
				return false, err
			}
			changed = changed || applied
		case "add_tag":
			applied, err := s.UpsertTagTx(ctx, tx, incidentID, recordID, action.RawText, action.NormalizedText, actorID, now)
			if err != nil {
				if errors.Is(err, ErrInvalidTag) {
					return false, collectionValidationError("note.tags")
				}
				return false, err
			}
			changed = changed || applied
		case "remove_tag":
			_, tagID, err := ParseRecordTagItemRef(action.ItemRef)
			if err != nil {
				return false, collectionValidationError(fieldKey)
			}
			applied, err := s.TombstoneTagTx(ctx, tx, incidentID, recordID, tagID, actorID, now)
			if err != nil {
				if errors.Is(err, ErrTagNotFound) {
					return false, collectionValidationError("note.tags")
				}
				return false, err
			}
			changed = changed || applied
		case "add_party_ref":
			applied, err := s.UpsertFieldReferenceTx(ctx, tx, incidentID, recordID, *action.PartyID, fieldKey, actorID, now)
			if err != nil {
				return false, err
			}
			changed = changed || applied
		case "remove_party_ref":
			dst, err := uuidFromItemRef(action.ItemRef, "party_ref:")
			if err != nil {
				return false, collectionValidationError(fieldKey)
			}
			applied, err := s.TombstoneFieldReferenceTx(ctx, tx, incidentID, recordID, dst, fieldKey, "party", actorID, now)
			if err != nil {
				if errors.Is(err, ErrFieldReferenceNotFound) {
					return false, collectionValidationError(fieldKey)
				}
				return false, err
			}
			changed = changed || applied
		case "add_risk_ref":
			applied, err := s.UpsertRiskRefTx(ctx, tx, incidentID, recordID, action.RiskRefText, action.NormalizedText, actorID, now)
			if err != nil {
				return false, err
			}
			changed = changed || applied
		case "remove_risk_ref":
			riskRefID, err := uuidFromItemRef(action.ItemRef, "risk_ref:")
			if err != nil {
				return false, collectionValidationError(fieldKey)
			}
			applied, err := s.TombstoneRiskRefTx(ctx, tx, incidentID, recordID, riskRefID, actorID, now)
			if err != nil {
				if errors.Is(err, ErrRiskRefNotFound) {
					return false, collectionValidationError("handoff.open_risk_refs")
				}
				return false, err
			}
			changed = changed || applied
		default:
			return false, collectionValidationError(fieldKey)
		}
	}
	return changed, nil
}

func validateCollectionTargetRecordTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordID uuid.UUID, expectedType string, field string) error {
	var exists bool
	if expectedType == "" {
		if err := tx.QueryRow(ctx, `
SELECT EXISTS (
    SELECT 1
      FROM records
     WHERE incident_id = $1
       AND record_id = $2
       AND deleted_at IS NULL
)
`, incidentID, recordID).Scan(&exists); err != nil {
			return fmt.Errorf("validate collection target: %w", err)
		}
	} else if err := tx.QueryRow(ctx, `
SELECT EXISTS (
    SELECT 1
      FROM records
     WHERE incident_id = $1
       AND record_id = $2
       AND record_type = $3
       AND deleted_at IS NULL
)
`, incidentID, recordID, expectedType).Scan(&exists); err != nil {
		return fmt.Errorf("validate collection target: %w", err)
	}
	if !exists {
		return collectionValidationError(field)
	}
	return nil
}

func expectedCollectionTargetType(fieldKey string) string {
	switch fieldKey {
	case "comm_log.decision_ids", "handoff.open_decision_ids", "status_review.open_decision_ids":
		return "decision"
	case "comm_log.action_task_ids", "handoff.open_task_ids", "status_review.blocked_task_ids", "lesson.follow_up_task_ids":
		return "task_request"
	case "status_review.pending_evidence_ids", "lesson.evidence_refs":
		return "evidence"
	case "task.linked_record_ids", "decision.support_refs", "decision.affected_record_ids",
		"finding.supporting_refs", "finding.contradictory_refs":
		return ""
	default:
		return ""
	}
}

func uuidFromItemRef(itemRef string, prefix string) (uuid.UUID, error) {
	if !strings.HasPrefix(itemRef, prefix) {
		return uuid.UUID{}, fmt.Errorf("invalid item ref")
	}
	value := strings.TrimPrefix(itemRef, prefix)
	parsed, err := uuid.Parse(value)
	if err != nil || parsed.String() != value {
		return uuid.UUID{}, fmt.Errorf("invalid item ref")
	}
	return parsed, nil
}

func collectionValidationError(field string) *CollectionValidationError {
	return &CollectionValidationError{Field: field, ReasonCode: "invalid_value"}
}
