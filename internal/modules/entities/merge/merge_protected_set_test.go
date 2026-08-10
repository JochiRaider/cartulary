package merge

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/assessments"
	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	"github.com/JochiRaider/cartulary/internal/modules/entities/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

type mergeProjectionWriterStub struct {
	workbookprojection.Writer
}

func TestMergeProtectedRecordIDsIncludesAssessmentSubjects(t *testing.T) {
	db := pgtest.Start(t).BeginRollbackDBT(t, "merge-protected-set")
	actor := seedMergeProtectedSetUser(t, db, "merge-protected@example.test", "Merge Protected")
	incident := createMergeProtectedSetIncident(t, db, actor, "txn-merge-protected-incident", "IR-MERGE-PROTECTED", "Merge protected set")
	survivorID := uuid.New()
	loserID := uuid.New()
	assessmentID := uuid.New()
	seedMergeProtectedSetHost(t, db, incident.ID, actor.ID, survivorID, "Survivor host", "survivor-host")
	seedMergeProtectedSetHost(t, db, incident.ID, actor.ID, loserID, "Loser host", "loser-host")
	seedMergeProtectedSetAssessment(t, db, incident.ID, actor.ID, assessmentID, loserID, "host", "confirmed")
	appender := newMergeProtectedSetAppender(t)
	store := NewStore(
		db,
		appender,
		WithAssessmentEffects(assessments.NewMergeEffects(
			mergeAssessmentProjectionStub{},
			appender,
		)),
		WithWorkbookProjection(mergeProjectionWriterStub{}),
	)

	tx, err := db.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin protected-set test tx: %v", err)
	}
	defer func() { _ = tx.Rollback(context.Background()) }()

	recordIDs, err := store.planMergeProtectedRecordIDsTx(context.Background(), tx, survivorID, loserID)
	if err != nil {
		t.Fatalf("plan merge protected set: %v", err)
	}
	recordSet := uuidSet(recordIDs)
	for _, recordID := range []uuid.UUID{survivorID, loserID, assessmentID} {
		if _, ok := recordSet[recordID]; !ok {
			t.Fatalf("expected protected set to include %s, got %#v", recordID, recordIDs)
		}
	}
}

func TestMergeAssessmentRepointRejectsUnprotectedAssessment(t *testing.T) {
	db := pgtest.Start(t).BeginRollbackDBT(t, "merge-protected-set-revalidate")
	appender := newMergeProtectedSetAppender(t)
	store := NewStore(
		db,
		appender,
		WithAssessmentEffects(assessments.NewMergeEffects(
			mergeAssessmentProjectionStub{},
			appender,
		)),
		WithWorkbookProjection(mergeProjectionWriterStub{}),
	)
	actor := seedMergeProtectedSetUser(t, db, "merge-protected-revalidate@example.test", "Merge Protected Revalidate")
	incident := createMergeProtectedSetIncident(t, db, actor, "txn-merge-protected-revalidate-incident", "IR-MERGE-PROTECTED-R", "Merge protected set revalidate")
	survivorID := uuid.New()
	loserID := uuid.New()
	assessmentID := uuid.New()
	seedMergeProtectedSetHost(t, db, incident.ID, actor.ID, survivorID, "Survivor host", "survivor-host")
	seedMergeProtectedSetHost(t, db, incident.ID, actor.ID, loserID, "Loser host", "loser-host")
	seedMergeProtectedSetAssessment(t, db, incident.ID, actor.ID, assessmentID, loserID, "host", "confirmed")

	tx, err := db.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin protected-set revalidation tx: %v", err)
	}
	defer func() { _ = tx.Rollback(context.Background()) }()

	_, _, err = store.ports.assessments.RepointMergedAssessmentsTx(context.Background(), tx, incident.ID, "host", survivorID, loserID, uuidSet([]uuid.UUID{survivorID, loserID}), time.Now().UTC())
	var precondition *MergePreconditionError
	if !errors.As(err, &precondition) || precondition.ReasonCode != "protected_set_changed" {
		t.Fatalf("expected protected_set_changed precondition, got %T %[1]v", err)
	}
}

func newMergeProtectedSetAppender(t testing.TB) *revisions.Appender {
	t.Helper()

	recordViews, err := revisions.NewRecordViewCatalog(
		[]revisions.ProviderContribution{assessments.RevisionProviderContribution()},
		[]revisions.RecordViewSurface{{
			SourceRecordTypes: []string{"assessment"},
			ViewSchemaID:      assessments.AssessmentsViewSchemaID,
		}},
		[]string{assessments.AssessmentsViewSchemaID},
	)
	if err != nil {
		t.Fatalf("build merge protected-set record/view catalog: %v", err)
	}
	snapshotCaptures, err := revisions.NewRecordSnapshotCaptureCatalogForRequirements(
		[]revisions.ProviderContribution{assessments.RevisionProviderContribution()},
		[]revisions.SnapshotSchemaRequirement{{
			RecordType:       "assessment",
			SourceOwner:      revisions.SourceOwnerAssessments,
			SnapshotSchemaID: "cartulary.revisions.snapshot.assessment.v1",
		}},
	)
	if err != nil {
		t.Fatalf("build merge protected-set snapshot capture catalog: %v", err)
	}
	targetSemantics, err := revisions.NewTargetSemanticsCatalog(
		[]revisions.TargetSemanticsRequirement{
			{TargetKind: "assessment", SourceOwnerID: "assessments", DispatchClass: revisions.RollbackDispatchRow, AdmittedRecordTypes: []string{"assessment"}, Addressability: revisions.HistorySingleEntry},
			{TargetKind: "record", SourceOwnerID: "record_source_owner", DispatchClass: revisions.RollbackDispatchRow, AdmittedRecordTypes: []string{"assessment"}, Addressability: revisions.HistorySingleEntry},
		},
		[]revisions.ProviderContribution{assessments.RevisionProviderContribution()},
	)
	if err != nil {
		t.Fatalf("build merge protected-set target-semantics catalog: %v", err)
	}
	appender, err := revisions.NewAppender(
		recordViews,
		mergeProtectedSetEnvelopeReader{},
		snapshotCaptures,
		targetSemantics,
		collaboration.NewHistoricalIntentPolicy(),
		collaboration.NewIntentAppender(),
	)
	if err != nil {
		t.Fatalf("build merge protected-set Revisions appender: %v", err)
	}
	return appender
}

type mergeProtectedSetEnvelopeReader struct{}

func (mergeProtectedSetEnvelopeReader) LoadEnvelopeTx(
	ctx context.Context,
	tx pgx.Tx,
	recordID uuid.UUID,
	lock bool,
) (revisions.RecordEnvelope, error) {
	envelope, err := records.NewStore().LoadEnvelopeTx(ctx, tx, recordID, lock)
	if err != nil {
		return revisions.RecordEnvelope{}, err
	}
	return revisions.RecordEnvelope{
		RecordID:        envelope.RecordID,
		IncidentID:      envelope.IncidentID,
		RecordType:      envelope.RecordType,
		RowVersion:      envelope.RowVersion,
		CreatedByUserID: envelope.CreatedByUserID,
		CreatedAt:       envelope.CreatedAt,
		UpdatedByUserID: envelope.UpdatedByUserID,
		UpdatedAt:       envelope.UpdatedAt,
		DeletedAt:       envelope.DeletedAt,
		DeletedByUserID: envelope.DeletedByUserID,
	}, nil
}

type mergeAssessmentProjectionStub struct{}

func (mergeAssessmentProjectionStub) RefreshAssessmentProjectionTx(
	context.Context,
	pgx.Tx,
	uuid.UUID,
) error {
	return nil
}

func seedMergeProtectedSetUser(t testing.TB, db postgres.DB, email string, displayName string) authn.UserRecord {
	t.Helper()

	hash, err := authn.HashPassword("MergeProtectedPass1!")
	if err != nil {
		t.Fatalf("hash merge protected set password: %v", err)
	}

	var record authn.UserRecord
	if err := db.QueryRow(context.Background(), `
INSERT INTO users (email, display_name, password_hash, mfa_required, is_active, is_deployment_admin)
VALUES ($1, $2, $3, false, true, true)
RETURNING id, email, display_name, password_hash, mfa_required, is_active, is_deployment_admin, created_at, updated_at, user_version
`, email, displayName, hash).Scan(
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
		t.Fatalf("seed merge protected set user: %v", err)
	}
	return record
}

func createMergeProtectedSetIncident(t testing.TB, db postgres.DB, actor authn.UserRecord, clientTxnID string, incidentKey string, title string) incidents.IncidentRecord {
	t.Helper()

	store := incidents.NewApplication(db)
	result, err := store.CreateIncident(context.Background(), actor, incidents.CreateIncidentRequest{
		ClientTxnID: clientTxnID,
		IncidentKey: incidentKey,
		Title:       title,
	}, []byte(clientTxnID), "req-"+clientTxnID, time.Now().UTC())
	if err != nil {
		t.Fatalf("create merge protected set incident: %v", err)
	}
	return result.Incident
}

func seedMergeProtectedSetHost(t testing.TB, db postgres.DB, incidentID uuid.UUID, actorUserID uuid.UUID, recordID uuid.UUID, displayName string, hostname string) {
	t.Helper()

	seedMergeProtectedSetRecord(t, db, incidentID, actorUserID, recordID, "host")
	if _, err := db.Exec(context.Background(), `
INSERT INTO hosts (record_id, incident_id, display_name, hostname, host_state, created_by_user_id, updated_by_user_id)
VALUES ($1, $2, $3, $4, 'canonical', $5, $5)
`, recordID, incidentID, displayName, hostname, actorUserID); err != nil {
		t.Fatalf("seed merge protected set host: %v", err)
	}
}

func seedMergeProtectedSetAssessment(t testing.TB, db postgres.DB, incidentID uuid.UUID, actorUserID uuid.UUID, assessmentID uuid.UUID, subjectID uuid.UUID, subjectType string, state string) {
	t.Helper()

	seedMergeProtectedSetRecord(t, db, incidentID, actorUserID, assessmentID, "assessment")
	if _, err := db.Exec(context.Background(), `
INSERT INTO assessments (record_id, incident_id, subject_record_id, subject_type, assessment_state, rationale, assessor_user_id)
VALUES ($1, $2, $3, $4, $5, 'Seeded merge protected set assessment.', $6)
`, assessmentID, incidentID, subjectID, subjectType, state, actorUserID); err != nil {
		t.Fatalf("seed merge protected set assessment: %v", err)
	}
}

func seedMergeProtectedSetRecord(t testing.TB, db postgres.DB, incidentID uuid.UUID, actorUserID uuid.UUID, recordID uuid.UUID, recordType string) {
	t.Helper()

	if _, err := db.Exec(context.Background(), `
INSERT INTO records (record_id, incident_id, record_type, created_by_user_id, updated_by_user_id)
VALUES ($1, $2, $3, $4, $4)
`, recordID, incidentID, recordType, actorUserID); err != nil {
		t.Fatalf("seed merge protected set record: %v", err)
	}
}
