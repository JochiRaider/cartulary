package testsupport

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/google/uuid"

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

	if _, err := execDB(db, `
INSERT INTO assessments (record_id, incident_id, subject_record_id, subject_type, assessment_state, rationale, assessor_user_id)
VALUES ($1, $2, $3, $4, $5, 'Seeded test assessment rationale.', $6)
`, assessmentID, incidentID, subjectID, subjectType, state, actorUserID); err != nil {
		t.Fatalf("seed assessment: %v", err)
	}
	if _, err := execDB(db, `
INSERT INTO assessment_grid_projection (
    record_id, incident_id, row_version, subject_ref, subject_type,
    assessment_state, confidence_band, rationale, assessor, assessed_at, supporting_link_count
)
SELECT a.record_id, a.incident_id, r.row_version, a.subject_record_id, a.subject_type,
       a.assessment_state, 'unset', a.rationale, a.assessor_user_id, a.assessed_at, 0
  FROM assessments a
  JOIN records r ON r.record_id = a.record_id
 WHERE a.record_id = $1
ON CONFLICT (record_id) DO NOTHING
`, assessmentID); err != nil {
		t.Fatalf("seed assessment projection: %v", err)
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
