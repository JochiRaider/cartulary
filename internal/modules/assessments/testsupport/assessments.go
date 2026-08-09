package testsupport

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	assessmentprojection "github.com/JochiRaider/cartulary/internal/modules/assessments/workbookprojection"
	projectiontestsupport "github.com/JochiRaider/cartulary/internal/modules/projections/testsupport"
	"github.com/JochiRaider/cartulary/internal/modules/records/testsupport/envelopetest"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

var (
	HostAssessmentID     = uuid.MustParse("40000000-0000-0000-0000-000000000901")
	IdentityAssessmentID = uuid.MustParse("40000000-0000-0000-0000-000000000902")
)

func SeedAssessment(
	t testing.TB,
	db any,
	incidentID uuid.UUID,
	actorUserID uuid.UUID,
	assessmentID uuid.UUID,
	subjectID uuid.UUID,
	subjectType string,
	state string,
) {
	t.Helper()
	envelopetest.SeedRecordEnvelope(t, db, incidentID, actorUserID, assessmentID, "assessment")
	assessedAt := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	if _, err := execDB(db, `
INSERT INTO assessments (
    record_id,
    incident_id,
    subject_record_id,
    subject_type,
    assessment_state,
    rationale,
    assessor_user_id,
    assessed_at
)
VALUES ($1, $2, $3, $4, $5, 'Seeded test assessment rationale.', $6, $7)
`, assessmentID, incidentID, subjectID, subjectType, state, actorUserID, assessedAt); err != nil {
		t.Fatalf("seed assessment: %v", err)
	}
	ctx := context.Background()
	mutation := assessmentprojection.ProjectionMutation{
		Kind:     assessmentprojection.ProjectionMutationUpsert,
		RecordID: assessmentID,
		Input: assessmentprojection.ProjectionInput{
			RecordID:            assessmentID,
			IncidentID:          incidentID,
			RowVersion:          1,
			SubjectRef:          subjectID,
			SubjectType:         subjectType,
			AssessmentState:     state,
			ConfidenceBand:      "unset",
			Rationale:           "Seeded test assessment rationale.",
			Assessor:            actorUserID,
			AssessedAt:          assessedAt,
			SupportingLinkCount: 0,
		},
	}
	switch database := db.(type) {
	case postgres.DB:
		tx, err := database.BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			t.Fatalf("begin seeded assessment projection: %v", err)
		}
		defer func() { _ = tx.Rollback(ctx) }()
		if err := projectiontestsupport.ApplyAssessmentFixtureMutationTx(ctx, database, tx, mutation); err != nil {
			t.Fatalf("apply seeded assessment projection: %v", err)
		}
		if err := tx.Commit(ctx); err != nil {
			t.Fatalf("commit seeded assessment projection: %v", err)
		}
	case *sql.DB:
		tx, err := database.BeginTx(ctx, nil)
		if err != nil {
			t.Fatalf("begin seeded assessment SQL projection: %v", err)
		}
		defer func() { _ = tx.Rollback() }()
		if err := applyAssessmentMutationSQLTx(
			ctx,
			tx,
			mutation,
		); err != nil {
			t.Fatalf("apply seeded assessment SQL projection: %v", err)
		}
		if err := tx.Commit(); err != nil {
			t.Fatalf("commit seeded assessment SQL projection: %v", err)
		}
	default:
		t.Fatalf("seed assessment projection participant: unsupported database %T", db)
	}
}

func applyAssessmentMutationSQLTx(
	ctx context.Context,
	tx *sql.Tx,
	mutation assessmentprojection.ProjectionMutation,
) error {
	if err := mutation.Validate(); err != nil {
		return err
	}
	switch mutation.Kind {
	case assessmentprojection.ProjectionMutationUpsert:
		input := mutation.Input
		if _, err := tx.ExecContext(ctx, assessmentFixtureUpsertSQL,
			input.RecordID, input.IncidentID, input.RowVersion, input.SubjectRef,
			input.SubjectType, input.AssessmentState, input.ConfidenceScore,
			input.ConfidenceBand, input.Rationale, input.Assessor,
			input.AssessedAt.UTC(), input.SupportingLinkCount,
		); err != nil {
			return fmt.Errorf("upsert assessment projection fixture row: %w", err)
		}
		return nil
	case assessmentprojection.ProjectionMutationDelete:
		if _, err := tx.ExecContext(
			ctx,
			`DELETE FROM assessment_grid_projection WHERE record_id = $1`,
			mutation.RecordID,
		); err != nil {
			return fmt.Errorf("delete assessment projection fixture row: %w", err)
		}
		return nil
	default:
		return fmt.Errorf("unsupported assessment projection fixture mutation kind %q", mutation.Kind)
	}
}

const assessmentFixtureUpsertSQL = `
INSERT INTO assessment_grid_projection (
    record_id, incident_id, row_version, subject_ref, subject_type,
    assessment_state, confidence_score, confidence_band, rationale,
    assessor, assessed_at, supporting_link_count
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
ON CONFLICT (record_id) DO UPDATE
SET incident_id = EXCLUDED.incident_id,
    row_version = EXCLUDED.row_version,
    subject_ref = EXCLUDED.subject_ref,
    subject_type = EXCLUDED.subject_type,
    assessment_state = EXCLUDED.assessment_state,
    confidence_score = EXCLUDED.confidence_score,
    confidence_band = EXCLUDED.confidence_band,
    rationale = EXCLUDED.rationale,
    assessor = EXCLUDED.assessor,
    assessed_at = EXCLUDED.assessed_at,
    supporting_link_count = EXCLUDED.supporting_link_count
`

func LookupAssessmentSubject(t testing.TB, db any, assessmentID uuid.UUID) uuid.UUID {
	t.Helper()
	var subjectID string
	if err := queryRowDB(db, `
SELECT subject_record_id::text FROM assessments WHERE record_id = $1
`, assessmentID).Scan(&subjectID); err != nil {
		t.Fatalf("lookup assessment subject: %v", err)
	}
	return uuid.MustParse(subjectID)
}

type rowScanner interface {
	Scan(dest ...any) error
}

func queryRowDB(db any, query string, args ...any) rowScanner {
	switch typed := db.(type) {
	case postgres.DB:
		return typed.QueryRow(context.Background(), query, args...)
	case *sql.DB:
		return typed.QueryRowContext(context.Background(), query, args...)
	default:
		panic(fmt.Sprintf("unsupported Assessments test database %T", db))
	}
}

func execDB(db any, query string, args ...any) (any, error) {
	switch typed := db.(type) {
	case postgres.DB:
		return typed.Exec(context.Background(), query, args...)
	case *sql.DB:
		return typed.ExecContext(context.Background(), query, args...)
	default:
		return nil, fmt.Errorf("unsupported Assessments test database %T", db)
	}
}
