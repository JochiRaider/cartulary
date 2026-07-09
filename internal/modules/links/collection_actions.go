package links

import (
	"context"
	"errors"
	"fmt"
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
	NormalizedText string
}

type CollectionFieldPolicy struct {
	FieldKey           string
	LinkType           string
	ExpectedTargetType string
	AllowRecordRefs    bool
	AllowPartyRefs     bool
	AllowTags          bool
}

type CollectionValidationError struct {
	Field      string
	ReasonCode string
}

func (e *CollectionValidationError) Error() string {
	return "links: invalid collection action"
}

func (s *Store) ValidateCollectionPayloadsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, policies map[string]CollectionFieldPolicy, collections map[string]CollectionActionPayload) error {
	for fieldKey, payload := range collections {
		policy, err := collectionPolicyForField(fieldKey, policies)
		if err != nil {
			return err
		}
		if err := s.ValidateCollectionPayloadTx(ctx, tx, incidentID, policy, payload); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) ValidateCollectionPayloadTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, policy CollectionFieldPolicy, payload CollectionActionPayload) error {
	if err := validateCollectionPolicy(policy); err != nil {
		return err
	}
	for _, action := range payload.Actions {
		switch action.Op {
		case "add_record_ref":
			if !policy.AllowRecordRefs || action.LinkedRecordID == nil {
				return collectionValidationError(policy.FieldKey)
			}
			if err := validateCollectionTargetRecordTx(ctx, tx, incidentID, *action.LinkedRecordID, policy.ExpectedTargetType, policy.FieldKey); err != nil {
				return err
			}
		case "remove_record_ref":
			if !policy.AllowRecordRefs {
				return collectionValidationError(policy.FieldKey)
			}
			if _, err := ParseRecordRefItemRef(action.ItemRef); err != nil {
				return collectionValidationError(policy.FieldKey)
			}
		case "add_party_ref":
			if !policy.AllowPartyRefs || action.PartyID == nil {
				return collectionValidationError(policy.FieldKey)
			}
			if err := validateCollectionTargetRecordTx(ctx, tx, incidentID, *action.PartyID, policy.ExpectedTargetType, policy.FieldKey); err != nil {
				return err
			}
		case "remove_party_ref":
			if !policy.AllowPartyRefs {
				return collectionValidationError(policy.FieldKey)
			}
			if _, err := ParsePartyRefItemRef(action.ItemRef); err != nil {
				return collectionValidationError(policy.FieldKey)
			}
		case "add_tag":
			if !policy.AllowTags || action.RawText == "" || action.NormalizedText == "" {
				return collectionValidationError(policy.FieldKey)
			}
		case "remove_tag":
			if !policy.AllowTags {
				return collectionValidationError(policy.FieldKey)
			}
			if _, _, err := ParseRecordTagItemRef(action.ItemRef); err != nil {
				return collectionValidationError(policy.FieldKey)
			}
		default:
			return collectionValidationError(policy.FieldKey)
		}
	}
	return nil
}

func (s *Store) ApplyCollectionPayloadsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordID uuid.UUID, actorID uuid.UUID, policies map[string]CollectionFieldPolicy, collections map[string]CollectionActionPayload, now time.Time) (bool, error) {
	changed := false
	for fieldKey, payload := range collections {
		policy, err := collectionPolicyForField(fieldKey, policies)
		if err != nil {
			return false, err
		}
		applied, err := s.ApplyCollectionPayloadTx(ctx, tx, incidentID, recordID, actorID, policy, payload, now)
		if err != nil {
			return false, err
		}
		changed = changed || applied
	}
	return changed, nil
}

func (s *Store) ApplyCollectionPayloadTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordID uuid.UUID, actorID uuid.UUID, policy CollectionFieldPolicy, payload CollectionActionPayload, now time.Time) (bool, error) {
	if err := validateCollectionPolicy(policy); err != nil {
		return false, err
	}
	changed := false
	tags := s.Tags()
	for _, action := range payload.Actions {
		switch action.Op {
		case "add_record_ref":
			if !policy.AllowRecordRefs || action.LinkedRecordID == nil {
				return false, collectionValidationError(policy.FieldKey)
			}
			applied, err := s.UpsertFieldReferenceTx(ctx, tx, incidentID, recordID, *action.LinkedRecordID, policy.FieldKey, policy.LinkType, actorID, now)
			if err != nil {
				return false, err
			}
			changed = changed || applied
		case "remove_record_ref":
			dst, err := ParseRecordRefItemRef(action.ItemRef)
			if !policy.AllowRecordRefs || err != nil {
				return false, collectionValidationError(policy.FieldKey)
			}
			applied, err := s.TombstoneFieldReferenceTx(ctx, tx, incidentID, recordID, dst, policy.FieldKey, policy.LinkType, policy.ExpectedTargetType, actorID, now)
			if err != nil {
				if errors.Is(err, ErrFieldReferenceNotFound) {
					return false, collectionValidationError(policy.FieldKey)
				}
				return false, err
			}
			changed = changed || applied
		case "add_tag":
			if !policy.AllowTags {
				return false, collectionValidationError(policy.FieldKey)
			}
			applied, err := tags.UpsertTagTx(ctx, tx, incidentID, recordID, action.RawText, action.NormalizedText, actorID, now)
			if err != nil {
				if errors.Is(err, ErrInvalidTag) {
					return false, collectionValidationError(policy.FieldKey)
				}
				return false, err
			}
			changed = changed || applied
		case "remove_tag":
			_, tagID, err := ParseRecordTagItemRef(action.ItemRef)
			if !policy.AllowTags || err != nil {
				return false, collectionValidationError(policy.FieldKey)
			}
			applied, err := tags.TombstoneTagTx(ctx, tx, incidentID, recordID, tagID, actorID, now)
			if err != nil {
				if errors.Is(err, ErrTagNotFound) {
					return false, collectionValidationError(policy.FieldKey)
				}
				return false, err
			}
			changed = changed || applied
		case "add_party_ref":
			if !policy.AllowPartyRefs || action.PartyID == nil {
				return false, collectionValidationError(policy.FieldKey)
			}
			applied, err := s.UpsertFieldReferenceTx(ctx, tx, incidentID, recordID, *action.PartyID, policy.FieldKey, policy.LinkType, actorID, now)
			if err != nil {
				return false, err
			}
			changed = changed || applied
		case "remove_party_ref":
			dst, err := ParsePartyRefItemRef(action.ItemRef)
			if !policy.AllowPartyRefs || err != nil {
				return false, collectionValidationError(policy.FieldKey)
			}
			applied, err := s.TombstoneFieldReferenceTx(ctx, tx, incidentID, recordID, dst, policy.FieldKey, policy.LinkType, policy.ExpectedTargetType, actorID, now)
			if err != nil {
				if errors.Is(err, ErrFieldReferenceNotFound) {
					return false, collectionValidationError(policy.FieldKey)
				}
				return false, err
			}
			changed = changed || applied
		default:
			return false, collectionValidationError(policy.FieldKey)
		}
	}
	return changed, nil
}

func collectionPolicyForField(fieldKey string, policies map[string]CollectionFieldPolicy) (CollectionFieldPolicy, error) {
	policy, ok := policies[fieldKey]
	if !ok {
		return CollectionFieldPolicy{}, collectionValidationError(fieldKey)
	}
	if policy.FieldKey == "" {
		policy.FieldKey = fieldKey
	}
	if policy.FieldKey != fieldKey {
		return CollectionFieldPolicy{}, collectionValidationError(fieldKey)
	}
	return policy, nil
}

func validateCollectionPolicy(policy CollectionFieldPolicy) error {
	if policy.FieldKey == "" {
		return collectionValidationError("field_key")
	}
	if (policy.AllowRecordRefs || policy.AllowPartyRefs) && policy.LinkType == "" {
		return collectionValidationError(policy.FieldKey)
	}
	if policy.AllowTags && (policy.AllowRecordRefs || policy.AllowPartyRefs) {
		return collectionValidationError(policy.FieldKey)
	}
	return nil
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

func collectionValidationError(field string) *CollectionValidationError {
	return &CollectionValidationError{Field: field, ReasonCode: "invalid_value"}
}
