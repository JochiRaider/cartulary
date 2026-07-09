package links

import (
	"context"
	"errors"

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

func (s *Store) ApplyRecordRefCollectionWithMutationValuesTx(ctx context.Context, tx pgx.Tx, command RecordRefCollectionCommand) (CollectionMutationResult, error) {
	if err := s.ValidateRecordRefCollectionTx(ctx, tx, RecordRefCollectionValidation{
		IncidentID:         command.IncidentID,
		FieldKey:           command.FieldKey,
		LinkType:           command.LinkType,
		ExpectedTargetType: command.ExpectedTargetType,
		AddRecordIDs:       command.AddRecordIDs,
		RemoveRecordIDs:    command.RemoveRecordIDs,
	}); err != nil {
		return CollectionMutationResult{}, err
	}
	result := CollectionMutationResult{
		RecordLinks: make([]RecordLinkMutation, 0),
		RecordTags:  make([]RecordTagMutation, 0),
	}
	linkType := command.LinkType.String()
	for _, recordID := range command.AddRecordIDs {
		record, inserted, err := s.UpsertFieldReferenceRecordTx(ctx, tx, command.IncidentID, command.SourceRecordID, recordID, command.FieldKey, linkType, command.ActorUserID, command.Now)
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
		result.RecordLinks = append(result.RecordLinks, RecordLinkMutation{RecordLinkID: record.RecordLinkID, Operation: "create", AfterValue: afterValue.Map()})
	}
	for _, recordID := range command.RemoveRecordIDs {
		existing, err := s.GetActiveFieldReferenceTx(ctx, tx, command.IncidentID, command.SourceRecordID, recordID, command.FieldKey, linkType)
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
		tombstoned, err := s.TombstoneFieldReferenceRecordTx(ctx, tx, command.IncidentID, command.SourceRecordID, recordID, command.FieldKey, linkType, command.ExpectedTargetType, command.ActorUserID, command.Now)
		if err != nil {
			return CollectionMutationResult{}, err
		}
		afterValue, err := s.LoadRecordLinkMutationValueTx(ctx, tx, tombstoned.RecordLinkID)
		if err != nil {
			return CollectionMutationResult{}, err
		}
		result.RecordLinks = append(result.RecordLinks, RecordLinkMutation{RecordLinkID: tombstoned.RecordLinkID, Operation: "delete", BeforeValue: beforeValue.Map(), AfterValue: afterValue.Map()})
	}
	return result, nil
}

func (s *Store) ApplyTagCollectionWithMutationValuesTx(ctx context.Context, tx pgx.Tx, command TagCollectionCommand) (CollectionMutationResult, error) {
	if err := s.ValidateTagCollectionTx(ctx, tx, TagCollectionValidation{
		FieldKey:   command.FieldKey,
		AddTags:    command.AddTags,
		RemoveTags: command.RemoveTags,
	}); err != nil {
		return CollectionMutationResult{}, err
	}
	result := CollectionMutationResult{
		RecordLinks: make([]RecordLinkMutation, 0),
		RecordTags:  make([]RecordTagMutation, 0),
	}
	for _, tag := range command.AddTags {
		tagID, inserted, err := s.Tags().UpsertTagRecordTx(ctx, tx, command.IncidentID, command.RecordID, tag.RawText, tag.NormalizedText, command.ActorUserID, command.Now)
		if err != nil {
			if errors.Is(err, ErrInvalidTag) {
				return CollectionMutationResult{}, collectionValidationError(command.FieldKey)
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
		result.RecordTags = append(result.RecordTags, RecordTagMutation{RecordTagID: tagID, RecordID: command.RecordID, Operation: "create", AfterValue: afterValue.Map()})
	}
	for _, tag := range command.RemoveTags {
		if tag.RecordID != command.RecordID {
			return CollectionMutationResult{}, collectionValidationError(command.FieldKey)
		}
		beforeValue, err := s.LoadRecordTagMutationValueTx(ctx, tx, tag.RecordTagID)
		if errors.Is(err, ErrTagNotFound) {
			continue
		}
		if err != nil {
			return CollectionMutationResult{}, err
		}
		deleted, err := s.Tags().TombstoneTagRecordTx(ctx, tx, command.IncidentID, command.RecordID, tag.RecordTagID, command.ActorUserID, command.Now)
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
		result.RecordTags = append(result.RecordTags, RecordTagMutation{RecordTagID: deleted, RecordID: command.RecordID, Operation: "delete", BeforeValue: beforeValue.Map(), AfterValue: afterValue.Map()})
	}
	return result, nil
}
