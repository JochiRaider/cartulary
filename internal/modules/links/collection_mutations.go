package links

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type CollectionMutationResult struct {
	RecordLinks []RecordLinkMutation
	RecordTags  []RecordTagMutation
}

type RecordLinkMutation struct {
	RecordLinkID uuid.UUID
	Operation    string
	BeforeValue  map[string]any
	AfterValue   map[string]any
}

type RecordTagMutation struct {
	RecordTagID uuid.UUID
	RecordID    uuid.UUID
	Operation   string
	BeforeValue map[string]any
	AfterValue  map[string]any
}

func (s *Store) ApplyCollectionPayloadWithMutationValuesTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordID uuid.UUID, actorID uuid.UUID, policy CollectionFieldPolicy, payload CollectionActionPayload, now time.Time) (CollectionMutationResult, error) {
	if err := validateCollectionPolicy(policy); err != nil {
		return CollectionMutationResult{}, err
	}
	if err := s.ValidateCollectionPayloadTx(ctx, tx, incidentID, policy, payload); err != nil {
		return CollectionMutationResult{}, err
	}
	result := CollectionMutationResult{
		RecordLinks: make([]RecordLinkMutation, 0),
		RecordTags:  make([]RecordTagMutation, 0),
	}
	for _, action := range payload.Actions {
		switch action.Op {
		case "add_record_ref":
			record, inserted, err := s.UpsertFieldReferenceRecordTx(ctx, tx, incidentID, recordID, *action.LinkedRecordID, policy.FieldKey, policy.LinkType, actorID, now)
			if err != nil {
				return CollectionMutationResult{}, err
			}
			if !inserted {
				continue
			}
			afterValue, err := s.LoadRecordLinkMutationValueTx(ctx, tx, record.RecordLinkID)
			if err != nil {
				return CollectionMutationResult{}, err
			}
			after := afterValue.Map()
			result.RecordLinks = append(result.RecordLinks, RecordLinkMutation{RecordLinkID: record.RecordLinkID, Operation: "create", AfterValue: after})
		case "remove_record_ref":
			dst, err := ParseRecordRefItemRef(action.ItemRef)
			if err != nil {
				return CollectionMutationResult{}, collectionValidationError(policy.FieldKey)
			}
			existing, err := s.GetActiveFieldReferenceTx(ctx, tx, incidentID, recordID, dst, policy.FieldKey, policy.LinkType)
			if errors.Is(err, ErrFieldReferenceNotFound) {
				continue
			}
			if err != nil {
				return CollectionMutationResult{}, err
			}
			beforeValue, err := s.LoadRecordLinkMutationValueTx(ctx, tx, existing.RecordLinkID)
			if err != nil {
				return CollectionMutationResult{}, err
			}
			before := beforeValue.Map()
			tombstoned, err := s.TombstoneFieldReferenceRecordTx(ctx, tx, incidentID, recordID, dst, policy.FieldKey, policy.LinkType, policy.ExpectedTargetType, actorID, now)
			if err != nil {
				return CollectionMutationResult{}, err
			}
			afterValue, err := s.LoadRecordLinkMutationValueTx(ctx, tx, tombstoned.RecordLinkID)
			if err != nil {
				return CollectionMutationResult{}, err
			}
			after := afterValue.Map()
			result.RecordLinks = append(result.RecordLinks, RecordLinkMutation{RecordLinkID: tombstoned.RecordLinkID, Operation: "delete", BeforeValue: before, AfterValue: after})
		case "add_tag":
			tagID, inserted, err := s.Tags().UpsertTagRecordTx(ctx, tx, incidentID, recordID, action.RawText, action.NormalizedText, actorID, now)
			if err != nil {
				if errors.Is(err, ErrInvalidTag) {
					return CollectionMutationResult{}, collectionValidationError(policy.FieldKey)
				}
				return CollectionMutationResult{}, err
			}
			if !inserted {
				continue
			}
			afterValue, err := s.LoadRecordTagMutationValueTx(ctx, tx, tagID)
			if err != nil {
				return CollectionMutationResult{}, err
			}
			after := afterValue.Map()
			result.RecordTags = append(result.RecordTags, RecordTagMutation{RecordTagID: tagID, RecordID: recordID, Operation: "create", AfterValue: after})
		case "remove_tag":
			itemRecordID, tagID, err := ParseRecordTagItemRef(action.ItemRef)
			if err != nil || itemRecordID != recordID {
				return CollectionMutationResult{}, collectionValidationError(policy.FieldKey)
			}
			beforeValue, err := s.LoadRecordTagMutationValueTx(ctx, tx, tagID)
			if errors.Is(err, ErrTagNotFound) {
				continue
			}
			if err != nil {
				return CollectionMutationResult{}, err
			}
			before := beforeValue.Map()
			deleted, err := s.Tags().TombstoneTagRecordTx(ctx, tx, incidentID, recordID, tagID, actorID, now)
			if errors.Is(err, ErrTagNotFound) {
				continue
			}
			if err != nil {
				return CollectionMutationResult{}, err
			}
			afterValue, err := s.LoadRecordTagMutationValueTx(ctx, tx, deleted)
			if err != nil {
				return CollectionMutationResult{}, err
			}
			after := afterValue.Map()
			result.RecordTags = append(result.RecordTags, RecordTagMutation{RecordTagID: deleted, RecordID: recordID, Operation: "delete", BeforeValue: before, AfterValue: after})
		default:
			return CollectionMutationResult{}, fmt.Errorf("unsupported collection action: %s", action.Op)
		}
	}
	return result, nil
}
