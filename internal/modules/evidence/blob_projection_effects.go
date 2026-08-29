package evidence

import (
	"context"
	"slices"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	evidenceprojection "github.com/JochiRaider/cartulary/internal/modules/evidence/projectionports"
)

func evidenceRowCellValue(row map[string]any, fieldKey string) any {
	cells, ok := row["cells"].(map[string]any)
	if !ok {
		return nil
	}
	cell, ok := cells[fieldKey].(map[string]any)
	if !ok {
		return nil
	}
	return cell["value"]
}

func int64FromAny(value any) int64 {
	switch typed := value.(type) {
	case int64:
		return typed
	case int:
		return int64(typed)
	case int32:
		return int64(typed)
	case float64:
		return int64(typed)
	default:
		return 0
	}
}

func (s *blobLifecycleService) refreshEvidenceSupportProjectionsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, evidenceRecordID uuid.UUID) ([]attachRecordChange, error) {
	return refreshEvidenceSupportProjectionsTx(ctx, tx, s.supportEffects, incidentID, evidenceRecordID)
}

func refreshEvidenceSupportProjectionsTx(
	ctx context.Context,
	tx pgx.Tx,
	effects evidenceprojection.AssociationEffects,
	incidentID uuid.UUID,
	evidenceRecordID uuid.UUID,
) ([]attachRecordChange, error) {
	subjects, err := loadEvidenceAssociationSubjectsTx(ctx, tx, incidentID, evidenceRecordID)
	if err != nil {
		return nil, err
	}
	if len(subjects) == 0 {
		return nil, nil
	}
	result, err := effects.RefreshEvidenceAssociationEffects(ctx, tx, evidenceprojection.EvidenceAssociationEffectsInput{
		IncidentID: incidentID,
		Subjects:   subjects,
	})
	if err != nil {
		return nil, err
	}
	return attachRecordChangesFromSupportEffects(result), nil
}

func loadEvidenceAssociationSubjectsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, evidenceRecordID uuid.UUID) ([]evidenceprojection.EvidenceAssociationSubject, error) {
	rows, err := tx.Query(ctx, `
SELECT r.record_id, r.record_type
  FROM active_record_links_v1 rl
  JOIN records r
    ON r.incident_id = rl.incident_id
   AND r.record_id = rl.src_record_id
   AND r.deleted_at IS NULL
 WHERE rl.incident_id = $1
   AND rl.dst_record_id = $2
   AND rl.link_type = 'attached_evidence'
 ORDER BY r.record_id, r.record_type
`, incidentID, evidenceRecordID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	subjects := make([]evidenceprojection.EvidenceAssociationSubject, 0)
	for rows.Next() {
		var subject evidenceprojection.EvidenceAssociationSubject
		if err := rows.Scan(&subject.RecordID, &subject.RecordType); err != nil {
			return nil, err
		}
		subjects = append(subjects, subject)
	}
	return subjects, rows.Err()
}

func attachRecordChangesFromSupportEffects(result evidenceprojection.EvidenceAssociationEffectsResult) []attachRecordChange {
	changes := make([]attachRecordChange, 0, len(result.Changes))
	for _, effect := range result.Changes {
		changedFieldKeys := make([]string, 0)
		for _, view := range effect.AffectedViews {
			changedFieldKeys = append(changedFieldKeys, view.ChangedFieldKeys...)
		}
		slices.Sort(changedFieldKeys)
		changedFieldKeys = slices.Compact(changedFieldKeys)
		changes = append(changes, attachRecordChange{
			RecordID:         effect.RecordID,
			RowVersion:       effect.RowVersion,
			ChangedFieldKeys: changedFieldKeys,
			AffectedViews:    append([]evidenceprojection.EvidenceAffectedViewChange(nil), effect.AffectedViews...),
		})
	}
	return changes
}
