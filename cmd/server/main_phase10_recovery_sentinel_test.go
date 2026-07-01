package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/app"
	"github.com/JochiRaider/cartulary/internal/modules/projections"
	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	"github.com/JochiRaider/cartulary/internal/modules/recovery/restorecontract"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/testutil/configtest"
	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/processtest"
)

func TestPhase10_U_10_02_RestoreReadinessAndCoherentStoreOrder(t *testing.T) {
	ctx := context.Background()
	fixture := capturePhase10RestoreSource(t, "phase10-u-10-02-source")

	failing := preparePhase10RestoreTarget(t, "phase10-u-10-02-failing-target")
	failingGate := &phase10RestoreReadinessGate{}
	failingRecorder := &phase10RestoreStepRecorder{}
	_, err := recovery.NewRestoreRunner(fixture.SourceStore, fixture.BackupStorage).RestoreLatestSuccessfulRetained(ctx, recovery.RestoreTarget{
		Postgres:    failing.Postgres,
		ObjectStore: failing.ObjectStore,
		Projections: phase10FailingProjectionRebuilder{},
		Readiness:   failingGate,
		Observer:    failingRecorder,
	}, fixture.AsOf)
	if !errors.Is(err, errPhase10ProjectionBlocked) {
		t.Fatalf("projection failure got %v want %v", err, errPhase10ProjectionBlocked)
	}
	if failingGate.Marked {
		t.Fatal("restore readiness was marked even though projection rebuild failed")
	}
	if got, want := failingRecorder.Steps, []recovery.RestoreStep{
		recovery.RestoreStepPostgresRestore,
		recovery.RestoreStepObjectStoreRestore,
		recovery.RestoreStepProjectionRebuild,
	}; !reflect.DeepEqual(got, want) {
		t.Fatalf("failing restore steps got %v want %v", got, want)
	}

	target := preparePhase10RestoreTarget(t, "phase10-u-10-02-target")
	gate := &phase10RestoreReadinessGate{}
	recorder := &phase10RestoreStepRecorder{}
	result, err := recovery.NewRestoreRunner(fixture.SourceStore, fixture.BackupStorage).RestoreLatestSuccessfulRetained(ctx, recovery.RestoreTarget{
		Postgres:    target.Postgres,
		ObjectStore: target.ObjectStore,
		Projections: phase10GuardedProjectionRebuilder{
			Inner: projections.NewRestoreRebuilder(target.Postgres),
			Gate:  gate,
		},
		Readiness: gate,
		Observer:  recorder,
	}, fixture.AsOf)
	if err != nil {
		t.Fatalf("restore latest retained backup: %v", err)
	}
	if result.BackupSet.BackupSetID != fixture.LatestBackupSetID {
		t.Fatalf("restore selected backup_set %s want %s", result.BackupSet.BackupSetID, fixture.LatestBackupSetID)
	}
	if !result.BackupSet.ConsistencyPointAt.Equal(fixture.ConsistencyPointAt) {
		t.Fatalf("restore selected consistency point %s want %s", result.BackupSet.ConsistencyPointAt, fixture.ConsistencyPointAt)
	}
	if result.BackupSet.PostgresRestoreAnchor == "" ||
		result.BackupSet.ObjectStoreRestoreAnchor == "" ||
		result.BackupSet.PostgresRestoreAnchor == result.BackupSet.ObjectStoreRestoreAnchor {
		t.Fatalf("restore did not use distinct declared store anchors from one backup set: %#v", result.BackupSet)
	}
	if !gate.Marked {
		t.Fatal("restore readiness was not marked after successful projection rebuild and consistency checks")
	}
	if got, want := recorder.Steps, []recovery.RestoreStep{
		recovery.RestoreStepPostgresRestore,
		recovery.RestoreStepObjectStoreRestore,
		recovery.RestoreStepProjectionRebuild,
		recovery.RestoreStepConsistencyCheck,
		recovery.RestoreStepReadiness,
	}; !reflect.DeepEqual(got, want) {
		t.Fatalf("restore steps got %v want %v", got, want)
	}
	if result.ConsistencyReport.AuthoritativeRowCount == 0 ||
		result.ConsistencyReport.ChangeSetRowCount == 0 ||
		result.ConsistencyReport.BlobCount != 1 {
		t.Fatalf("restore consistency report did not cover rows, change sets, and blob hashes: %#v", result.ConsistencyReport)
	}
}

func TestPhase10_U_10_03_FailClosedRestoreVerificationBlocked(t *testing.T) {
	ctx := context.Background()
	fixture := capturePhase10RestoreSource(t, "phase10-u-10-03-source")
	backupSet, err := fixture.SourceStore.GetBackupSet(ctx, fixture.LatestBackupSetID)
	if err != nil {
		t.Fatalf("load selected backup set: %v", err)
	}
	integrityBody := phase10RequireStoredArtifactProof(t, fixture.BackupStorage, recovery.BackupArtifactProof{
		Key:       backupSet.IntegrityManifestKey,
		SHA256:    backupSet.IntegrityManifestSHA256,
		SizeBytes: backupSet.IntegrityManifestSizeBytes,
	})
	integrityManifest, err := recovery.DecodeIntegrityManifest(integrityBody)
	if err != nil {
		t.Fatalf("decode selected integrity manifest: %v", err)
	}
	if integrityManifest.ObjectStoreBackupManifestArtifact == nil {
		t.Fatal("selected backup is missing Phase E object-store backup manifest proof")
	}
	for index, tc := range []struct {
		name    string
		storage recovery.BackupStorage
	}{
		{
			name: "missing integrity manifest",
			storage: phase10TamperedBackupStorage{
				Inner:   fixture.BackupStorage,
				Missing: map[string]bool{backupSet.IntegrityManifestKey: true},
			},
		},
		{
			name: "corrupt manifest hash",
			storage: phase10TamperedBackupStorage{
				Inner:        fixture.BackupStorage,
				Replacements: map[string][]byte{backupSet.IntegrityManifestKey: []byte(`{"schema_id":"cartulary.backup_integrity_manifest.v1"}`)},
			},
		},
		{
			name: "missing postgres artifact",
			storage: phase10TamperedBackupStorage{
				Inner:   fixture.BackupStorage,
				Missing: map[string]bool{backupSet.PostgresArtifactKey: true},
			},
		},
		{
			name: "missing object artifact",
			storage: phase10TamperedBackupStorage{
				Inner:   fixture.BackupStorage,
				Missing: map[string]bool{backupSet.ObjectStoreArtifactKey: true},
			},
		},
		{
			name: "missing object manifest",
			storage: phase10TamperedBackupStorage{
				Inner:   fixture.BackupStorage,
				Missing: map[string]bool{integrityManifest.ObjectStoreBackupManifestArtifact.Key: true},
			},
		},
		{
			name: "corrupt postgres artifact hash",
			storage: phase10TamperedBackupStorage{
				Inner:        fixture.BackupStorage,
				Replacements: map[string][]byte{backupSet.PostgresArtifactKey: []byte(`{"schema_id":"cartulary.postgres_snapshot_artifact.v1","tables":[]}`)},
			},
		},
		{
			name: "corrupt object artifact hash",
			storage: phase10TamperedBackupStorage{
				Inner:        fixture.BackupStorage,
				Replacements: map[string][]byte{backupSet.ObjectStoreArtifactKey: []byte(`{"schema_id":"cartulary.object_store_snapshot_artifact.v2","objects":[]}`)},
			},
		},
		{
			name: "corrupt object manifest hash",
			storage: phase10TamperedBackupStorage{
				Inner:        fixture.BackupStorage,
				Replacements: map[string][]byte{integrityManifest.ObjectStoreBackupManifestArtifact.Key: []byte(`{"schema_id":"cartulary.object_store_backup_manifest.v1"}`)},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			target := preparePhase10RestoreTarget(t, fmt.Sprintf("phase10-u-10-03-target-%02d", index))
			gate := &phase10RestoreReadinessGate{}
			recorder := &phase10RestoreStepRecorder{}
			_, err := recovery.NewRestoreRunner(fixture.SourceStore, tc.storage).RestoreBackupSet(ctx, recovery.RestoreTarget{
				Postgres:    target.Postgres,
				ObjectStore: target.ObjectStore,
				Projections: projections.NewRestoreRebuilder(target.Postgres),
				Readiness:   gate,
				Observer:    recorder,
			}, backupSet)
			if err == nil {
				t.Fatalf("tampered backup %q unexpectedly restored", tc.name)
			}
			if gate.Marked {
				t.Fatalf("tampered backup %q marked readiness", tc.name)
			}
			for _, step := range recorder.Steps {
				if step == recovery.RestoreStepReadiness {
					t.Fatalf("tampered backup %q reached readiness step: %v", tc.name, recorder.Steps)
				}
			}
		})
	}

	basis, err := recovery.RestoreVerificationBasisSHA256(map[string]string{
		"backup_mechanism": "phase10.process.restore.v1",
		"source_root":      "phase10-u-10-03-source",
	})
	if err != nil {
		t.Fatalf("restore verification basis: %v", err)
	}
	successTarget := preparePhase10RestoreTarget(t, "phase10-u-10-03-verify-success")
	successResult, err := recovery.NewRestoreVerificationService(fixture.SourceStore, recovery.NewRestoreRunner(fixture.SourceStore, fixture.BackupStorage)).VerifyLatestSuccessfulRetained(ctx, recovery.RestoreVerificationTarget{
		RestoreTarget: recovery.RestoreTarget{
			Postgres:    successTarget.Postgres,
			ObjectStore: successTarget.ObjectStore,
			Projections: projections.NewRestoreRebuilder(successTarget.Postgres),
		},
		Probe: app.RestoreVerificationWorkbookProbe{Postgres: successTarget.Postgres},
	}, fixture.AsOf, basis)
	if err != nil {
		t.Fatalf("restore verification success: %v", err)
	}
	if successResult.BackupSet.VerificationState != recovery.VerificationVerified ||
		successResult.BackupSet.LastVerifiedRestoreAt == nil ||
		successResult.BackupSet.LastVerificationBasisSHA256 != basis ||
		successResult.Run.VerificationState != recovery.VerificationVerified {
		t.Fatalf("restore verification success did not record verified state and run: %#v", successResult)
	}
	restoreArtifactBody := phase10RequireStoredArtifactProof(t, fixture.BackupStorage, successResult.ArtifactProof)
	restoreArtifactPath := phase10WriteEvidenceArtifact(t, "phase-e-backup-restore", "restore-verification.json", restoreArtifactBody)
	restoreArtifact, err := recovery.DecodeRestoreVerificationArtifact(restoreArtifactBody)
	if err != nil {
		t.Fatalf("decode retained restore verification artifact: %v", err)
	}
	if restoreArtifact.SchemaID != recovery.RestoreVerificationArtifactSchemaID ||
		restoreArtifact.BackupSetID != fixture.LatestBackupSetID.String() ||
		restoreArtifact.Result != "pass" ||
		restoreArtifact.ManifestCheckResult != "pass" ||
		restoreArtifact.ProjectionRebuildResult != "pass" ||
		restoreArtifact.IncidentOpenCheck.Status != "pass" ||
		restoreArtifact.QueryViewSchemaID != recovery.RestoreVerificationTimelineViewID ||
		restoreArtifact.BlobCheckCounts.Total != 1 ||
		restoreArtifact.BlobCheckCounts.Failed != 0 ||
		restoreArtifact.SelectedIncidentID == nil ||
		*restoreArtifact.SelectedIncidentID != fixture.IncidentID {
		t.Fatalf("restore verification artifact does not satisfy Phase E predicate at %s: %#v", restoreArtifactPath, restoreArtifact)
	}
	t.Logf("phase_e_restore_verification_artifact=%s", restoreArtifactPath)

	failureTarget := preparePhase10RestoreTarget(t, "phase10-u-10-03-verify-failure")
	_, err = recovery.NewRestoreVerificationService(fixture.SourceStore, recovery.NewRestoreRunner(fixture.SourceStore, fixture.BackupStorage)).VerifyLatestSuccessfulRetained(ctx, recovery.RestoreVerificationTarget{
		RestoreTarget: recovery.RestoreTarget{
			Postgres:    failureTarget.Postgres,
			ObjectStore: failureTarget.ObjectStore,
			Projections: phase10FailingProjectionRebuilder{},
		},
	}, fixture.AsOf, basis)
	if !errors.Is(err, errPhase10ProjectionBlocked) {
		t.Fatalf("restore verification failure got %v want %v", err, errPhase10ProjectionBlocked)
	}
	failed, err := fixture.SourceStore.GetBackupSet(ctx, fixture.LatestBackupSetID)
	if err != nil {
		t.Fatalf("reload failed restore verification state: %v", err)
	}
	if failed.VerificationState != recovery.VerificationFailed ||
		failed.LastVerifiedRestoreAt == nil ||
		failed.LastVerificationBasisSHA256 != basis {
		t.Fatalf("restore verification failure did not record failed state: %#v", failed)
	}
}

func TestPhase10_I_10_02_FreshEnvironmentRestoreWorkbookConsistency(t *testing.T) {
	ctx := context.Background()
	fixture := capturePhase10RestoreSource(t, "phase10-i-10-02-source")
	target := preparePhase10RestoreTarget(t, "phase10-i-10-02-target")
	gate := &phase10RestoreReadinessGate{}

	result, err := recovery.NewRestoreRunner(fixture.SourceStore, fixture.BackupStorage).RestoreLatestSuccessfulRetained(ctx, recovery.RestoreTarget{
		Postgres:    target.Postgres,
		ObjectStore: target.ObjectStore,
		Projections: projections.NewRestoreRebuilder(target.Postgres),
		Readiness:   gate,
	}, fixture.AsOf)
	if err != nil {
		t.Fatalf("restore latest retained backup into fresh environment: %v", err)
	}
	if !gate.Marked {
		t.Fatal("restore did not mark readiness after fresh environment restore")
	}
	if result.BackupSet.BackupSetID != fixture.LatestBackupSetID {
		t.Fatalf("fresh restore selected backup_set %s want %s", result.BackupSet.BackupSetID, fixture.LatestBackupSetID)
	}
	if result.ConsistencyReport.AuthoritativeRowsSHA256 == "" ||
		result.ConsistencyReport.ChangeSetsSHA256 == "" ||
		result.ConsistencyReport.BlobHashesSHA256 == "" {
		t.Fatalf("fresh restore did not publish complete consistency hashes: %#v", result.ConsistencyReport)
	}

	server := processtest.StartServer(t, processtest.ServerOptions{Env: target.Env})
	defer server.Stop(t)
	server.WaitForReady(t)

	login := phase1LoginLocalUserWithSecondFactor(t, server, phase1BootstrapAdminEmail, phase1BootstrapAdminPassword, phase1GenerateTOTPCode(t, fixture.AdminTOTPSecret))
	getIncident := phase1DoJSON(t, server, http.MethodGet, "/api/v1/incidents/"+fixture.IncidentID, nil, withCookies(login.sessionCookie))
	httptestx.RequireSuccessEnvelope(t, getIncident, http.StatusOK)
	phase5RequireTimelineEvidenceCount(t, server, login, fixture.IncidentID, fixture.TimelineRecordID, 1, true)
}

func TestPhase10_I_10_03_MissingArtifactFailsBeforeReadinessBlocked(t *testing.T) {
	ctx := context.Background()
	fixture := capturePhase10RestoreSource(t, "phase10-i-10-03-source")
	backupSet, err := fixture.SourceStore.GetBackupSet(ctx, fixture.LatestBackupSetID)
	if err != nil {
		t.Fatalf("load selected backup set: %v", err)
	}
	target := preparePhase10RestoreTarget(t, "phase10-i-10-03-target")
	gate := &phase10RestoreReadinessGate{}
	recorder := &phase10RestoreStepRecorder{}

	_, err = recovery.NewRestoreRunner(fixture.SourceStore, phase10TamperedBackupStorage{
		Inner:   fixture.BackupStorage,
		Missing: map[string]bool{backupSet.ObjectStoreArtifactKey: true},
	}).RestoreBackupSet(ctx, recovery.RestoreTarget{
		Postgres:    target.Postgres,
		ObjectStore: target.ObjectStore,
		Projections: projections.NewRestoreRebuilder(target.Postgres),
		Readiness:   gate,
		Observer:    recorder,
	}, backupSet)
	if err == nil {
		t.Fatal("restore with missing object artifact unexpectedly succeeded")
	}
	if gate.Marked {
		t.Fatal("restore with missing object artifact marked readiness")
	}
	if len(recorder.Steps) != 0 {
		t.Fatalf("missing object artifact should fail before target mutation steps, got %v", recorder.Steps)
	}
}

type phase10SourceBackupFixture struct {
	SourceStore          *recovery.Store
	BackupStorage        recovery.BackupStorage
	AsOf                 time.Time
	ConsistencyPointAt   time.Time
	LatestBackupSetID    uuid.UUID
	IncidentID           string
	TimelineRecordID     string
	AdminTOTPSecret      string
	ManifestEvidencePath string
	SummaryEvidencePath  string
}

type phase10RestoreTargetFixture struct {
	Env         map[string]string
	Postgres    *pgxpool.Pool
	ObjectStore objectstore.Store
}

type phase10RestoreStepRecorder struct {
	Steps []recovery.RestoreStep
}

type phase10TamperedBackupStorage struct {
	Inner        recovery.BackupStorage
	Missing      map[string]bool
	Replacements map[string][]byte
}

func (storage phase10TamperedBackupStorage) WriteArtifact(ctx context.Context, key string, body []byte, contentType string) (recovery.BackupArtifactProof, error) {
	return storage.Inner.WriteArtifact(ctx, key, body, contentType)
}

func (storage phase10TamperedBackupStorage) ReadArtifact(ctx context.Context, key string) ([]byte, error) {
	if storage.Missing[key] {
		return nil, fmt.Errorf("phase10 tampered backup artifact %s is missing", key)
	}
	if replacement, ok := storage.Replacements[key]; ok {
		return replacement, nil
	}
	return storage.Inner.ReadArtifact(ctx, key)
}

func (recorder *phase10RestoreStepRecorder) RecordRestoreStep(step recovery.RestoreStep) {
	recorder.Steps = append(recorder.Steps, step)
}

type phase10RestoreReadinessGate struct {
	Marked bool
	Result recovery.RestoreResult
}

func (gate *phase10RestoreReadinessGate) MarkRestoreReady(_ context.Context, result recovery.RestoreResult) error {
	gate.Marked = true
	gate.Result = result
	return nil
}

type phase10GuardedProjectionRebuilder struct {
	Inner restorecontract.ProjectionRebuilder
	Gate  *phase10RestoreReadinessGate
}

func (rebuilder phase10GuardedProjectionRebuilder) RebuildRestoreProjections(ctx context.Context, request restorecontract.ProjectionRebuildRequest) (restorecontract.ProjectionRebuildResult, error) {
	if rebuilder.Gate.Marked {
		return restorecontract.ProjectionRebuildResult{
			RestoreOperationID: request.RestoreOperationID,
			Status:             restorecontract.ProjectionRebuildStatusFailed,
			ReadinessOutcome:   restorecontract.ProjectionReadinessIncomplete,
		}, fmt.Errorf("readiness marked before projection rebuild started")
	}
	result, err := rebuilder.Inner.RebuildRestoreProjections(ctx, request)
	if err != nil {
		return result, err
	}
	if rebuilder.Gate.Marked {
		result.Status = restorecontract.ProjectionRebuildStatusFailed
		result.ReadinessOutcome = restorecontract.ProjectionReadinessIncomplete
		return result, fmt.Errorf("readiness marked before projection rebuild completed")
	}
	return result, nil
}

var errPhase10ProjectionBlocked = errors.New("phase10 projection rebuild blocked")

type phase10FailingProjectionRebuilder struct{}

func (phase10FailingProjectionRebuilder) RebuildRestoreProjections(_ context.Context, request restorecontract.ProjectionRebuildRequest) (restorecontract.ProjectionRebuildResult, error) {
	return restorecontract.ProjectionRebuildResult{
		RestoreOperationID: request.RestoreOperationID,
		Status:             restorecontract.ProjectionRebuildStatusFailed,
		ReadinessOutcome:   restorecontract.ProjectionReadinessIncomplete,
		Errors: []restorecontract.ProjectionRebuildMessage{{
			Code:    "phase10_projection_blocked",
			Message: errPhase10ProjectionBlocked.Error(),
		}},
	}, errPhase10ProjectionBlocked
}

func capturePhase10RestoreSource(t testing.TB, prefix string) phase10SourceBackupFixture {
	t.Helper()

	ctx := context.Background()
	postgresHarness, s3Harness := sharedProcessHarnesses(t)
	sourceDB := postgresHarness.PrepareDatabaseT(t, prefix)
	sourcePool, err := pgxpool.New(ctx, sourceDB.DSN)
	if err != nil {
		t.Fatalf("open source pgx pool: %v", err)
	}
	t.Cleanup(sourcePool.Close)

	bucket := phase0BucketName(prefix)
	t.Cleanup(func() {
		cleanupPhase0Bucket(t, s3Harness, bucket)
	})
	configPath := writePhase0Config(t, string(fixtures.MustRead("config", "valid.toml")))
	sourceEnv := phase0ServerEnv(t, sourceDB.Env(), s3Harness.Env(bucket), configPath, fixtures.Path("bootstrap-admin", "canonical.json"))
	sourceStore := phase10OpenObjectStore(t, sourceEnv)
	t.Cleanup(func() {
		_ = sourceStore.Close()
	})

	server := processtest.StartServer(t, processtest.ServerOptions{Env: sourceEnv})
	defer server.Stop(t)
	server.WaitForReady(t)

	adminLogin, adminSecret := phase1ProvisionBootstrapAdmin(t, server)
	incident := phase2CreateIncident(t, server, adminLogin.sessionCookie, adminLogin.csrfCookie, map[string]any{
		"client_txn_id": "txn-" + prefix + "-incident",
		"incident_key":  "IR-PHASE10-RESTORE",
		"title":         "Phase 10 restore source",
	})
	incidentID := incident["incident_id"].(string)
	timeline := phase5CreateViewRow(t, server, adminLogin, incidentID, "cartulary.view.timeline.v2", map[string]any{
		"client_txn_id":                   "txn-" + prefix + "-timeline",
		"timeline.activity_synopsis_text": "restore source timeline",
	})
	timelineRow := timeline["row"].(map[string]any)
	timelineRecordID := timelineRow["record_id"].(string)
	timelineRowVersion := int(timelineRow["row_version"].(float64))

	evidence := phase5CreateViewRow(t, server, adminLogin, incidentID, "cartulary.view.evidence.v1", map[string]any{
		"client_txn_id":  "txn-" + prefix + "-evidence",
		"evidence.title": "restore source evidence",
	})
	evidenceRecordID := evidence["row"].(map[string]any)["record_id"].(string)
	payload := []byte("phase10 restore coherent blob payload")
	blobCreate := phase1DoJSON(t, server, http.MethodPost, "/api/v1/object-blobs", map[string]any{
		"incident_id":       incidentID,
		"client_txn_id":     "txn-" + prefix + "-blob",
		"byte_size":         len(payload),
		"filename_hint":     "restore.txt",
		"content_type_hint": "text/plain",
	}, withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie), withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value))
	blobData := httptestx.RequireSuccessEnvelope(t, blobCreate, http.StatusCreated)["data"].(map[string]any)
	uploadTarget := blobData["upload_target"].(map[string]any)
	phase5PutObject(t, server.BaseURL, uploadTarget["href"].(string), payload, "text/plain")

	attach := phase1DoJSON(t, server, http.MethodPost, "/api/v1/evidence-records/"+evidenceRecordID+"/attach-blob", map[string]any{
		"object_blob_id":   blobData["object_blob_id"],
		"base_row_version": 1,
		"client_txn_id":    "txn-" + prefix + "-attach",
	}, withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie), withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value))
	httptestx.RequireSuccessEnvelope(t, attach, http.StatusOK)

	phase5PatchRecord(t, server, adminLogin, timelineRecordID, map[string]any{
		"view_schema_id":   "cartulary.view.timeline.v2",
		"base_row_version": timelineRowVersion,
		"client_txn_id":    "txn-" + prefix + "-link",
		"changes": []map[string]any{{
			"field_key": "timeline.attached_evidence_ids",
			"action_payload": map[string]any{
				"kind": "collection_actions_v1",
				"actions": []map[string]any{{
					"op":               "add_record_ref",
					"linked_record_id": evidenceRecordID,
				}},
			},
		}},
	})
	phase5RequireTimelineEvidenceCount(t, server, adminLogin, incidentID, timelineRecordID, 1, true)

	postgresArtifact, err := recovery.CapturePostgresSnapshotArtifact(ctx, sourcePool)
	if err != nil {
		t.Fatalf("capture source postgres restore artifact: %v", err)
	}
	if !bytes.Contains(postgresArtifact, []byte("restore source timeline")) {
		t.Fatalf("postgres restore artifact does not include source row data")
	}
	blobIndex, err := recovery.AvailableBlobObjectIDsByStorageRef(ctx, sourcePool)
	if err != nil {
		t.Fatalf("index source blob storage refs: %v", err)
	}

	olderBackupSetID := uuid.MustParse("00000000-0000-0000-0000-000000100201")
	latestBackupSetID := uuid.MustParse("00000000-0000-0000-0000-000000100202")
	asOf := time.Now().UTC().Truncate(time.Second)
	olderCreatedAt := asOf.Add(-10 * time.Minute)
	olderConsistencyPointAt := asOf.Add(-9 * time.Minute)
	olderObjectArtifacts, err := recovery.CaptureSeaweedFSS3ObjectStoreBackupArtifacts(ctx, sourceStore, recovery.ObjectStoreBackupCaptureParams{
		BackupSetID:               olderBackupSetID,
		ConsistencyPointAt:        olderConsistencyPointAt,
		Bucket:                    bucket,
		BlobObjectIDsByStorageRef: blobIndex,
	})
	if err != nil {
		t.Fatalf("capture older object-store backup artifacts: %v", err)
	}
	latestCreatedAt := asOf.Add(-2 * time.Minute)
	consistencyPointAt := asOf.Add(-time.Minute)
	latestObjectArtifacts, err := recovery.CaptureSeaweedFSS3ObjectStoreBackupArtifacts(ctx, sourceStore, recovery.ObjectStoreBackupCaptureParams{
		BackupSetID:               latestBackupSetID,
		ConsistencyPointAt:        consistencyPointAt,
		Bucket:                    bucket,
		BlobObjectIDsByStorageRef: blobIndex,
	})
	if err != nil {
		t.Fatalf("capture latest object-store backup artifacts: %v", err)
	}
	if !bytes.Contains(latestObjectArtifacts.SnapshotBody, []byte("body_base64")) {
		t.Fatalf("object-store restore artifact does not include restorable object bodies")
	}

	backupRoot := t.TempDir()
	backupStorage := phase10EncryptedBackupStorage(t, backupRoot)
	capture := recovery.NewCaptureService(recovery.NewStore(sourcePool), backupStorage)
	if _, err := capture.CaptureBackupSet(ctx, phase10CaptureParams(recovery.CaptureBackupSetParams{
		BackupSetID:        olderBackupSetID,
		ConsistencyPointAt: olderConsistencyPointAt,
		CreatedAt:          olderCreatedAt,
		RetainedUntil:      olderCreatedAt.Add(31 * 24 * time.Hour),
		PostgresArtifact: recovery.BackupArtifact{
			Body:        postgresArtifact,
			ContentType: "application/json",
		},
		ObjectStoreArtifact: recovery.BackupArtifact{
			Body:        olderObjectArtifacts.SnapshotBody,
			ContentType: "application/json",
		},
		ObjectStoreBackupManifestArtifact: recovery.BackupArtifact{Body: olderObjectArtifacts.ManifestBody, ContentType: "application/json"},
		ObjectStoreBackupSummaryArtifact:  recovery.BackupArtifact{Body: olderObjectArtifacts.SummaryBody, ContentType: "application/json"},
	})); err != nil {
		t.Fatalf("capture older retained backup set: %v", err)
	}
	if _, err := capture.CaptureBackupSet(ctx, phase10CaptureParams(recovery.CaptureBackupSetParams{
		BackupSetID:        latestBackupSetID,
		ConsistencyPointAt: consistencyPointAt,
		CreatedAt:          latestCreatedAt,
		RetainedUntil:      latestCreatedAt.Add(31 * 24 * time.Hour),
		PostgresArtifact: recovery.BackupArtifact{
			Body:        postgresArtifact,
			ContentType: "application/json",
		},
		ObjectStoreArtifact: recovery.BackupArtifact{
			Body:        latestObjectArtifacts.SnapshotBody,
			ContentType: "application/json",
		},
		ObjectStoreBackupManifestArtifact: recovery.BackupArtifact{Body: latestObjectArtifacts.ManifestBody, ContentType: "application/json"},
		ObjectStoreBackupSummaryArtifact:  recovery.BackupArtifact{Body: latestObjectArtifacts.SummaryBody, ContentType: "application/json"},
	})); err != nil {
		t.Fatalf("capture latest retained backup set: %v", err)
	}
	rawPostgresArtifact, err := os.ReadFile(filepath.Join(backupRoot, filepath.FromSlash("backup_sets/"+latestBackupSetID.String()+"/postgres-artifact.json")))
	if err != nil {
		t.Fatalf("read raw encrypted process backup artifact: %v", err)
	}
	if bytes.Contains(rawPostgresArtifact, []byte("restore source timeline")) {
		t.Fatalf("raw process backup artifact contains plaintext incident data: %s", rawPostgresArtifact)
	}
	manifestPath := phase10WriteEvidenceArtifact(t, "phase-e-backup-restore", "object-store-backup-manifest.json", latestObjectArtifacts.ManifestBody)
	summaryPath := phase10WriteEvidenceArtifact(t, "phase-e-backup-restore", "object-store-backup-summary.json", latestObjectArtifacts.SummaryBody)
	manifest, err := recovery.DecodeObjectStoreBackupManifestArtifact(latestObjectArtifacts.ManifestBody)
	if err != nil {
		t.Fatalf("decode retained object-store backup manifest: %v", err)
	}
	if manifest.BackupSetID != latestBackupSetID.String() ||
		!manifest.ConsistencyPointAt.Equal(consistencyPointAt) ||
		manifest.ObjectStoreBackend != recovery.ObjectStoreBackendSeaweedFSS3 ||
		manifest.ObjectCount == 0 {
		t.Fatalf("object-store backup manifest does not satisfy Phase E predicate at %s: %#v", manifestPath, manifest)
	}
	hasBlobID := false
	for _, object := range manifest.Objects {
		if object.SHA256 == "" || object.BackupMemberSHA256 == "" {
			t.Fatalf("manifest object has missing sha256 proof at %s: %#v", manifestPath, object)
		}
		if object.ObjectBlobID != "" {
			hasBlobID = true
		}
	}
	if !hasBlobID {
		t.Fatalf("manifest at %s did not include object_blob_id for the uploaded durable blob: %#v", manifestPath, manifest.Objects)
	}
	if _, err := recovery.DecodeObjectStoreBackupSummaryArtifact(latestObjectArtifacts.SummaryBody); err != nil {
		t.Fatalf("decode retained object-store backup summary: %v", err)
	}
	if bytes.Contains(latestObjectArtifacts.SummaryBody, []byte(bucket)) {
		t.Fatalf("shareable object-store backup summary leaked raw bucket name at %s", summaryPath)
	}
	for _, object := range manifest.Objects {
		if bytes.Contains(latestObjectArtifacts.SummaryBody, []byte(object.StorageRef)) {
			t.Fatalf("shareable object-store backup summary leaked raw storage ref at %s", summaryPath)
		}
	}
	t.Logf("phase_e_object_store_backup_manifest=%s", manifestPath)
	t.Logf("phase_e_object_store_backup_summary=%s", summaryPath)

	return phase10SourceBackupFixture{
		SourceStore:          recovery.NewStore(sourcePool),
		BackupStorage:        backupStorage,
		AsOf:                 asOf,
		ConsistencyPointAt:   consistencyPointAt,
		LatestBackupSetID:    latestBackupSetID,
		IncidentID:           incidentID,
		TimelineRecordID:     timelineRecordID,
		AdminTOTPSecret:      adminSecret,
		ManifestEvidencePath: manifestPath,
		SummaryEvidencePath:  summaryPath,
	}
}

const phase10RecoveryMasterKey = "MDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWY="

func phase10RequireStoredArtifactProof(t testing.TB, storage recovery.BackupStorage, proof recovery.BackupArtifactProof) []byte {
	t.Helper()
	body, err := recovery.VerifyArtifactProof(context.Background(), storage, proof)
	if err != nil {
		t.Fatalf("verify stored artifact proof for %s: %v", proof.Key, err)
	}
	return body
}

func phase10WriteEvidenceArtifact(t testing.TB, group string, name string, body []byte) string {
	t.Helper()
	dir := phase10EvidenceDir(group)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatalf("create Phase E evidence dir: %v", err)
	}
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, body, 0o600); err != nil {
		t.Fatalf("write Phase E evidence artifact: %v", err)
	}
	return path
}

func phase10EvidenceDir(group string) string {
	resultsDir := strings.TrimSpace(os.Getenv("CARTULARY_TEST_RESULTS_DIR"))
	runID := strings.TrimSpace(os.Getenv("CARTULARY_TEST_RUN_ID"))
	target := strings.TrimSpace(os.Getenv("CARTULARY_TEST_TARGET"))
	if target == "" {
		target = "backend-process"
	}
	if resultsDir == "" || runID == "" {
		resultsDir = filepath.Join(".cartulary", "test-results")
		runID = "local-phase-e"
	}
	return filepath.Join(resultsDir, runID, target, group)
}

func phase10EncryptedBackupStorage(t testing.TB, root string) recovery.BackupStorage {
	t.Helper()
	rawStorage, err := recovery.NewFilesystemBackupStorage(root)
	if err != nil {
		t.Fatalf("create backup storage: %v", err)
	}
	key, err := recovery.ParseRecoveryEncryptionKey(phase10RecoveryMasterKey)
	if err != nil {
		t.Fatalf("parse recovery encryption key: %v", err)
	}
	storage, err := recovery.NewEncryptedBackupStorage(rawStorage, key)
	if err != nil {
		t.Fatalf("create encrypted backup storage: %v", err)
	}
	return storage
}

func preparePhase10RestoreTarget(t testing.TB, prefix string) phase10RestoreTargetFixture {
	t.Helper()

	ctx := context.Background()
	postgresHarness, s3Harness := sharedProcessHarnesses(t)
	targetDB := postgresHarness.PrepareDatabaseT(t, prefix)
	targetPool, err := pgxpool.New(ctx, targetDB.DSN)
	if err != nil {
		t.Fatalf("open target pgx pool: %v", err)
	}
	t.Cleanup(targetPool.Close)

	bucket := phase0BucketName(prefix)
	t.Cleanup(func() {
		cleanupPhase0Bucket(t, s3Harness, bucket)
	})
	configPath := writePhase0Config(t, string(fixtures.MustRead("config", "valid.toml")))
	env := phase0ServerEnv(t, targetDB.Env(), s3Harness.Env(bucket), configPath, fixtures.Path("bootstrap-admin", "canonical.json"))
	store := phase10OpenObjectStore(t, env)
	t.Cleanup(func() {
		_ = store.Close()
	})
	return phase10RestoreTargetFixture{
		Env:         env,
		Postgres:    targetPool,
		ObjectStore: store,
	}
}

func phase10OpenObjectStore(t testing.TB, env map[string]string) objectstore.Store {
	t.Helper()
	cfg := configtest.LoadEffectiveFixture(t, []string{"config", "valid.toml"}, env)
	store, err := objectstore.SetupWithEnv(context.Background(), cfg, env)
	if err != nil {
		t.Fatalf("open object store: %v", err)
	}
	return store
}

func phase10CaptureParams(params recovery.CaptureBackupSetParams) recovery.CaptureBackupSetParams {
	if params.PostgresRestoreAnchorRetainedUntil.IsZero() {
		params.PostgresRestoreAnchorRetainedUntil = params.RetainedUntil
	}
	if params.ObjectStoreRestoreAnchorRetainedUntil.IsZero() {
		params.ObjectStoreRestoreAnchorRetainedUntil = params.RetainedUntil
	}
	return params
}

func TestPhase10_E_10_03_PublicRouteInventoryAbsence(t *testing.T) {
	postgresHarness, s3Harness := sharedProcessHarnesses(t)
	testDB := postgresHarness.PrepareDatabaseT(t, "phase10-e-10-03-route-absence")

	bucket := phase0BucketName("phase10-e-10-03")
	defer cleanupPhase0Bucket(t, s3Harness, bucket)

	configPath := writePhase0Config(t, string(fixtures.MustRead("config", "valid.toml")))
	env := phase0ServerEnv(t, testDB.Env(), s3Harness.Env(bucket), configPath, fixtures.Path("bootstrap-admin", "canonical.json"))
	server := processtest.StartServer(t, processtest.ServerOptions{Env: env})
	defer server.Stop(t)
	server.WaitForReady(t)

	for _, path := range []string{
		"/api/v1/backups",
		"/api/v1/backups/latest",
		"/api/v1/restores",
		"/api/v1/restore-verifications",
		"/ws/v1/backups",
		"/ws/v1/restores",
		"/ws/v1/restore-verifications",
	} {
		server.RequireStatus(t, path, http.StatusNotFound)
	}
}
