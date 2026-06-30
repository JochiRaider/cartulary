package links

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type Store struct{}

var ErrRecordLinkNotFound = errors.New("links: record link not found")
var ErrFieldReferenceNotFound = errors.New("links: field reference not found")
var ErrRiskRefNotFound = errors.New("links: risk ref not found")
var ErrTagNotFound = errors.New("links: tag not found")
var ErrInvalidTag = errors.New("links: invalid tag")

type RecordLink struct {
	RecordLinkID uuid.UUID
	IncidentID   uuid.UUID
	SrcRecordID  uuid.UUID
	DstRecordID  uuid.UUID
	LinkType     string
	Provenance   string
	Confidence   *int
	OwnerUserID  uuid.UUID
	DecidedAt    time.Time
	CreatedAt    time.Time
	DeletedAt    *time.Time
}

type SupersedesLink struct {
	RecordLinkID uuid.UUID
	IncidentID   uuid.UUID
	SrcRecordID  uuid.UUID
	DstRecordID  uuid.UUID
}

func NewStore() *Store {
	return &Store{}
}

func (s *Store) GetActiveLinkTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, srcRecordID uuid.UUID, dstRecordID uuid.UUID, linkType string) (RecordLink, error) {
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
   AND deleted_at IS NULL
 ORDER BY created_at DESC, record_link_id DESC
 LIMIT 1
 FOR UPDATE
`, incidentID, srcRecordID, dstRecordID, linkType)
	record, err := scanRecordLink(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return RecordLink{}, ErrRecordLinkNotFound
	}
	if err != nil {
		return RecordLink{}, fmt.Errorf("get active link: %w", err)
	}
	return record, nil
}

func (s *Store) UpsertLinkTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, srcRecordID uuid.UUID, dstRecordID uuid.UUID, linkType string, provenance string, confidence *int, ownerUserID uuid.UUID, now time.Time) (RecordLink, bool, error) {
	existing, err := s.GetActiveLinkTx(ctx, tx, incidentID, srcRecordID, dstRecordID, linkType)
	if err == nil {
		if existing.Provenance == provenance && intPointersEqual(existing.Confidence, confidence) {
			return existing, false, nil
		}
		row := tx.QueryRow(ctx, `
UPDATE record_links
   SET provenance = $2,
       confidence = $3,
       owner_user_id = $4,
       decided_at = $5
 WHERE record_link_id = $1
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
`, existing.RecordLinkID, provenance, confidence, ownerUserID, now.UTC())
		record, err := scanRecordLink(row)
		if err != nil {
			return RecordLink{}, false, fmt.Errorf("update link: %w", err)
		}
		return record, false, nil
	}
	if !errors.Is(err, ErrRecordLinkNotFound) {
		return RecordLink{}, false, err
	}

	row := tx.QueryRow(ctx, `
INSERT INTO record_links (
    incident_id,
    src_record_id,
    dst_record_id,
    link_type,
    provenance,
    confidence,
    owner_user_id,
    created_by_user_id,
    decided_at,
    created_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $7, $8, $8)
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
`, incidentID, srcRecordID, dstRecordID, linkType, provenance, confidence, ownerUserID, now.UTC())
	record, err := scanRecordLink(row)
	if err != nil {
		return RecordLink{}, false, fmt.Errorf("insert link: %w", err)
	}
	return record, true, nil
}

func (s *Store) TombstoneLinkTx(ctx context.Context, tx pgx.Tx, recordLinkID uuid.UUID, actorUserID uuid.UUID, now time.Time) (RecordLink, error) {
	row := tx.QueryRow(ctx, `
UPDATE record_links
   SET deleted_at = COALESCE(deleted_at, $2),
       deleted_by_user_id = COALESCE(deleted_by_user_id, $3)
 WHERE record_link_id = $1
   AND deleted_at IS NULL
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
`, recordLinkID, now.UTC(), actorUserID)
	record, err := scanRecordLink(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return RecordLink{}, ErrRecordLinkNotFound
	}
	if err != nil {
		return RecordLink{}, fmt.Errorf("tombstone link: %w", err)
	}
	return record, nil
}

func (s *Store) InsertSupersedesTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, replacementRecordID uuid.UUID, supersededRecordID uuid.UUID, ownerUserID uuid.UUID, now time.Time) (SupersedesLink, error) {
	var link SupersedesLink
	if err := tx.QueryRow(ctx, `
INSERT INTO record_links (
    incident_id,
    src_record_id,
    dst_record_id,
    link_type,
    provenance,
    confidence,
    owner_user_id,
    created_by_user_id,
    decided_at,
    created_at
)
VALUES ($1, $2, $3, 'supersedes', 'manual', NULL, $4, $4, $5, $5)
RETURNING record_link_id, incident_id, src_record_id, dst_record_id
`, incidentID, replacementRecordID, supersededRecordID, ownerUserID, now.UTC()).Scan(&link.RecordLinkID, &link.IncidentID, &link.SrcRecordID, &link.DstRecordID); err != nil {
		return SupersedesLink{}, fmt.Errorf("insert supersedes link: %w", err)
	}
	return link, nil
}

func scanRecordLink(row pgx.Row) (RecordLink, error) {
	var (
		record     RecordLink
		confidence pgtype.Int4
		deletedAt  pgtype.Timestamptz
	)
	if err := row.Scan(
		&record.RecordLinkID,
		&record.IncidentID,
		&record.SrcRecordID,
		&record.DstRecordID,
		&record.LinkType,
		&record.Provenance,
		&confidence,
		&record.OwnerUserID,
		&record.DecidedAt,
		&record.CreatedAt,
		&deletedAt,
	); err != nil {
		return RecordLink{}, err
	}
	if confidence.Valid {
		value := int(confidence.Int32)
		record.Confidence = &value
	}
	if deletedAt.Valid {
		value := deletedAt.Time.UTC()
		record.DeletedAt = &value
	}
	record.DecidedAt = record.DecidedAt.UTC()
	record.CreatedAt = record.CreatedAt.UTC()
	return record, nil
}

func intPointersEqual(left *int, right *int) bool {
	switch {
	case left == nil && right == nil:
		return true
	case left == nil || right == nil:
		return false
	default:
		return *left == *right
	}
}
