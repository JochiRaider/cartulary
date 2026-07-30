package testsupport

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	assessmentprojection "github.com/JochiRaider/cartulary/internal/modules/assessments/projectionprovider"
	"github.com/JochiRaider/cartulary/internal/modules/projections"
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
		rows := projections.NewAssessmentRows(
			database,
			nil,
			assessmentprojection.QuerySurfaces()...,
		)
		if err := rows.ApplyMutationTx(ctx, tx, mutation); err != nil {
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
		if err := projections.ApplyAssessmentMutationSQLTx(
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
