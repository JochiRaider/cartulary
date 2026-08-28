package links

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/JochiRaider/cartulary/internal/modules/links/internal/mutationvalue"
	"github.com/JochiRaider/cartulary/internal/modules/links/internal/valuecodec"
)

type CollectionMutationResult struct {
	mutations []Mutation
}

func (result CollectionMutationResult) Mutations() []Mutation {
	return mutationvalue.Copy(result.mutations)
}

func (s *Store) SyncFieldReferenceWithMutationValuesTx(ctx context.Context, tx pgx.Tx, command SyncFieldReferenceCommand) (CollectionMutationResult, error) {
	if err := validateLinkCollectionPolicy(command.FieldKey, command.LinkType); err != nil {
		return CollectionMutationResult{}, err
	}
	if command.IncidentID == uuid.Nil || command.SrcRecordID == uuid.Nil || command.ActorUserID == uuid.Nil || command.Now.IsZero() {
		return CollectionMutationResult{}, collectionValidationError(command.FieldKey)
	}
	if command.TargetID != nil {
		if err := validateRecordLinkCommand(command.LinkType, LinkProvenanceManual, nil, command.SrcRecordID, *command.TargetID); err != nil {
			return CollectionMutationResult{}, err
		}
		if err := validateActiveLinkEndpointsTx(ctx, tx, command.IncidentID, command.SrcRecordID, *command.TargetID); err != nil {
			return CollectionMutationResult{}, err
		}
	}

	before, err := listActiveFieldReferenceStatesTx(ctx, tx, command.IncidentID, command.SrcRecordID, command.FieldKey, command.LinkType)
	if err != nil {
		return CollectionMutationResult{}, err
	}
	result := newCollectionMutationResult()
	retainedTarget := false
	for _, state := range before {
		if command.TargetID != nil && state.dstRecordID == *command.TargetID {
			retainedTarget = true
			continue
		}
		after, err := tombstoneRecordLinkStateTx(ctx, tx, state.recordLinkID, command.ActorUserID, command.Now.UTC())
		if err != nil {
			return CollectionMutationResult{}, err
		}
		mutation, err := newRecordLinkMutation("delete", &state, &after)
		if err != nil {
			return CollectionMutationResult{}, err
		}
		result.mutations = append(result.mutations, mutation)
	}
	if command.TargetID == nil || retainedTarget {
		return result, nil
	}
	created, inserted, err := upsertFieldReferenceStateTx(ctx, tx, command.IncidentID, command.SrcRecordID, *command.TargetID, command.FieldKey, command.LinkType, command.ActorUserID, command.Now.UTC())
	if err != nil {
		return CollectionMutationResult{}, err
	}
	if inserted {
		mutation, err := newRecordLinkMutation("create", nil, &created)
		if err != nil {
			return CollectionMutationResult{}, err
		}
		result.mutations = append(result.mutations, mutation)
	}
	return result, nil
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
	return s.applyReferenceCollectionWithMutationValuesTx(ctx, tx, referenceCollectionCommand{
		incidentID:         command.IncidentID,
		sourceRecordID:     command.SourceRecordID,
		actorUserID:        command.ActorUserID,
		fieldKey:           command.FieldKey,
		linkType:           command.LinkType,
		expectedTargetType: command.ExpectedTargetType,
		addRecordIDs:       command.AddRecordIDs,
		removeRecordIDs:    command.RemoveRecordIDs,
		now:                command.Now,
	})
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
	return s.applyReferenceCollectionWithMutationValuesTx(ctx, tx, referenceCollectionCommand{
		incidentID:         command.IncidentID,
		sourceRecordID:     command.SourceRecordID,
		actorUserID:        command.ActorUserID,
		fieldKey:           command.FieldKey,
		linkType:           command.LinkType,
		expectedTargetType: command.ExpectedTargetType,
		addRecordIDs:       command.AddPartyIDs,
		removeRecordIDs:    command.RemovePartyIDs,
		now:                command.Now,
	})
}

type referenceCollectionCommand struct {
	incidentID         uuid.UUID
	sourceRecordID     uuid.UUID
	actorUserID        uuid.UUID
	fieldKey           string
	linkType           LinkType
	expectedTargetType string
	addRecordIDs       []uuid.UUID
	removeRecordIDs    []uuid.UUID
	now                time.Time
}

func (s *Store) applyReferenceCollectionWithMutationValuesTx(ctx context.Context, tx pgx.Tx, command referenceCollectionCommand) (CollectionMutationResult, error) {
	if command.incidentID == uuid.Nil || command.sourceRecordID == uuid.Nil || command.actorUserID == uuid.Nil || command.now.IsZero() {
		return CollectionMutationResult{}, collectionValidationError(command.fieldKey)
	}
	result := newCollectionMutationResult()
	for _, recordID := range command.addRecordIDs {
		state, inserted, err := upsertFieldReferenceStateTx(ctx, tx, command.incidentID, command.sourceRecordID, recordID, command.fieldKey, command.linkType, command.actorUserID, command.now.UTC())
		if err != nil {
			return CollectionMutationResult{}, err
		}
		if inserted {
			mutation, err := newRecordLinkMutation("create", nil, &state)
			if err != nil {
				return CollectionMutationResult{}, err
			}
			result.mutations = append(result.mutations, mutation)
		}
	}
	for _, recordID := range command.removeRecordIDs {
		before, err := getActiveFieldReferenceStateTx(ctx, tx, command.incidentID, command.sourceRecordID, recordID, command.fieldKey, command.linkType)
		if errors.Is(err, errFieldReferenceNotFound) {
			return CollectionMutationResult{}, collectionValidationError(command.fieldKey)
		}
		if err != nil {
			return CollectionMutationResult{}, err
		}
		after, err := tombstoneFieldReferenceStateTx(ctx, tx, before.recordLinkID, command.expectedTargetType, command.actorUserID, command.now.UTC())
		if errors.Is(err, errFieldReferenceNotFound) {
			return CollectionMutationResult{}, collectionValidationError(command.fieldKey)
		}
		if err != nil {
			return CollectionMutationResult{}, err
		}
		mutation, err := newRecordLinkMutation("delete", &before, &after)
		if err != nil {
			return CollectionMutationResult{}, err
		}
		result.mutations = append(result.mutations, mutation)
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
	if command.IncidentID == uuid.Nil || command.RecordID == uuid.Nil || command.ActorUserID == uuid.Nil || command.Now.IsZero() {
		return CollectionMutationResult{}, collectionValidationError(command.FieldKey)
	}
	result := newCollectionMutationResult()
	for _, tag := range command.AddTags {
		state, inserted, err := upsertRecordTagStateTx(ctx, tx, command.IncidentID, command.RecordID, tag.RawText, tag.NormalizedText, command.ActorUserID, command.Now.UTC())
		if errors.Is(err, errInvalidTag) {
			return CollectionMutationResult{}, collectionValidationError(command.FieldKey)
		}
		if err != nil {
			return CollectionMutationResult{}, err
		}
		if inserted {
			mutation, err := newRecordTagMutation("create", nil, &state)
			if err != nil {
				return CollectionMutationResult{}, err
			}
			result.mutations = append(result.mutations, mutation)
		}
	}
	for _, tag := range command.RemoveTags {
		if tag.RecordID != command.RecordID {
			return CollectionMutationResult{}, collectionValidationError(command.FieldKey)
		}
		before, err := getActiveRecordTagStateTx(ctx, tx, command.IncidentID, command.RecordID, tag.RecordTagID)
		if errors.Is(err, errTagNotFound) {
			return CollectionMutationResult{}, collectionValidationError(command.FieldKey)
		}
		if err != nil {
			return CollectionMutationResult{}, err
		}
		after, err := tombstoneRecordTagStateTx(ctx, tx, tag.RecordTagID, command.ActorUserID, command.Now.UTC())
		if errors.Is(err, errTagNotFound) {
			return CollectionMutationResult{}, collectionValidationError(command.FieldKey)
		}
		if err != nil {
			return CollectionMutationResult{}, err
		}
		mutation, err := newRecordTagMutation("delete", &before, &after)
		if err != nil {
			return CollectionMutationResult{}, err
		}
		result.mutations = append(result.mutations, mutation)
	}
	return result, nil
}

func newCollectionMutationResult() CollectionMutationResult {
	return CollectionMutationResult{mutations: make([]Mutation, 0)}
}

type recordTagState struct {
	recordTagID       uuid.UUID
	incidentID        uuid.UUID
	recordID          uuid.UUID
	tagName           string
	normalizedTagName string
	createdByUserID   uuid.UUID
	createdAt         time.Time
	updatedAt         time.Time
	deletedAt         *time.Time
	deletedByUserID   *uuid.UUID
}

func upsertRecordTagStateTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordID uuid.UUID, tagName string, normalizedTagName string, actorUserID uuid.UUID, now time.Time) (recordTagState, bool, error) {
	if tagName == "" || normalizedTagName == "" {
		return recordTagState{}, false, errInvalidTag
	}
	row := tx.QueryRow(ctx, `
INSERT INTO record_tags (
    incident_id, record_id, tag_name, normalized_tag_name,
    created_by_user_id, created_at, updated_at
) VALUES ($1, $2, $3, $4, $5, $6, $6)
ON CONFLICT (incident_id, record_id, normalized_tag_name)
WHERE deleted_at IS NULL
DO NOTHING
RETURNING
    record_tag_id, incident_id, record_id, tag_name, normalized_tag_name,
    created_by_user_id, created_at, updated_at, deleted_at,
    deleted_by_user_id
`, incidentID, recordID, tagName, normalizedTagName, actorUserID, now.UTC())
	state, err := scanRecordTagState(row)
	if err == nil {
		return state, true, nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return recordTagState{}, false, nil
	}
	return recordTagState{}, false, fmt.Errorf("upsert record tag state: %w", err)
}

func getActiveRecordTagStateTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordID uuid.UUID, recordTagID uuid.UUID) (recordTagState, error) {
	row := tx.QueryRow(ctx, `
SELECT
    record_tag_id, incident_id, record_id, tag_name, normalized_tag_name,
    created_by_user_id, created_at, updated_at, deleted_at,
    deleted_by_user_id
  FROM record_tags
 WHERE incident_id = $1
   AND record_id = $2
   AND record_tag_id = $3
   AND deleted_at IS NULL
 FOR UPDATE
`, incidentID, recordID, recordTagID)
	state, err := scanRecordTagState(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return recordTagState{}, errTagNotFound
	}
	if err != nil {
		return recordTagState{}, fmt.Errorf("get active record tag state: %w", err)
	}
	return state, nil
}

func tombstoneRecordTagStateTx(ctx context.Context, tx pgx.Tx, recordTagID uuid.UUID, actorUserID uuid.UUID, now time.Time) (recordTagState, error) {
	row := tx.QueryRow(ctx, `
UPDATE record_tags
   SET deleted_at = $3,
       deleted_by_user_id = $2,
       updated_at = $3
 WHERE record_tag_id = $1
   AND deleted_at IS NULL
RETURNING
    record_tag_id, incident_id, record_id, tag_name, normalized_tag_name,
    created_by_user_id, created_at, updated_at, deleted_at,
    deleted_by_user_id
`, recordTagID, actorUserID, now.UTC())
	state, err := scanRecordTagState(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return recordTagState{}, errTagNotFound
	}
	if err != nil {
		return recordTagState{}, fmt.Errorf("tombstone record tag state: %w", err)
	}
	return state, nil
}

func scanRecordTagState(row pgx.Row) (recordTagState, error) {
	var (
		state           recordTagState
		deletedAt       pgtype.Timestamptz
		deletedByUserID pgtype.UUID
	)
	if err := row.Scan(
		&state.recordTagID,
		&state.incidentID,
		&state.recordID,
		&state.tagName,
		&state.normalizedTagName,
		&state.createdByUserID,
		&state.createdAt,
		&state.updatedAt,
		&deletedAt,
		&deletedByUserID,
	); err != nil {
		return recordTagState{}, err
	}
	state.createdAt = state.createdAt.UTC()
	state.updatedAt = state.updatedAt.UTC()
	if deletedAt.Valid {
		value := deletedAt.Time.UTC()
		state.deletedAt = &value
	}
	if deletedByUserID.Valid {
		value := uuid.UUID(deletedByUserID.Bytes)
		state.deletedByUserID = &value
	}
	return state, nil
}

func (s recordTagState) mutationValue() valuecodec.RecordTagMutationValue {
	return valuecodec.BuildRecordTagMutationValue(valuecodec.RecordTagMutationInput{
		RecordTagID:       s.recordTagID,
		IncidentID:        s.incidentID,
		RecordID:          s.recordID,
		TagName:           s.tagName,
		NormalizedTagName: s.normalizedTagName,
		CreatedByUserID:   s.createdByUserID,
		CreatedAt:         s.createdAt,
		UpdatedAt:         s.updatedAt,
		DeletedAt:         copyTimePointer(s.deletedAt),
		DeletedByUserID:   copyUUIDPointer(s.deletedByUserID),
	})
}

func newRecordTagMutation(operation string, before *recordTagState, after *recordTagState) (Mutation, error) {
	var recordID uuid.UUID
	var recordTagID uuid.UUID
	var beforeValue map[string]any
	var afterValue map[string]any
	if before != nil {
		recordTagID = before.recordTagID
		recordID = before.recordID
		beforeValue = before.mutationValue().Map()
	}
	if after != nil {
		recordTagID = after.recordTagID
		recordID = after.recordID
		afterValue = after.mutationValue().Map()
	}
	return mutationvalue.New(
		mutationvalue.TargetRecordTag,
		"record_tag:"+recordID.String()+":"+recordTagID.String(),
		operation,
		beforeValue,
		afterValue,
	)
}
