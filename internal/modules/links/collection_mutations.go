package links

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (s *Store) SyncFieldReferenceWithMutationValuesTx(ctx context.Context, tx pgx.Tx, command SyncFieldReferenceCommand) (CollectionMutationResult, error) {
	before, err := s.loadActiveFieldReferenceMutationValuesTx(ctx, tx, command)
	if err != nil {
		return CollectionMutationResult{}, err
	}
	if _, err := s.SyncFieldReferenceCommandTx(ctx, tx, command); err != nil {
		return CollectionMutationResult{}, err
	}
	after, err := s.loadActiveFieldReferenceMutationValuesTx(ctx, tx, command)
	if err != nil {
		return CollectionMutationResult{}, err
	}
	result := CollectionMutationResult{RecordLinks: make([]RecordLinkMutation, 0)}
	for _, item := range sortedFieldReferenceMutationValues(before) {
		if _, retained := after[item.RecordLinkID]; retained {
			continue
		}
		afterValue, err := s.loadRecordLinkMutationValueTx(ctx, tx, item.RecordLinkID)
		if err != nil {
			return CollectionMutationResult{}, err
		}
		result.RecordLinks = append(result.RecordLinks, RecordLinkMutation{
			RecordLinkID: item.RecordLinkID,
			Operation:    "delete",
			BeforeValue:  item.Value,
			AfterValue:   afterValue.Map(),
		})
	}
	for _, item := range sortedFieldReferenceMutationValues(after) {
		if _, existed := before[item.RecordLinkID]; existed {
			continue
		}
		result.RecordLinks = append(result.RecordLinks, RecordLinkMutation{
			RecordLinkID: item.RecordLinkID,
			Operation:    "create",
			AfterValue:   item.Value,
		})
	}
	return result, nil
}

func sortedFieldReferenceMutationValues(values map[uuid.UUID]fieldReferenceMutationValue) []fieldReferenceMutationValue {
	result := make([]fieldReferenceMutationValue, 0, len(values))
	for _, value := range values {
		result = append(result, value)
	}
	slices.SortFunc(result, func(left fieldReferenceMutationValue, right fieldReferenceMutationValue) int {
		return strings.Compare(left.RecordLinkID.String(), right.RecordLinkID.String())
	})
	return result
}

type fieldReferenceMutationValue struct {
	RecordLinkID uuid.UUID
	Value        map[string]any
}

func (s *Store) loadActiveFieldReferenceMutationValuesTx(ctx context.Context, tx pgx.Tx, command SyncFieldReferenceCommand) (map[uuid.UUID]fieldReferenceMutationValue, error) {
	rows, err := tx.Query(ctx, `
SELECT record_link_id
  FROM record_links
 WHERE incident_id = $1
   AND src_record_id = $2
   AND field_key = $3
   AND link_type = $4
   AND deleted_at IS NULL
 ORDER BY record_link_id
`, command.IncidentID, command.SrcRecordID, command.FieldKey, command.LinkType.String())
	if err != nil {
		return nil, fmt.Errorf("list field-reference mutation values: %w", err)
	}
	defer rows.Close()
	ids := make([]uuid.UUID, 0)
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan field-reference mutation id: %w", err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate field-reference mutation ids: %w", err)
	}
	result := make(map[uuid.UUID]fieldReferenceMutationValue, len(ids))
	for _, id := range ids {
		value, err := s.loadRecordLinkMutationValueTx(ctx, tx, id)
		if err != nil {
			return nil, err
		}
		result[id] = fieldReferenceMutationValue{RecordLinkID: id, Value: value.Map()}
	}
	return result, nil
}

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
		afterValue, err := s.loadRecordLinkMutationValueTx(ctx, tx, record.RecordLinkID)
		if err != nil {
			return CollectionMutationResult{}, err
		}
		result.RecordLinks = append(result.RecordLinks, RecordLinkMutation{RecordLinkID: record.RecordLinkID, Operation: "create", AfterValue: afterValue.Map()})
	}
	for _, recordID := range command.RemoveRecordIDs {
		existing, err := s.GetActiveFieldReferenceTx(ctx, tx, command.IncidentID, command.SourceRecordID, recordID, command.FieldKey, linkType)
		if errors.Is(err, ErrFieldReferenceNotFound) {
			return CollectionMutationResult{}, collectionValidationError(command.FieldKey)
		}
		if err != nil {
			return CollectionMutationResult{}, err
		}
		beforeValue, err := s.loadRecordLinkMutationValueTx(ctx, tx, existing.RecordLinkID)
		if err != nil {
			return CollectionMutationResult{}, err
		}
		tombstoned, err := s.TombstoneFieldReferenceRecordTx(ctx, tx, command.IncidentID, command.SourceRecordID, recordID, command.FieldKey, linkType, command.ExpectedTargetType, command.ActorUserID, command.Now)
		if err != nil {
			return CollectionMutationResult{}, err
		}
		afterValue, err := s.loadRecordLinkMutationValueTx(ctx, tx, tombstoned.RecordLinkID)
		if err != nil {
			return CollectionMutationResult{}, err
		}
		result.RecordLinks = append(result.RecordLinks, RecordLinkMutation{RecordLinkID: tombstoned.RecordLinkID, Operation: "delete", BeforeValue: beforeValue.Map(), AfterValue: afterValue.Map()})
	}
	return result, nil
}

func (s *Store) ApplyPartyRefCollectionWithMutationValuesTx(ctx context.Context, tx pgx.Tx, command PartyRefCollectionCommand) (CollectionMutationResult, error) {
	if err := s.ValidatePartyRefCollectionTx(ctx, tx, PartyRefCollectionValidation{
		IncidentID:         command.IncidentID,
		FieldKey:           command.FieldKey,
		LinkType:           command.LinkType,
		ExpectedTargetType: command.ExpectedTargetType,
		AddPartyIDs:        command.AddPartyIDs,
		RemovePartyIDs:     command.RemovePartyIDs,
	}); err != nil {
		return CollectionMutationResult{}, err
	}
	result := CollectionMutationResult{RecordLinks: make([]RecordLinkMutation, 0)}
	linkType := command.LinkType.String()
	for _, partyID := range command.AddPartyIDs {
		record, inserted, err := s.UpsertFieldReferenceRecordTx(ctx, tx, command.IncidentID, command.SourceRecordID, partyID, command.FieldKey, linkType, command.ActorUserID, command.Now)
		if err != nil {
			return CollectionMutationResult{}, err
		}
		if !inserted {
			continue
		}
		afterValue, err := s.loadRecordLinkMutationValueTx(ctx, tx, record.RecordLinkID)
		if err != nil {
			return CollectionMutationResult{}, err
		}
		result.RecordLinks = append(result.RecordLinks, RecordLinkMutation{RecordLinkID: record.RecordLinkID, Operation: "create", AfterValue: afterValue.Map()})
	}
	for _, partyID := range command.RemovePartyIDs {
		existing, err := s.GetActiveFieldReferenceTx(ctx, tx, command.IncidentID, command.SourceRecordID, partyID, command.FieldKey, linkType)
		if errors.Is(err, ErrFieldReferenceNotFound) {
			return CollectionMutationResult{}, collectionValidationError(command.FieldKey)
		}
		if err != nil {
			return CollectionMutationResult{}, err
		}
		beforeValue, err := s.loadRecordLinkMutationValueTx(ctx, tx, existing.RecordLinkID)
		if err != nil {
			return CollectionMutationResult{}, err
		}
		tombstoned, err := s.TombstoneFieldReferenceRecordTx(ctx, tx, command.IncidentID, command.SourceRecordID, partyID, command.FieldKey, linkType, command.ExpectedTargetType, command.ActorUserID, command.Now)
		if err != nil {
			return CollectionMutationResult{}, err
		}
		afterValue, err := s.loadRecordLinkMutationValueTx(ctx, tx, tombstoned.RecordLinkID)
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
		afterValue, err := s.loadRecordTagMutationValueTx(ctx, tx, tagID)
		if err != nil {
			return CollectionMutationResult{}, err
		}
		result.RecordTags = append(result.RecordTags, RecordTagMutation{RecordTagID: tagID, RecordID: command.RecordID, Operation: "create", AfterValue: afterValue.Map()})
	}
	for _, tag := range command.RemoveTags {
		if tag.RecordID != command.RecordID {
			return CollectionMutationResult{}, collectionValidationError(command.FieldKey)
		}
		beforeValue, err := s.loadRecordTagMutationValueTx(ctx, tx, tag.RecordTagID)
		if errors.Is(err, ErrTagNotFound) {
			return CollectionMutationResult{}, collectionValidationError(command.FieldKey)
		}
		if err != nil {
			return CollectionMutationResult{}, err
		}
		deleted, err := s.Tags().TombstoneTagRecordTx(ctx, tx, command.IncidentID, command.RecordID, tag.RecordTagID, command.ActorUserID, command.Now)
		if errors.Is(err, ErrTagNotFound) {
			return CollectionMutationResult{}, collectionValidationError(command.FieldKey)
		}
		if err != nil {
			return CollectionMutationResult{}, err
		}
		afterValue, err := s.loadRecordTagMutationValueTx(ctx, tx, deleted)
		if err != nil {
			return CollectionMutationResult{}, err
		}
		result.RecordTags = append(result.RecordTags, RecordTagMutation{RecordTagID: deleted, RecordID: command.RecordID, Operation: "delete", BeforeValue: beforeValue.Map(), AfterValue: afterValue.Map()})
	}
	return result, nil
}
