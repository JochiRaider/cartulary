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
var ErrTagNotFound = errors.New("links: tag not found")
var ErrInvalidTag = errors.New("links: invalid tag")
var ErrInvalidRecordLink = errors.New("links: invalid record link")

const (
	LinkTypeObservedOnHost      = "observed_on_host"
	LinkTypeObservedAsIdentity  = "observed_as_identity"
	LinkTypeReferencesIndicator = "references_indicator"
	LinkTypeAttachedEvidence    = "attached_evidence"
	LinkTypeReferencesArtifact  = "references_artifact"
	LinkTypeDerivedFrom         = "derived_from"
	LinkTypeMergedInto          = "merged_into"
	LinkTypeSupportedBy         = "supported_by"
	LinkTypeReferencesRecord    = "references_record"
	LinkTypeSupersedes          = "supersedes"

	LinkProvenanceManual    = "manual"
	LinkProvenanceAutoMatch = "auto_match"
)

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
   AND field_key IS NULL
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
	return s.UpsertLinkCommandTx(ctx, tx, UpsertLinkCommand{
		IncidentID:  incidentID,
		SrcRecordID: srcRecordID,
		DstRecordID: dstRecordID,
		LinkType:    LinkType(linkType),
		Provenance:  LinkProvenance(provenance),
		Confidence:  confidence,
		OwnerUserID: ownerUserID,
		Now:         now,
	})
}

func (s *Store) UpsertLinkCommandTx(ctx context.Context, tx pgx.Tx, command UpsertLinkCommand) (RecordLink, bool, error) {
	linkType := command.LinkType.String()
	provenance := command.Provenance.String()
	now := command.Now.UTC()
	if err := validateRecordLinkCommand(linkType, provenance, command.Confidence, command.SrcRecordID, command.DstRecordID); err != nil {
		return RecordLink{}, false, err
	}
	if err := validateActiveLinkEndpointsTx(ctx, tx, command.IncidentID, command.SrcRecordID, command.DstRecordID); err != nil {
		return RecordLink{}, false, err
	}
	existing, err := s.GetActiveLinkTx(ctx, tx, command.IncidentID, command.SrcRecordID, command.DstRecordID, linkType)
	if err == nil {
		if existing.Provenance == provenance && intPointersEqual(existing.Confidence, command.Confidence) {
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
`, existing.RecordLinkID, provenance, command.Confidence, command.OwnerUserID, now)
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
`, command.IncidentID, command.SrcRecordID, command.DstRecordID, linkType, provenance, command.Confidence, command.OwnerUserID, now)
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
	return s.InsertSupersedesCommandTx(ctx, tx, InsertSupersedesCommand{
		IncidentID:          incidentID,
		ReplacementRecordID: replacementRecordID,
		SupersededRecordID:  supersededRecordID,
		OwnerUserID:         ownerUserID,
		Now:                 now,
	})
}

func (s *Store) InsertSupersedesCommandTx(ctx context.Context, tx pgx.Tx, command InsertSupersedesCommand) (SupersedesLink, error) {
	now := command.Now.UTC()
	if err := validateRecordLinkCommand(LinkTypeSupersedes, LinkProvenanceManual, nil, command.ReplacementRecordID, command.SupersededRecordID); err != nil {
		return SupersedesLink{}, err
	}
	if err := validateSupersedesEndpointsTx(ctx, tx, command.IncidentID, command.ReplacementRecordID, command.SupersededRecordID); err != nil {
		return SupersedesLink{}, err
	}
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
`, command.IncidentID, command.ReplacementRecordID, command.SupersededRecordID, command.OwnerUserID, now).Scan(&link.RecordLinkID, &link.IncidentID, &link.SrcRecordID, &link.DstRecordID); err != nil {
		return SupersedesLink{}, fmt.Errorf("insert supersedes link: %w", err)
	}
	return link, nil
}

func (s *Store) HasActiveIncomingSupersedesLinkForUpdateTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordID uuid.UUID) (bool, error) {
	var linkID uuid.UUID
	if err := tx.QueryRow(ctx, `
SELECT record_link_id
  FROM record_links
 WHERE incident_id = $1
   AND dst_record_id = $2
   AND link_type = 'supersedes'
   AND deleted_at IS NULL
 ORDER BY created_at DESC, record_link_id DESC
 LIMIT 1
 FOR UPDATE
`, incidentID, recordID).Scan(&linkID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("query active incoming supersedes link: %w", err)
	}
	return true, nil
}

func validateRecordLinkCommand(linkType string, provenance string, confidence *int, srcRecordID uuid.UUID, dstRecordID uuid.UUID) error {
	if srcRecordID == uuid.Nil || dstRecordID == uuid.Nil || srcRecordID == dstRecordID {
		return fmt.Errorf("%w: invalid endpoints", ErrInvalidRecordLink)
	}
	if !isKnownLinkType(linkType) {
		return fmt.Errorf("%w: unsupported link type %q", ErrInvalidRecordLink, linkType)
	}
	switch provenance {
	case LinkProvenanceManual:
		if confidence != nil {
			return fmt.Errorf("%w: manual links must not carry confidence", ErrInvalidRecordLink)
		}
	case LinkProvenanceAutoMatch:
		if linkType != LinkTypeObservedOnHost && linkType != LinkTypeObservedAsIdentity {
			return fmt.Errorf("%w: auto_match provenance is only valid for entity observation links", ErrInvalidRecordLink)
		}
		if confidence == nil || *confidence < 0 || *confidence > 100 {
			return fmt.Errorf("%w: auto_match confidence must be between 0 and 100", ErrInvalidRecordLink)
		}
	default:
		return fmt.Errorf("%w: unsupported provenance %q", ErrInvalidRecordLink, provenance)
	}
	return nil
}

func isKnownLinkType(linkType string) bool {
	switch linkType {
	case LinkTypeObservedOnHost,
		LinkTypeObservedAsIdentity,
		LinkTypeReferencesIndicator,
		LinkTypeAttachedEvidence,
		LinkTypeReferencesArtifact,
		LinkTypeDerivedFrom,
		LinkTypeMergedInto,
		LinkTypeSupportedBy,
		LinkTypeReferencesRecord,
		LinkTypeSupersedes:
		return true
	default:
		return false
	}
}

func validateActiveLinkEndpointsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, srcRecordID uuid.UUID, dstRecordID uuid.UUID) error {
	var count int
	if err := tx.QueryRow(ctx, `
SELECT count(*)
  FROM records
 WHERE incident_id = $1
   AND record_id IN ($2, $3)
   AND deleted_at IS NULL
`, incidentID, srcRecordID, dstRecordID).Scan(&count); err != nil {
		return fmt.Errorf("validate link endpoints: %w", err)
	}
	if count != 2 {
		return fmt.Errorf("%w: endpoints must be active same-incident records", ErrInvalidRecordLink)
	}
	return nil
}

func validateSupersedesEndpointsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, replacementRecordID uuid.UUID, supersededRecordID uuid.UUID) error {
	var srcType, dstType string
	if err := tx.QueryRow(ctx, `
SELECT src.record_type, dst.record_type
  FROM records src
  JOIN records dst
    ON dst.incident_id = src.incident_id
   AND dst.record_id = $3
   AND dst.deleted_at IS NULL
 WHERE src.incident_id = $1
   AND src.record_id = $2
   AND src.deleted_at IS NULL
`, incidentID, replacementRecordID, supersededRecordID).Scan(&srcType, &dstType); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("%w: supersedes endpoints must be active same-incident records", ErrInvalidRecordLink)
		}
		return fmt.Errorf("validate supersedes endpoints: %w", err)
	}
	if srcType != dstType || (srcType != "timeline_event" && srcType != "decision") {
		return fmt.Errorf("%w: supersedes endpoints must both be timeline events or both be decisions", ErrInvalidRecordLink)
	}
	return nil
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
