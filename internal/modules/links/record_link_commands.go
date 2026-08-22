package links

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/JochiRaider/cartulary/internal/modules/links/internal/valuecodec"
)

type recordLinkState struct {
	recordLinkID    uuid.UUID
	incidentID      uuid.UUID
	srcRecordID     uuid.UUID
	dstRecordID     uuid.UUID
	linkType        LinkType
	fieldKey        *string
	provenance      LinkProvenance
	confidence      *int
	ownerUserID     uuid.UUID
	createdByUserID uuid.UUID
	decidedAt       time.Time
	createdAt       time.Time
	deletedAt       *time.Time
	deletedByUserID *uuid.UUID
}

func (s *Store) UpsertLinkCommandTx(ctx context.Context, tx pgx.Tx, command UpsertLinkCommand) (RecordLinkCommandResult, error) {
	linkType := command.LinkType.String()
	provenance := command.Provenance.String()
	now := command.Now.UTC()
	if err := validateRecordLinkCommand(linkType, provenance, command.Confidence, command.SrcRecordID, command.DstRecordID); err != nil {
		return RecordLinkCommandResult{}, err
	}
	if command.IncidentID == uuid.Nil || command.OwnerUserID == uuid.Nil || command.Now.IsZero() {
		return RecordLinkCommandResult{}, fmt.Errorf("%w: missing command identity or time", errInvalidRecordLink)
	}
	if err := validateActiveLinkEndpointsTx(ctx, tx, command.IncidentID, command.SrcRecordID, command.DstRecordID); err != nil {
		return RecordLinkCommandResult{}, err
	}

	existing, err := getActiveRecordLinkStateTx(ctx, tx, command.IncidentID, command.SrcRecordID, command.DstRecordID, command.LinkType)
	if err == nil {
		if existing.provenance == command.Provenance && intPointersEqual(existing.confidence, command.Confidence) {
			return newRecordLinkCommandResult(existing, nil), nil
		}
		updated, err := updateRecordLinkMetadataTx(ctx, tx, existing.recordLinkID, command.Provenance, command.Confidence, command.OwnerUserID, now)
		if err != nil {
			return RecordLinkCommandResult{}, err
		}
		mutation := newRecordLinkMutation("patch", &existing, &updated)
		return newRecordLinkCommandResult(updated, &mutation), nil
	}
	if !errors.Is(err, errRecordLinkNotFound) {
		return RecordLinkCommandResult{}, err
	}

	created, err := insertRecordLinkStateTx(ctx, tx, recordLinkInsert{
		incidentID:      command.IncidentID,
		srcRecordID:     command.SrcRecordID,
		dstRecordID:     command.DstRecordID,
		linkType:        command.LinkType,
		provenance:      command.Provenance,
		confidence:      command.Confidence,
		ownerUserID:     command.OwnerUserID,
		createdByUserID: command.OwnerUserID,
		decidedAt:       now,
		createdAt:       now,
	})
	if err != nil {
		return RecordLinkCommandResult{}, fmt.Errorf("insert link: %w", err)
	}
	mutation := newRecordLinkMutation("create", nil, &created)
	return newRecordLinkCommandResult(created, &mutation), nil
}

func (s *Store) InsertSupersedesCommandTx(ctx context.Context, tx pgx.Tx, command InsertSupersedesCommand) (RecordLinkCommandResult, error) {
	now := command.Now.UTC()
	if err := validateRecordLinkCommand(LinkTypeSupersedes, LinkProvenanceManual, nil, command.ReplacementRecordID, command.SupersededRecordID); err != nil {
		return RecordLinkCommandResult{}, err
	}
	if command.IncidentID == uuid.Nil || command.OwnerUserID == uuid.Nil || command.Now.IsZero() {
		return RecordLinkCommandResult{}, fmt.Errorf("%w: missing command identity or time", errInvalidRecordLink)
	}
	if err := validateSupersedesEndpointsTx(ctx, tx, command.IncidentID, command.ReplacementRecordID, command.SupersededRecordID); err != nil {
		return RecordLinkCommandResult{}, err
	}
	created, err := insertRecordLinkStateTx(ctx, tx, recordLinkInsert{
		incidentID:      command.IncidentID,
		srcRecordID:     command.ReplacementRecordID,
		dstRecordID:     command.SupersededRecordID,
		linkType:        LinkType(LinkTypeSupersedes),
		provenance:      LinkProvenance(LinkProvenanceManual),
		ownerUserID:     command.OwnerUserID,
		createdByUserID: command.OwnerUserID,
		decidedAt:       now,
		createdAt:       now,
	})
	if err != nil {
		return RecordLinkCommandResult{}, fmt.Errorf("insert supersedes link: %w", err)
	}
	mutation := newRecordLinkMutation("create", nil, &created)
	return newRecordLinkCommandResult(created, &mutation), nil
}

func (s *Store) TombstoneActiveLinkCommandTx(ctx context.Context, tx pgx.Tx, command TombstoneActiveLinkCommand) (RecordLinkCommandResult, bool, error) {
	if command.IncidentID == uuid.Nil || command.ActorUserID == uuid.Nil || command.Now.IsZero() {
		return RecordLinkCommandResult{}, false, fmt.Errorf("%w: missing command identity or time", errInvalidRecordLink)
	}
	if !isKnownLinkType(command.LinkType.String()) || command.SrcRecordID == uuid.Nil || command.DstRecordID == uuid.Nil || command.SrcRecordID == command.DstRecordID {
		return RecordLinkCommandResult{}, false, fmt.Errorf("%w: invalid link tuple", errInvalidRecordLink)
	}
	existing, err := getActiveRecordLinkStateTx(ctx, tx, command.IncidentID, command.SrcRecordID, command.DstRecordID, command.LinkType)
	if errors.Is(err, errRecordLinkNotFound) {
		return RecordLinkCommandResult{}, false, nil
	}
	if err != nil {
		return RecordLinkCommandResult{}, false, err
	}
	updated, err := tombstoneRecordLinkStateTx(ctx, tx, existing.recordLinkID, command.ActorUserID, command.Now.UTC())
	if err != nil {
		return RecordLinkCommandResult{}, false, err
	}
	mutation := newRecordLinkMutation("delete", &existing, &updated)
	return newRecordLinkCommandResult(updated, &mutation), true, nil
}

type recordLinkInsert struct {
	incidentID      uuid.UUID
	srcRecordID     uuid.UUID
	dstRecordID     uuid.UUID
	linkType        LinkType
	fieldKey        *string
	provenance      LinkProvenance
	confidence      *int
	ownerUserID     uuid.UUID
	createdByUserID uuid.UUID
	decidedAt       time.Time
	createdAt       time.Time
}

func insertRecordLinkStateTx(ctx context.Context, tx pgx.Tx, input recordLinkInsert) (recordLinkState, error) {
	row := tx.QueryRow(ctx, `
INSERT INTO record_links (
    incident_id, src_record_id, dst_record_id, link_type, field_key,
    provenance, confidence, owner_user_id, created_by_user_id, decided_at,
    created_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING
    record_link_id, incident_id, src_record_id, dst_record_id, link_type,
    field_key, provenance, confidence, owner_user_id, created_by_user_id,
    decided_at, created_at, deleted_at, deleted_by_user_id
`, input.incidentID, input.srcRecordID, input.dstRecordID, input.linkType.String(), input.fieldKey,
		input.provenance.String(), input.confidence, input.ownerUserID, input.createdByUserID,
		input.decidedAt.UTC(), input.createdAt.UTC())
	return scanRecordLinkState(row)
}

func getActiveRecordLinkStateTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, srcRecordID uuid.UUID, dstRecordID uuid.UUID, linkType LinkType) (recordLinkState, error) {
	row := tx.QueryRow(ctx, `
SELECT
    record_link_id, incident_id, src_record_id, dst_record_id, link_type,
    field_key, provenance, confidence, owner_user_id, created_by_user_id,
    decided_at, created_at, deleted_at, deleted_by_user_id
  FROM record_links
 WHERE incident_id = $1
   AND src_record_id = $2
   AND dst_record_id = $3
   AND link_type = $4
   AND field_key IS NULL
   AND deleted_at IS NULL
 ORDER BY created_at DESC, record_link_id DESC
 LIMIT 1
 FOR UPDATE
`, incidentID, srcRecordID, dstRecordID, linkType.String())
	state, err := scanRecordLinkState(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return recordLinkState{}, errRecordLinkNotFound
	}
	if err != nil {
		return recordLinkState{}, fmt.Errorf("get active link state: %w", err)
	}
	return state, nil
}

func updateRecordLinkMetadataTx(ctx context.Context, tx pgx.Tx, recordLinkID uuid.UUID, provenance LinkProvenance, confidence *int, ownerUserID uuid.UUID, decidedAt time.Time) (recordLinkState, error) {
	row := tx.QueryRow(ctx, `
UPDATE record_links
   SET provenance = $2,
       confidence = $3,
       owner_user_id = $4,
       decided_at = $5
 WHERE record_link_id = $1
   AND deleted_at IS NULL
RETURNING
    record_link_id, incident_id, src_record_id, dst_record_id, link_type,
    field_key, provenance, confidence, owner_user_id, created_by_user_id,
    decided_at, created_at, deleted_at, deleted_by_user_id
`, recordLinkID, provenance.String(), confidence, ownerUserID, decidedAt.UTC())
	state, err := scanRecordLinkState(row)
	if err != nil {
		return recordLinkState{}, fmt.Errorf("update link metadata: %w", err)
	}
	return state, nil
}

func tombstoneRecordLinkStateTx(ctx context.Context, tx pgx.Tx, recordLinkID uuid.UUID, actorUserID uuid.UUID, deletedAt time.Time) (recordLinkState, error) {
	row := tx.QueryRow(ctx, `
UPDATE record_links
   SET deleted_at = $2,
       deleted_by_user_id = $3
 WHERE record_link_id = $1
   AND deleted_at IS NULL
RETURNING
    record_link_id, incident_id, src_record_id, dst_record_id, link_type,
    field_key, provenance, confidence, owner_user_id, created_by_user_id,
    decided_at, created_at, deleted_at, deleted_by_user_id
`, recordLinkID, deletedAt.UTC(), actorUserID)
	state, err := scanRecordLinkState(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return recordLinkState{}, errRecordLinkNotFound
	}
	if err != nil {
		return recordLinkState{}, fmt.Errorf("tombstone link: %w", err)
	}
	return state, nil
}

func scanRecordLinkState(row pgx.Row) (recordLinkState, error) {
	var (
		state           recordLinkState
		linkType        string
		fieldKey        pgtype.Text
		provenance      string
		confidence      pgtype.Int4
		deletedAt       pgtype.Timestamptz
		deletedByUserID pgtype.UUID
	)
	if err := row.Scan(
		&state.recordLinkID,
		&state.incidentID,
		&state.srcRecordID,
		&state.dstRecordID,
		&linkType,
		&fieldKey,
		&provenance,
		&confidence,
		&state.ownerUserID,
		&state.createdByUserID,
		&state.decidedAt,
		&state.createdAt,
		&deletedAt,
		&deletedByUserID,
	); err != nil {
		return recordLinkState{}, err
	}
	state.linkType = LinkType(linkType)
	state.provenance = LinkProvenance(provenance)
	if fieldKey.Valid {
		value := fieldKey.String
		state.fieldKey = &value
	}
	if confidence.Valid {
		value := int(confidence.Int32)
		state.confidence = &value
	}
	state.decidedAt = state.decidedAt.UTC()
	state.createdAt = state.createdAt.UTC()
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

func (s recordLinkState) mutationValue() valuecodec.RecordLinkMutationValue {
	return valuecodec.BuildRecordLinkMutationValue(valuecodec.RecordLinkMutationInput{
		RecordLinkID:    s.recordLinkID,
		IncidentID:      s.incidentID,
		SrcRecordID:     s.srcRecordID,
		DstRecordID:     s.dstRecordID,
		LinkType:        s.linkType.String(),
		FieldKey:        copyStringPointer(s.fieldKey),
		Provenance:      s.provenance.String(),
		Confidence:      copyIntPointer(s.confidence),
		OwnerUserID:     s.ownerUserID,
		CreatedByUserID: s.createdByUserID,
		DecidedAt:       s.decidedAt,
		CreatedAt:       s.createdAt,
		DeletedAt:       copyTimePointer(s.deletedAt),
		DeletedByUserID: copyUUIDPointer(s.deletedByUserID),
	})
}

func newRecordLinkMutation(operation string, before *recordLinkState, after *recordLinkState) RecordLinkMutation {
	mutation := RecordLinkMutation{Operation: operation}
	if before != nil {
		mutation.RecordLinkID = before.recordLinkID
		mutation.BeforeValue = before.mutationValue().Map()
	}
	if after != nil {
		mutation.RecordLinkID = after.recordLinkID
		mutation.AfterValue = after.mutationValue().Map()
	}
	return mutation
}

func newRecordLinkCommandResult(state recordLinkState, mutation *RecordLinkMutation) RecordLinkCommandResult {
	result := RecordLinkCommandResult{
		RecordLinkID: state.recordLinkID,
		SrcRecordID:  state.srcRecordID,
		DstRecordID:  state.dstRecordID,
		LinkType:     state.linkType,
	}
	if mutation != nil {
		cloned := cloneRecordLinkMutation(*mutation)
		result.Mutation = &cloned
	}
	return result
}

func cloneRecordLinkMutation(mutation RecordLinkMutation) RecordLinkMutation {
	return RecordLinkMutation{
		RecordLinkID: mutation.RecordLinkID,
		Operation:    mutation.Operation,
		BeforeValue:  cloneMutationMap(mutation.BeforeValue),
		AfterValue:   cloneMutationMap(mutation.AfterValue),
	}
}

func cloneMutationMap(value map[string]any) map[string]any {
	if value == nil {
		return nil
	}
	cloned := make(map[string]any, len(value))
	for key, item := range value {
		switch typed := item.(type) {
		case map[string]any:
			cloned[key] = cloneMutationMap(typed)
		case []any:
			items := make([]any, len(typed))
			for index, nested := range typed {
				if nestedMap, ok := nested.(map[string]any); ok {
					items[index] = cloneMutationMap(nestedMap)
				} else {
					items[index] = nested
				}
			}
			cloned[key] = items
		default:
			cloned[key] = item
		}
	}
	return cloned
}

func copyStringPointer(value *string) *string {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}

func copyIntPointer(value *int) *int {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}

func copyTimePointer(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	copy := value.UTC()
	return &copy
}

func copyUUIDPointer(value *uuid.UUID) *uuid.UUID {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}
