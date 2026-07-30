package links

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type AssessmentSupportFacts struct {
	TargetRecordIDs []uuid.UUID
}

type AssessmentFactReader struct{}

func (AssessmentFactReader) LoadSupportFactsTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	assessmentID uuid.UUID,
) (AssessmentSupportFacts, error) {
	rows, err := tx.Query(ctx, `
SELECT dst_record_id
  FROM active_record_links_v1
 WHERE incident_id = $1
   AND src_record_id = $2
   AND link_type = 'supported_by'
 ORDER BY dst_record_id
`, incidentID, assessmentID)
	if err != nil {
		return AssessmentSupportFacts{}, fmt.Errorf("load assessment support link facts: %w", err)
	}
	defer rows.Close()
	facts := AssessmentSupportFacts{TargetRecordIDs: []uuid.UUID{}}
	for rows.Next() {
		var targetID uuid.UUID
		if err := rows.Scan(&targetID); err != nil {
			return AssessmentSupportFacts{}, fmt.Errorf("scan assessment support link fact: %w", err)
		}
		facts.TargetRecordIDs = append(facts.TargetRecordIDs, targetID)
	}
	if err := rows.Err(); err != nil {
		return AssessmentSupportFacts{}, fmt.Errorf("iterate assessment support link facts: %w", err)
	}
	return facts, nil
}
