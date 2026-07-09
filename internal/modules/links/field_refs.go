package links

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (s *Store) UpsertFieldReferenceTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, src uuid.UUID, dst uuid.UUID, fieldKey string, linkType string, actorID uuid.UUID, now time.Time) (bool, error) {
	_, inserted, err := s.UpsertFieldReferenceRecordTx(ctx, tx, incidentID, src, dst, fieldKey, linkType, actorID, now)
	return inserted, err
}

func (s *Store) UpsertFieldReferenceRecordTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, src uuid.UUID, dst uuid.UUID, fieldKey string, linkType string, actorID uuid.UUID, now time.Time) (RecordLink, bool, error) {
	if err := validateRecordLinkCommand(linkType, LinkProvenanceManual, nil, src, dst); err != nil {
		return RecordLink{}, false, err
	}
	if err := validateActiveLinkEndpointsTx(ctx, tx, incidentID, src, dst); err != nil {
		return RecordLink{}, false, err
	}
	row := tx.QueryRow(ctx, `
INSERT INTO record_links (
    incident_id, src_record_id, dst_record_id, link_type, field_key,
    provenance, confidence, owner_user_id, created_by_user_id, decided_at, created_at
) VALUES ($1, $2, $3, $7, $4, 'manual', NULL, $5, $5, $6, $6)
ON CONFLICT (incident_id, src_record_id, dst_record_id, link_type, field_key)
WHERE deleted_at IS NULL AND field_key IS NOT NULL
DO NOTHING
RETURNING
    record_link_id,
    incident_id,
    src_record_id,
    dst_record_id,
    link_type,
    provenance,
    confidence,
    owner_user_id,
    decided_at,
    created_at,
    deleted_at
`, incidentID, src, dst, fieldKey, actorID, now, linkType)
	record, err := scanRecordLink(row)
	if err == nil {
		return record, true, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return RecordLink{}, false, fmt.Errorf("upsert reference link: %w", err)
	}

	record, err = s.GetActiveFieldReferenceTx(ctx, tx, incidentID, src, dst, fieldKey, linkType)
	if err != nil {
		return RecordLink{}, false, err
	}
	return record, false, nil
}

func (s *Store) GetActiveFieldReferenceTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, src uuid.UUID, dst uuid.UUID, fieldKey string, linkType string) (RecordLink, error) {
	row := tx.QueryRow(ctx, `
SELECT
    record_link_id,
    incident_id,
    src_record_id,
    dst_record_id,
    link_type,
    provenance,
    confidence,
    owner_user_id,
    decided_at,
    created_at,
    deleted_at
  FROM record_links
 WHERE incident_id = $1
   AND src_record_id = $2
   AND dst_record_id = $3
   AND link_type = $4
   AND field_key = $5
   AND deleted_at IS NULL
 ORDER BY created_at DESC, record_link_id DESC
 LIMIT 1
 FOR UPDATE
`, incidentID, src, dst, linkType, fieldKey)
	record, err := scanRecordLink(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return RecordLink{}, ErrFieldReferenceNotFound
	}
	if err != nil {
		return RecordLink{}, fmt.Errorf("get active reference link: %w", err)
	}
	return record, nil
}

func (s *Store) SyncFieldReferenceTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, src uuid.UUID, targetID *uuid.UUID, fieldKey string, linkType string, actorID uuid.UUID, now time.Time) (bool, error) {
	return s.SyncFieldReferenceCommandTx(ctx, tx, SyncFieldReferenceCommand{
		IncidentID:  incidentID,
		SrcRecordID: src,
		TargetID:    targetID,
		FieldKey:    fieldKey,
		LinkType:    LinkType(linkType),
		ActorUserID: actorID,
		Now:         now,
	})
}

func (s *Store) SyncFieldReferenceCommandTx(ctx context.Context, tx pgx.Tx, command SyncFieldReferenceCommand) (bool, error) {
	changed := false
	now := command.Now.UTC()
	linkType := command.LinkType.String()
	args := []any{command.IncidentID, command.SrcRecordID, command.FieldKey, linkType, command.ActorUserID, now}
	keepPredicate := ""
	if command.TargetID != nil {
		args = append(args, *command.TargetID)
		keepPredicate = "AND dst_record_id <> $7"
	}
	tag, err := tx.Exec(ctx, `
UPDATE record_links
   SET deleted_at = $6,
       deleted_by_user_id = $5
 WHERE incident_id = $1
   AND src_record_id = $2
   AND field_key = $3
   AND link_type = $4
   AND deleted_at IS NULL
   `+keepPredicate, args...)
	if err != nil {
		return false, fmt.Errorf("sync field reference link: %w", err)
	}
	changed = changed || tag.RowsAffected() > 0
	if command.TargetID == nil {
		return changed, nil
	}
	inserted, err := s.UpsertFieldReferenceTx(ctx, tx, command.IncidentID, command.SrcRecordID, *command.TargetID, command.FieldKey, linkType, command.ActorUserID, now)
	if err != nil {
		return false, err
	}
	return changed || inserted, nil
}

func (s *Store) InsertLinkedNoteReferenceTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, src uuid.UUID, dst uuid.UUID, actorID uuid.UUID, now time.Time) (RecordLink, bool, error) {
	if err := validateRecordLinkCommand(LinkTypeReferencesArtifact, LinkProvenanceManual, nil, src, dst); err != nil {
		return RecordLink{}, false, err
	}
	if err := validateActiveLinkEndpointsTx(ctx, tx, incidentID, src, dst); err != nil {
		return RecordLink{}, false, err
	}
	row := tx.QueryRow(ctx, `
INSERT INTO record_links (
    incident_id, src_record_id, dst_record_id, link_type, field_key,
    provenance, confidence, owner_user_id, created_by_user_id, decided_at, created_at
) VALUES ($1, $2, $3, 'references_artifact', NULL, 'manual', NULL, $4, $4, $5, $5)
ON CONFLICT (incident_id, src_record_id, dst_record_id, link_type)
WHERE deleted_at IS NULL AND field_key IS NULL
DO NOTHING
RETURNING
    record_link_id,
    incident_id,
    src_record_id,
    dst_record_id,
    link_type,
    provenance,
    confidence,
    owner_user_id,
    decided_at,
    created_at,
    deleted_at
`, incidentID, src, dst, actorID, now)
	record, err := scanRecordLink(row)
	if err == nil {
		return record, true, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return RecordLink{}, false, fmt.Errorf("insert linked note reference: %w", err)
	}
	record, err = s.GetActiveLinkTx(ctx, tx, incidentID, src, dst, "references_artifact")
	if err != nil {
		return RecordLink{}, false, err
	}
	return record, false, nil
}

func (s *Store) TombstoneFieldReferenceTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, src uuid.UUID, dst uuid.UUID, fieldKey string, linkType string, expectedTargetType string, actorID uuid.UUID, now time.Time) (bool, error) {
	_, err := s.TombstoneFieldReferenceRecordTx(ctx, tx, incidentID, src, dst, fieldKey, linkType, expectedTargetType, actorID, now)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (s *Store) TombstoneFieldReferenceRecordTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, src uuid.UUID, dst uuid.UUID, fieldKey string, linkType string, expectedTargetType string, actorID uuid.UUID, now time.Time) (RecordLink, error) {
	targetTypePredicate := ""
	args := []any{incidentID, src, dst, fieldKey, actorID, now, linkType}
	if expectedTargetType != "" {
		targetTypePredicate = "AND dst.record_type = $8"
		args = append(args, expectedTargetType)
	}
	row := tx.QueryRow(ctx, `
UPDATE record_links
   SET deleted_at = $6,
       deleted_by_user_id = $5
 WHERE incident_id = $1
   AND src_record_id = $2
   AND dst_record_id = $3
   AND link_type = $7
   AND field_key = $4
   AND deleted_at IS NULL
   AND EXISTS (
       SELECT 1
         FROM records dst
        WHERE dst.incident_id = record_links.incident_id
          AND dst.record_id = record_links.dst_record_id
          AND dst.deleted_at IS NULL
          `+targetTypePredicate+`
   )
RETURNING
    record_link_id,
    incident_id,
    src_record_id,
    dst_record_id,
    link_type,
    provenance,
    confidence,
    owner_user_id,
    decided_at,
    created_at,
    deleted_at
`, args...)
	record, err := scanRecordLink(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return RecordLink{}, ErrFieldReferenceNotFound
	}
	if err != nil {
		return RecordLink{}, fmt.Errorf("remove reference link: %w", err)
	}
	return record, nil
}

type TagStore struct{}

func NewTagStore() *TagStore {
	return &TagStore{}
}

func (s *Store) Tags() *TagStore {
	return NewTagStore()
}

func (s *TagStore) UpsertTagTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordID uuid.UUID, tagName string, normalizedTagName string, actorID uuid.UUID, now time.Time) (bool, error) {
	_, inserted, err := s.UpsertTagRecordTx(ctx, tx, incidentID, recordID, tagName, normalizedTagName, actorID, now)
	return inserted, err
}

func (s *TagStore) UpsertTagRecordTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordID uuid.UUID, tagName string, normalizedTagName string, actorID uuid.UUID, now time.Time) (uuid.UUID, bool, error) {
	if tagName == "" || normalizedTagName == "" {
		return uuid.Nil, false, ErrInvalidTag
	}
	var tagID uuid.UUID
	err := tx.QueryRow(ctx, `
INSERT INTO record_tags (
    incident_id, record_id, tag_name, normalized_tag_name,
    created_by_user_id, created_at, updated_at
) VALUES ($1, $2, $3, $4, $5, $6, $6)
ON CONFLICT (incident_id, record_id, normalized_tag_name)
WHERE deleted_at IS NULL
DO NOTHING
RETURNING record_tag_id
`, incidentID, recordID, tagName, normalizedTagName, actorID, now).Scan(&tagID)
	if err == nil {
		return tagID, true, nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, false, nil
	}
	if err != nil {
		return uuid.Nil, false, fmt.Errorf("upsert record tag: %w", err)
	}
	return tagID, true, nil
}

func (s *TagStore) TombstoneTagTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordID uuid.UUID, tagID uuid.UUID, actorID uuid.UUID, now time.Time) (bool, error) {
	_, err := s.TombstoneTagRecordTx(ctx, tx, incidentID, recordID, tagID, actorID, now)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (s *TagStore) TombstoneTagRecordTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordID uuid.UUID, tagID uuid.UUID, actorID uuid.UUID, now time.Time) (uuid.UUID, error) {
	var deleted uuid.UUID
	err := tx.QueryRow(ctx, `
UPDATE record_tags
   SET deleted_at = $5,
       deleted_by_user_id = $4,
       updated_at = $5
 WHERE incident_id = $1
   AND record_id = $2
   AND record_tag_id = $3
   AND deleted_at IS NULL
RETURNING record_tag_id
`, incidentID, recordID, tagID, actorID, now).Scan(&deleted)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, ErrTagNotFound
	}
	if err != nil {
		return uuid.Nil, fmt.Errorf("remove record tag: %w", err)
	}
	return deleted, nil
}
