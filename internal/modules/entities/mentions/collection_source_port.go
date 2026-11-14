package mentions

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// CollectionSourceFact preserves current source facts, including hidden dismissed
// Mentions. Reading these facts does not authorize a lifecycle transition.
type CollectionSourceFact struct {
	MentionID        uuid.UUID
	EntityType       string
	RawText          string
	ResolutionStatus string
	RowVersion       int64
	Ordinal          int32
	ResolvedRecordID *uuid.UUID
}

func (*Store) LoadCollectionSourceFactsTx(ctx context.Context, tx pgx.Tx, sourceID uuid.UUID, field string) ([]CollectionSourceFact, error) {
	kind := ""
	switch field {
	case "timeline.host_refs":
		kind = "host"
	case "timeline.identity_refs":
		kind = "identity"
	default:
		return nil, ErrInvalidMentionResolution
	}
	rows, err := tx.Query(ctx, `SELECT entity_mention_id, entity_type, raw_text, resolution_status, row_version, ordinal, resolved_record_id
 FROM entity_mentions WHERE source_record_id=$1 AND source_field_key=$2 AND entity_type=$3
 ORDER BY ordinal, entity_mention_id`, sourceID, field, kind)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	facts := make([]CollectionSourceFact, 0)
	for rows.Next() {
		var fact CollectionSourceFact
		if err := rows.Scan(&fact.MentionID, &fact.EntityType, &fact.RawText, &fact.ResolutionStatus, &fact.RowVersion, &fact.Ordinal, &fact.ResolvedRecordID); err != nil {
			return nil, err
		}
		facts = append(facts, fact)
	}
	return facts, rows.Err()
}
