package merge_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/assessments"
	entitymerge "github.com/JochiRaider/cartulary/internal/modules/entities/merge"
	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/mentioneffects"
	workbookstartuppostgres "github.com/JochiRaider/cartulary/internal/modules/workbook/startup/postgres"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
	"github.com/JochiRaider/cartulary/internal/testutil/revisionsupport"
)

func TestMergeProtectedRecordIDsIncludesAssessmentSubjects(t *testing.T) {
	db := pgtest.Start(t).BeginRollbackDBT(t, "merge-protected-set-composition")
	actor := seedComposedMergeUser(t, db)
	incident := createComposedMergeIncident(t, db, actor)
	survivorID := uuid.New()
	loserID := uuid.New()
	assessmentID := uuid.New()
	seedComposedMergeHost(t, db, incident.ID, actor.ID, survivorID, "Survivor host", "survivor-host")
	seedComposedMergeHost(t, db, incident.ID, actor.ID, loserID, "Loser host", "loser-host")
	seedComposedMergeAssessment(t, db, incident.ID, actor.ID, assessmentID, loserID)

	composition := revisionsupport.MustComposition(t)
	appender := composition.Runtime.Appender()
	assessmentEffects, err := assessments.NewMergeEffects(composedMergeAssessmentProjection{}, appender)
	if err != nil {
		t.Fatalf("construct assessment merge effects: %v", err)
	}
	store := entitymerge.NewStore(
		db,
		appender,
		entitymerge.WithAssessmentEffects(assessmentEffects),
		entitymerge.WithLinkEffects(composedMergeLinkEffects{}),
		entitymerge.WithWorkbookProjection(composedMergeWorkbookProjection{}),
		entitymerge.WithTimelineEffects(composedMergeTimelineEffects{}),
		entitymerge.WithCollaborationIntents(composition.Intents),
	)
	result, err := store.MergeEntity(
		context.Background(),
		actor,
		survivorID,
		entitymerge.MergeRequest{
			LoserRecordID:          loserID,
			SurvivorBaseRowVersion: 1,
			LoserBaseRowVersion:    1,
			ClientTxnID:            "txn-merge-protected-set-composition",
		},
		[]byte("merge-protected-set-composition"),
		"req-merge-protected-set-composition",
		time.Now().UTC(),
	)
	if err != nil {
		t.Fatalf("merge with protected assessment: %v", err)
	}
	if result.MergeSummary.RepointedAssessmentCount != 1 {
		t.Fatalf("repointed assessment count = %d, want 1", result.MergeSummary.RepointedAssessmentCount)
	}
	var subjectRecordID uuid.UUID
	if err := db.QueryRow(context.Background(), `SELECT subject_record_id FROM assessments WHERE record_id = $1`, assessmentID).Scan(&subjectRecordID); err != nil {
		t.Fatalf("load repointed assessment: %v", err)
	}
	if subjectRecordID != survivorID {
		t.Fatalf("assessment subject = %s, want survivor %s", subjectRecordID, survivorID)
	}
}

type composedMergeLinkEffects struct{}

func (composedMergeLinkEffects) RepointLinksTx(
	context.Context,
	pgx.Tx,
	entitymerge.RepointLinksCommand,
) (entitymerge.RepointLinksResult, error) {
	return entitymerge.RepointLinksResult{
		Mutations:                 []entitymerge.LinkEffectMutation{},
		LinkTypesBySourceRecordID: map[uuid.UUID][]string{},
	}, nil
}

func (composedMergeLinkEffects) RepointTagsTx(
	context.Context,
	pgx.Tx,
	entitymerge.RepointTagsCommand,
) (entitymerge.RepointTagsResult, error) {
	return entitymerge.RepointTagsResult{Mutations: []entitymerge.LinkEffectMutation{}}, nil
}

type composedMergeAssessmentProjection struct{}

func (composedMergeAssessmentProjection) RefreshAssessmentProjectionTx(context.Context, pgx.Tx, uuid.UUID) error {
	return nil
}

type composedMergeWorkbookProjection struct{}

func (composedMergeWorkbookProjection) RefreshHostTx(context.Context, pgx.Tx, uuid.UUID) error {
	return nil
}

func (composedMergeWorkbookProjection) RefreshIdentityTx(context.Context, pgx.Tx, uuid.UUID) error {
	return nil
}

func (composedMergeWorkbookProjection) DeleteHostTx(context.Context, pgx.Tx, uuid.UUID) error {
	return nil
}

func (composedMergeWorkbookProjection) DeleteIdentityTx(context.Context, pgx.Tx, uuid.UUID) error {
	return nil
}

func (composedMergeWorkbookProjection) RebuildHostsTx(context.Context, pgx.Tx, uuid.UUID) error {
	return nil
}

func (composedMergeWorkbookProjection) RebuildIdentitiesTx(context.Context, pgx.Tx, uuid.UUID) error {
	return nil
}

type composedMergeTimelineEffects struct{}

func (composedMergeTimelineEffects) LoadTimelineInvalidationsTx(context.Context, pgx.Tx, map[uuid.UUID][]string) ([]mentioneffects.TimelineInvalidation, error) {
	return nil, nil
}

func (composedMergeTimelineEffects) LoadRelationshipInvalidationsTx(context.Context, pgx.Tx, map[uuid.UUID][]string) ([]mentioneffects.TimelineInvalidation, error) {
	return nil, nil
}

func (composedMergeTimelineEffects) RefreshTimelineProjectionRowsTx(context.Context, pgx.Tx, []uuid.UUID) error {
	return nil
}

func seedComposedMergeUser(t testing.TB, db postgres.DB) authn.UserRecord {
	t.Helper()
	hash, err := authn.HashPassword("MergeProtectedPass1!")
	if err != nil {
		t.Fatalf("hash merge protected-set password: %v", err)
	}
	var record authn.UserRecord
	if err := db.QueryRow(context.Background(), `
INSERT INTO users (email, display_name, password_hash, mfa_required, is_active, is_deployment_admin)
VALUES ('merge-protected-composition@example.test', 'Merge Protected', $1, false, true, true)
RETURNING id, email, display_name, password_hash, mfa_required, is_active, is_deployment_admin, created_at, updated_at, user_version
`, hash).Scan(
		&record.ID,
		&record.Email,
		&record.DisplayName,
		&record.PasswordHash,
		&record.MFARequired,
		&record.IsActive,
		&record.IsDeploymentAdmin,
		&record.CreatedAt,
		&record.UpdatedAt,
		&record.UserVersion,
	); err != nil {
		t.Fatalf("seed merge protected-set user: %v", err)
	}
	return record
}

func createComposedMergeIncident(t testing.TB, db postgres.DB, actor authn.UserRecord) incidents.IncidentRecord {
	t.Helper()
	application, err := incidents.NewApplication(incidents.ApplicationDependencies{
		Postgres:            db,
		PreferenceBootstrap: workbookstartuppostgres.NewWriter(),
	})
	if err != nil {
		t.Fatalf("construct Incidents application: %v", err)
	}
	result, err := application.CreateIncident(
		context.Background(),
		actor,
		incidents.CreateIncidentRequest{
			ClientTxnID: "txn-merge-protected-incident-composition",
			IncidentKey: "IR-MERGE-PROTECTED-COMPOSITION",
			Title:       "Merge protected set composition",
		},
		"req-merge-protected-incident-composition",
		time.Now().UTC(),
	)
	if err != nil {
		t.Fatalf("create merge protected-set incident: %v", err)
	}
	return result.Incident
}

func seedComposedMergeHost(t testing.TB, db postgres.DB, incidentID uuid.UUID, actorID uuid.UUID, recordID uuid.UUID, displayName string, hostname string) {
	t.Helper()
	seedComposedMergeRecord(t, db, incidentID, actorID, recordID, "host")
	if _, err := db.Exec(context.Background(), `
INSERT INTO hosts (record_id, incident_id, display_name, hostname, host_state, created_by_user_id, updated_by_user_id)
VALUES ($1, $2, $3, $4, 'canonical', $5, $5)
`, recordID, incidentID, displayName, hostname, actorID); err != nil {
		t.Fatalf("seed merge protected-set host: %v", err)
	}
}

func seedComposedMergeAssessment(t testing.TB, db postgres.DB, incidentID uuid.UUID, actorID uuid.UUID, assessmentID uuid.UUID, subjectID uuid.UUID) {
	t.Helper()
	seedComposedMergeRecord(t, db, incidentID, actorID, assessmentID, "assessment")
	if _, err := db.Exec(context.Background(), `
INSERT INTO assessments (record_id, incident_id, subject_record_id, subject_type, assessment_state, rationale, assessor_user_id)
VALUES ($1, $2, $3, 'host', 'confirmed', 'Seeded merge protected set assessment.', $4)
`, assessmentID, incidentID, subjectID, actorID); err != nil {
		t.Fatalf("seed merge protected-set assessment: %v", err)
	}
}

func seedComposedMergeRecord(t testing.TB, db postgres.DB, incidentID uuid.UUID, actorID uuid.UUID, recordID uuid.UUID, recordType string) {
	t.Helper()
	if _, err := db.Exec(context.Background(), `
INSERT INTO records (record_id, incident_id, record_type, created_by_user_id, updated_by_user_id)
VALUES ($1, $2, $3, $4, $4)
`, recordID, incidentID, recordType, actorID); err != nil {
		t.Fatalf("seed merge protected-set record: %v", err)
	}
}
