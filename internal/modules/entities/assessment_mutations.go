package entities

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type AssessmentFieldValue struct {
	Text      *string
	Timestamp *time.Time
	UUID      *uuid.UUID
	Number    *int64
	Bool      *bool
}

type AssessmentCreateParams struct {
	Values map[string]AssessmentFieldValue
}

func (s *Store) InsertAssessmentTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, incidentID uuid.UUID, actorID uuid.UUID, params AssessmentCreateParams, now time.Time) error {
	assessor := nullableAssessmentUUIDValue(params.Values, "assessment.assessor")
	if assessor == nil {
		assessor = actorID
	}
	assessedAt := nullableAssessmentTimestampValue(params.Values, "assessment.assessed_at")
	if assessedAt == nil {
		assessedAt = now.UTC()
	}
	_, err := tx.Exec(ctx, `
INSERT INTO assessments (
    record_id,
    incident_id,
    subject_record_id,
    subject_type,
    assessment_state,
    confidence_score,
    rationale,
    assessor_user_id,
    assessed_at,
    deleted_at
) VALUES (
    $1, $2, $3, $4, $5,
    $6, $7, $8, $9, NULL
)
`, recordID, incidentID,
		assessmentUUIDValue(params.Values, "assessment.subject_ref"),
		assessmentTextValue(params.Values, "assessment.subject_type"),
		assessmentTextValue(params.Values, "assessment.assessment_state"),
		nullableAssessmentNumberValue(params.Values, "assessment.confidence_score"),
		assessmentTextValue(params.Values, "assessment.rationale"),
		assessor,
		assessedAt)
	if err != nil {
		return fmt.Errorf("insert assessment: %w", err)
	}
	return nil
}

func assessmentTextValue(values map[string]AssessmentFieldValue, field string) string {
	if value, ok := values[field]; ok && value.Text != nil {
		return *value.Text
	}
	return ""
}

func assessmentUUIDValue(values map[string]AssessmentFieldValue, field string) uuid.UUID {
	if value, ok := values[field]; ok && value.UUID != nil {
		return *value.UUID
	}
	return uuid.Nil
}

func nullableAssessmentUUIDValue(values map[string]AssessmentFieldValue, field string) any {
	if value, ok := values[field]; ok && value.UUID != nil {
		return *value.UUID
	}
	return nil
}

func nullableAssessmentTimestampValue(values map[string]AssessmentFieldValue, field string) any {
	if value, ok := values[field]; ok && value.Timestamp != nil {
		return value.Timestamp.UTC()
	}
	return nil
}

func nullableAssessmentNumberValue(values map[string]AssessmentFieldValue, field string) any {
	if value, ok := values[field]; ok && value.Number != nil {
		return *value.Number
	}
	return nil
}
