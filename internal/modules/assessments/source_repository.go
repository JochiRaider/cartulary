package assessments

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type assessmentSourceRepository struct{}

type assessmentSourceCreate struct {
	RecordID        uuid.UUID
	IncidentID      uuid.UUID
	SubjectRef      uuid.UUID
	SubjectType     string
	AssessmentState string
	ConfidenceScore *int
	Rationale       string
	Assessor        uuid.UUID
	AssessedAt      time.Time
	Now             time.Time
}

func (assessmentSourceRepository) InsertTx(ctx context.Context, tx pgx.Tx, create assessmentSourceCreate) error {
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
    created_at,
    updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $10)
`, create.RecordID, create.IncidentID, create.SubjectRef, create.SubjectType, create.AssessmentState, create.ConfidenceScore, create.Rationale, create.Assessor, create.AssessedAt.UTC(), create.Now.UTC())
	if err != nil {
		return fmt.Errorf("insert assessment: %w", err)
	}
	return nil
}
