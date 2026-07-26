package links

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/JochiRaider/cartulary/internal/modules/timeline/workbookprojection"
)

type TimelineFacts struct {
	ResolvedLinks       []workbookprojection.LinkFact
	Tags                []workbookprojection.TagFact
	AttachedEvidenceIDs []uuid.UUID
	ReplacementRecordID *uuid.UUID
}

type TimelineFactReader struct{}

func (TimelineFactReader) LoadTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordID uuid.UUID) (TimelineFacts, error) {
	result := TimelineFacts{
		ResolvedLinks:       []workbookprojection.LinkFact{},
		Tags:                []workbookprojection.TagFact{},
		AttachedEvidenceIDs: []uuid.UUID{},
	}
	rows, err := tx.Query(ctx, `
SELECT dst_record_id, link_type, provenance, confidence
  FROM active_record_links_v1
 WHERE incident_id = $1
   AND src_record_id = $2
   AND link_type IN ('observed_on_host', 'observed_as_identity', 'attached_evidence')
 ORDER BY link_type ASC, created_at ASC, record_link_id ASC
`, incidentID, recordID)
	if err != nil {
		return TimelineFacts{}, fmt.Errorf("load timeline link facts: %w", err)
	}
	for rows.Next() {
		var targetID uuid.UUID
		var linkType string
		var provenance string
		var confidence pgtype.Int4
		if err := rows.Scan(&targetID, &linkType, &provenance, &confidence); err != nil {
			rows.Close()
			return TimelineFacts{}, fmt.Errorf("scan timeline link fact: %w", err)
		}
		if linkType == "attached_evidence" {
			result.AttachedEvidenceIDs = append(result.AttachedEvidenceIDs, targetID)
			continue
		}
		fact := workbookprojection.LinkFact{
			TargetRecordID: targetID,
			LinkType:       linkType,
			Provenance:     provenance,
		}
		if confidence.Valid {
			value := int(confidence.Int32)
			fact.Confidence = &value
		}
		result.ResolvedLinks = append(result.ResolvedLinks, fact)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return TimelineFacts{}, fmt.Errorf("iterate timeline link facts: %w", err)
	}
	rows.Close()

	tagRows, err := tx.Query(ctx, `
SELECT record_tag_id, tag_name
  FROM active_record_tags_v1
 WHERE incident_id = $1
   AND record_id = $2
 ORDER BY normalized_tag_name ASC, record_tag_id ASC
`, incidentID, recordID)
	if err != nil {
		return TimelineFacts{}, fmt.Errorf("load timeline tag facts: %w", err)
	}
	for tagRows.Next() {
		var fact workbookprojection.TagFact
		if err := tagRows.Scan(&fact.RecordTagID, &fact.TagName); err != nil {
			tagRows.Close()
			return TimelineFacts{}, fmt.Errorf("scan timeline tag fact: %w", err)
		}
		result.Tags = append(result.Tags, fact)
	}
	if err := tagRows.Err(); err != nil {
		tagRows.Close()
		return TimelineFacts{}, fmt.Errorf("iterate timeline tag facts: %w", err)
	}
	tagRows.Close()

	var replacementID uuid.UUID
	err = tx.QueryRow(ctx, `
SELECT src_record_id
  FROM active_record_links_v1
 WHERE incident_id = $1
   AND dst_record_id = $2
   AND link_type = 'supersedes'
 ORDER BY created_at DESC, record_link_id DESC
 LIMIT 1
`, incidentID, recordID).Scan(&replacementID)
	switch err {
	case nil:
		result.ReplacementRecordID = &replacementID
	case pgx.ErrNoRows:
	default:
		return TimelineFacts{}, fmt.Errorf("load timeline replacement fact: %w", err)
	}
	return result, nil
}
