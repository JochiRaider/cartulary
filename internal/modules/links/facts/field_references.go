package facts

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/links/vocabulary"
)

type FieldReferenceFact struct {
	DestinationRecordID uuid.UUID
	LinkType            vocabulary.LinkType
}

// LoadFieldReferencesTx reads materialized bindings, including those whose
// endpoints are deleted. Only the binding's own tombstone removes it here.
func (FactReader) LoadFieldReferencesTx(ctx context.Context, tx pgx.Tx, incidentID, sourceID uuid.UUID, fieldKey string) ([]FieldReferenceFact, error) {
	rows, err := tx.Query(ctx, `
SELECT dst_record_id, link_type
  FROM record_links
 WHERE incident_id=$1 AND src_record_id=$2 AND field_key=$3 AND deleted_at IS NULL
 ORDER BY record_link_id
`, incidentID, sourceID, fieldKey)
	if err != nil {
		return nil, &FactReadError{err: err}
	}
	defer rows.Close()
	facts := make([]FieldReferenceFact, 0)
	for rows.Next() {
		var fact FieldReferenceFact
		var kind string
		if err := rows.Scan(&fact.DestinationRecordID, &kind); err != nil {
			return nil, &FactReadError{err: err}
		}
		parsed, err := vocabulary.ParseLinkType(kind)
		if err != nil {
			return nil, &FactReadError{err: err}
		}
		fact.LinkType = parsed
		facts = append(facts, fact)
	}
	if err := rows.Err(); err != nil {
		return nil, &FactReadError{err: err}
	}
	return facts, nil
}
