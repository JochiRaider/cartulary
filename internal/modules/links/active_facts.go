package links

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type RecordLinkFact struct {
	RecordLinkID    uuid.UUID
	SrcRecordID     uuid.UUID
	DstRecordID     uuid.UUID
	LinkType        string
	FieldKey        *string
	Provenance      string
	Confidence      *int
	OwnerUserID     uuid.UUID
	CreatedByUserID uuid.UUID
	DecidedAt       time.Time
	CreatedAt       time.Time
}

type RecordTagFact struct {
	RecordTagID       uuid.UUID
	RecordID          uuid.UUID
	TagName           string
	NormalizedTagName string
	CreatedByUserID   uuid.UUID
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type ActiveFacts struct {
	RecordLinks []RecordLinkFact
	RecordTags  []RecordTagFact
}

type ActiveFactsReadError struct {
	err error
}

func (e *ActiveFactsReadError) Error() string {
	return "links: active fact read failed"
}

func (e *ActiveFactsReadError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.err
}

type ActiveFactReader struct{}

func (ActiveFactReader) LoadTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) (ActiveFacts, error) {
	return loadActiveFactsTx(ctx, tx, incidentID, nil)
}

// RecordFactReader provides the same owner facts for one record scope without
// making callers query the entire incident when they need one record.
type RecordFactReader struct{}

func (RecordFactReader) LoadTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	recordID uuid.UUID,
) (ActiveFacts, error) {
	return loadActiveFactsTx(ctx, tx, incidentID, &recordID)
}

func loadActiveFactsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordID *uuid.UUID) (ActiveFacts, error) {
	facts := ActiveFacts{
		RecordLinks: []RecordLinkFact{},
		RecordTags:  []RecordTagFact{},
	}
	linkQuery := `
SELECT
    record_link_id,
    src_record_id,
    dst_record_id,
    link_type,
    field_key,
    provenance,
    confidence,
    owner_user_id,
    created_by_user_id,
    decided_at,
    created_at
  FROM active_record_links_v1
 WHERE incident_id = $1
`
	linkArgs := []any{incidentID}
	if recordID != nil {
		linkQuery += `
   AND (src_record_id = $2 OR (dst_record_id = $2 AND link_type = 'supersedes'))
`
		linkArgs = append(linkArgs, *recordID)
	}
	linkQuery += `
 ORDER BY src_record_id ASC, dst_record_id ASC, record_link_id ASC
`
	rows, err := tx.Query(ctx, linkQuery, linkArgs...)
	if err != nil {
		return ActiveFacts{}, &ActiveFactsReadError{err: err}
	}
	for rows.Next() {
		var (
			fact       RecordLinkFact
			fieldKey   pgtype.Text
			confidence pgtype.Int4
		)
		if err := rows.Scan(
			&fact.RecordLinkID,
			&fact.SrcRecordID,
			&fact.DstRecordID,
			&fact.LinkType,
			&fieldKey,
			&fact.Provenance,
			&confidence,
			&fact.OwnerUserID,
			&fact.CreatedByUserID,
			&fact.DecidedAt,
			&fact.CreatedAt,
		); err != nil {
			rows.Close()
			return ActiveFacts{}, &ActiveFactsReadError{err: err}
		}
		if fieldKey.Valid {
			value := fieldKey.String
			fact.FieldKey = &value
		}
		if confidence.Valid {
			value := int(confidence.Int32)
			fact.Confidence = &value
		}
		fact.DecidedAt = fact.DecidedAt.UTC()
		fact.CreatedAt = fact.CreatedAt.UTC()
		facts.RecordLinks = append(facts.RecordLinks, fact)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return ActiveFacts{}, &ActiveFactsReadError{err: err}
	}
	rows.Close()

	tagQuery := `
SELECT
    record_tag_id,
    record_id,
    tag_name,
    normalized_tag_name,
    created_by_user_id,
    created_at,
    updated_at
  FROM active_record_tags_v1
 WHERE incident_id = $1
`
	tagArgs := []any{incidentID}
	if recordID != nil {
		tagQuery += `
   AND record_id = $2
`
		tagArgs = append(tagArgs, *recordID)
	}
	tagQuery += `
 ORDER BY record_tag_id ASC
`
	tagRows, err := tx.Query(ctx, tagQuery, tagArgs...)
	if err != nil {
		return ActiveFacts{}, &ActiveFactsReadError{err: err}
	}
	for tagRows.Next() {
		var fact RecordTagFact
		if err := tagRows.Scan(
			&fact.RecordTagID,
			&fact.RecordID,
			&fact.TagName,
			&fact.NormalizedTagName,
			&fact.CreatedByUserID,
			&fact.CreatedAt,
			&fact.UpdatedAt,
		); err != nil {
			tagRows.Close()
			return ActiveFacts{}, &ActiveFactsReadError{err: err}
		}
		fact.CreatedAt = fact.CreatedAt.UTC()
		fact.UpdatedAt = fact.UpdatedAt.UTC()
		facts.RecordTags = append(facts.RecordTags, fact)
	}
	if err := tagRows.Err(); err != nil {
		tagRows.Close()
		return ActiveFacts{}, &ActiveFactsReadError{err: err}
	}
	tagRows.Close()
	return facts, nil
}
