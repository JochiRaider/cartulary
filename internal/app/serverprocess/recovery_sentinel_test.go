package serverprocess

import (
	"bytes"
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/app/extensionassembly"
	"github.com/JochiRaider/cartulary/internal/app/recoveryassembly"
	"github.com/JochiRaider/cartulary/internal/app/timelineassembly"
	"github.com/JochiRaider/cartulary/internal/modules/evidence/recoveryprovider"
	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/configtest"
	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/processtest"
)

func TestFreshEnvironmentRestoreWorkbookConsistency_Integration(t *testing.T) {
	ctx := context.Background()
	fixture := captureRestoreSource(t, "backup_restore-i-10-02-source")
	target := prepareRestoreTarget(t, "backup_restore-i-10-02-target")
	gate := &RestoreReadinessGate{}

	result, err := recovery.NewRestoreRunner(fixture.SourceStore, fixture.BackupStorage, RecoveryExtensionCatalog(t)).RestoreLatestSuccessfulRetained(ctx, recovery.RestoreTarget{
		Stopped:         true,
		Postgres:        target.Postgres,
		ObjectStore:     target.ObjectStore,
		EvidenceObjects: recoveryprovider.New(target.Postgres),
		Projections:     timelineassembly.NewRestoreRebuilder(target.Postgres),
		Readiness:       gate,
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

	basis, err := recovery.RestoreVerificationBasisSHA256(map[string]string{
		"backup_mechanism": "backup_restore.process.restore.v1",
		"source_root":      "backup_restore-i-10-02-source",
	})
	if err != nil {
		t.Fatalf("restore verification basis: %v", err)
	}
	verificationTarget := prepareRestoreTarget(t, "backup_restore-i-10-02-verification-target")
	verificationRebuilder, verificationQuery := timelineassembly.NewRecoveryProjectionServices(verificationTarget.Postgres)
	verification, err := recovery.NewRestoreVerificationService(
		fixture.SourceStore,
		recovery.NewRestoreRunner(fixture.SourceStore, fixture.BackupStorage, RecoveryExtensionCatalog(t)),
	).VerifyLatestSuccessfulRetained(ctx, recovery.RestoreVerificationTarget{
		RestoreTarget: recovery.RestoreTarget{
			Stopped:         true,
			Postgres:        verificationTarget.Postgres,
			ObjectStore:     verificationTarget.ObjectStore,
			EvidenceObjects: recoveryprovider.New(verificationTarget.Postgres),
			Projections:     verificationRebuilder,
		},
		Probe: recovery.RestoreVerificationWorkbookProbe{Postgres: verificationTarget.Postgres, Query: verificationQuery},
	}, fixture.AsOf, basis)
	if err != nil {
		t.Fatalf("verify restored process fixture: %v", err)
	}
	if verification.BackupSet.VerificationState != recovery.VerificationVerified || verification.Artifact.Result != "pass" {
		t.Fatalf("restore verification did not pass: %#v", verification)
	}
	verificationBody := RequireStoredArtifactProof(t, fixture.BackupStorage, verification.ArtifactProof)
	verificationPath := WriteEvidenceArtifact(t, "backup-restore", "restore-verification.json", verificationBody)
	t.Logf("backup_restore_verification_artifact=%s", verificationPath)

	server := processtest.StartServer(t, processtest.ServerOptions{Env: target.Env})
	defer server.Stop(t)
	server.WaitForReady(t)

	login := LoginLocalUserWithSecondFactor(t, server, authenticationBootstrapAdminEmail, authenticationBootstrapAdminPassword, GenerateTOTPCode(t, fixture.AdminTOTPSecret))
	getIncident := DoJSON(t, server, http.MethodGet, "/api/v1/incidents/"+fixture.IncidentID, nil, withCookies(login.sessionCookie))
	httptestx.RequireSuccessEnvelope(t, getIncident, http.StatusOK)
	RequireTimelineEvidenceCount(t, server, login, fixture.IncidentID, fixture.TimelineRecordID, 1, true)
}

type SourceBackupFixture struct {
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

type RestoreTargetFixture struct {
	Env         map[string]string
	Postgres    *pgxpool.Pool
	ObjectStore objectstore.Store
}

type RestoreReadinessGate struct {
	Marked bool
	Result recovery.RestoreResult
}

func (gate *RestoreReadinessGate) MarkRestoreReady(_ context.Context, result recovery.RestoreResult) error {
	gate.Marked = true
	gate.Result = result
	return nil
}

func captureRestoreSource(t testing.TB, prefix string) SourceBackupFixture {
	t.Helper()

	ctx := context.Background()
	postgresHarness, s3Harness := sharedProcessHarnesses(t)
	sourceDB := postgresHarness.PrepareIsolatedDatabaseT(t, prefix)
	sourcePool, err := pgxpool.New(ctx, sourceDB.DSN)
	if err != nil {
		t.Fatalf("open source pgx pool: %v", err)
	}
	t.Cleanup(sourcePool.Close)

	bucket := BucketName(prefix)
	t.Cleanup(func() {
		cleanupBucket(t, s3Harness, bucket)
	})
	configPath := writeConfig(t, string(fixtures.MustRead("config", "valid.toml")))
	sourceEnv := ServerEnv(t, sourceDB.Env(), s3Harness.Env(bucket), configPath, fixtures.Path("bootstrap-admin", "canonical.json"))
	sourceStore := OpenObjectStore(t, sourceEnv)
	t.Cleanup(func() {
		_ = sourceStore.Close()
	})

	server := processtest.StartServer(t, processtest.ServerOptions{Env: sourceEnv})
	defer server.Stop(t)
	server.WaitForReady(t)

	adminLogin, adminSecret := ProvisionBootstrapAdmin(t, server)
	incident := CreateIncident(t, server, adminLogin.sessionCookie, adminLogin.csrfCookie, map[string]any{
		"client_txn_id": "txn-" + prefix + "-incident",
		"incident_key":  "IR-BACKUP-RESTORE-RESTORE",
		"title":         "Recovery and coordination restore source",
	})
	incidentID := incident["incident_id"].(string)
	timeline := CreateViewRow(t, server, adminLogin, incidentID, "cartulary.view.timeline.v2", map[string]any{
		"client_txn_id":                   "txn-" + prefix + "-timeline",
		"timeline.activity_synopsis_text": "restore source timeline",
	})
	timelineRow := timeline["row"].(map[string]any)
	timelineRecordID := timelineRow["record_id"].(string)
	timelineRowVersion := int(timelineRow["row_version"].(float64))

	evidence := CreateViewRow(t, server, adminLogin, incidentID, "cartulary.view.evidence.v1", map[string]any{
		"client_txn_id":  "txn-" + prefix + "-evidence",
		"evidence.title": "restore source evidence",
	})
	evidenceRecordID := evidence["row"].(map[string]any)["record_id"].(string)
	payload := []byte("backup_restore restore coherent blob payload")
	blobCreate := DoJSON(t, server, http.MethodPost, "/api/v1/object-blobs", map[string]any{
		"incident_id":       incidentID,
		"client_txn_id":     "txn-" + prefix + "-blob",
		"byte_size":         len(payload),
		"filename_hint":     "restore.txt",
		"content_type_hint": "text/plain",
	}, withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie), withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value))
	blobData := httptestx.RequireSuccessEnvelope(t, blobCreate, http.StatusCreated)["data"].(map[string]any)
	uploadTarget := blobData["upload_target"].(map[string]any)
	PutObject(t, server.BaseURL, uploadTarget["href"].(string), payload, "text/plain")

	attach := DoJSON(t, server, http.MethodPost, "/api/v1/evidence-records/"+evidenceRecordID+"/attach-blob", map[string]any{
		"object_blob_id":   blobData["object_blob_id"],
		"base_row_version": 1,
		"client_txn_id":    "txn-" + prefix + "-attach",
	}, withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie), withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value))
	httptestx.RequireSuccessEnvelope(t, attach, http.StatusOK)

	PatchRecord(t, server, adminLogin, timelineRecordID, map[string]any{
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
	RequireTimelineEvidenceCount(t, server, adminLogin, incidentID, timelineRecordID, 1, true)

	postgresArtifact, err := recovery.CapturePostgresSnapshotArtifact(ctx, sourcePool)
	if err != nil {
		t.Fatalf("capture source postgres restore artifact: %v", err)
	}
	if !bytes.Contains(postgresArtifact, []byte("restore source timeline")) {
		t.Fatalf("postgres restore artifact does not include source row data")
	}
	blobIndex, err := recovery.AvailableBlobObjectIDsByStorageRef(ctx, recoveryprovider.New(sourcePool))
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
	backupStorage := EncryptedBackupStorage(t, backupRoot)
	capture := recovery.NewCaptureService(recovery.NewStore(sourcePool), backupStorage, RecoveryExtensionCatalog(t))
	if _, err := capture.CaptureBackupSet(ctx, CaptureParams(recovery.CaptureBackupSetParams{
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
	if _, err := capture.CaptureBackupSet(ctx, CaptureParams(recovery.CaptureBackupSetParams{
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
	manifestPath := WriteEvidenceArtifact(t, "backup-restore", "object-store-backup-manifest.json", latestObjectArtifacts.ManifestBody)
	summaryPath := WriteEvidenceArtifact(t, "backup-restore", "object-store-backup-summary.json", latestObjectArtifacts.SummaryBody)
	manifest, err := recovery.DecodeObjectStoreBackupManifestArtifact(latestObjectArtifacts.ManifestBody)
	if err != nil {
		t.Fatalf("decode retained object-store backup manifest: %v", err)
	}
	if manifest.BackupSetID != latestBackupSetID.String() ||
		!manifest.ConsistencyPointAt.Equal(consistencyPointAt) ||
		manifest.ObjectStoreBackend != recovery.ObjectStoreBackendSeaweedFSS3 ||
		manifest.ObjectCount == 0 {
		t.Fatalf("object-store backup manifest does not satisfy backup predicate at %s: %#v", manifestPath, manifest)
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
	t.Logf("object_store_backup_manifest=%s", manifestPath)
	t.Logf("object_store_backup_summary=%s", summaryPath)

	return SourceBackupFixture{
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

const RecoveryMasterKey = "MDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWY="

func RequireStoredArtifactProof(t testing.TB, storage recovery.BackupStorage, proof recovery.BackupArtifactProof) []byte {
	t.Helper()
	body, err := recovery.VerifyArtifactProof(context.Background(), storage, proof)
	if err != nil {
		t.Fatalf("verify stored artifact proof for %s: %v", proof.Key, err)
	}
	return body
}

func WriteEvidenceArtifact(t testing.TB, group string, name string, body []byte) string {
	t.Helper()
	dir := EvidenceDir(group)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatalf("create backup-restore evidence dir: %v", err)
	}
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, body, 0o600); err != nil {
		t.Fatalf("write backup-restore evidence artifact: %v", err)
	}
	return path
}

func EvidenceDir(group string) string {
	resultsDir := strings.TrimSpace(os.Getenv("CARTULARY_TEST_RESULTS_DIR"))
	runID := strings.TrimSpace(os.Getenv("CARTULARY_TEST_RUN_ID"))
	target := strings.TrimSpace(os.Getenv("CARTULARY_TEST_TARGET"))
	if target == "" {
		target = "backend-process"
	}
	if resultsDir == "" || runID == "" {
		resultsDir = filepath.Join(".cartulary", "test-results")
		runID = "local-backup-restore"
	}
	return filepath.Join(resultsDir, runID, target, group)
}

func EncryptedBackupStorage(t testing.TB, root string) recovery.BackupStorage {
	t.Helper()
	rawStorage, err := recoveryassembly.NewFilesystemStorage(root)
	if err != nil {
		t.Fatalf("create backup storage: %v", err)
	}
	key, err := recovery.ParseRecoveryEncryptionKey(RecoveryMasterKey)
	if err != nil {
		t.Fatalf("parse recovery encryption key: %v", err)
	}
	storage, err := recovery.NewEncryptedBackupStorage(rawStorage, key)
	if err != nil {
		t.Fatalf("create encrypted backup storage: %v", err)
	}
	return storage
}

func prepareRestoreTarget(t testing.TB, prefix string) RestoreTargetFixture {
	t.Helper()

	ctx := context.Background()
	postgresHarness, s3Harness := sharedProcessHarnesses(t)
	targetDB := postgresHarness.PrepareIsolatedDatabaseT(t, prefix)
	targetPool, err := pgxpool.New(ctx, targetDB.DSN)
	if err != nil {
		t.Fatalf("open target pgx pool: %v", err)
	}
	t.Cleanup(targetPool.Close)

	bucket := BucketName(prefix)
	t.Cleanup(func() {
		cleanupBucket(t, s3Harness, bucket)
	})
	configPath := writeConfig(t, string(fixtures.MustRead("config", "valid.toml")))
	env := ServerEnv(t, targetDB.Env(), s3Harness.Env(bucket), configPath, fixtures.Path("bootstrap-admin", "canonical.json"))
	store := OpenObjectStore(t, env)
	t.Cleanup(func() {
		_ = store.Close()
	})
	return RestoreTargetFixture{
		Env:         env,
		Postgres:    targetPool,
		ObjectStore: store,
	}
}

func OpenObjectStore(t testing.TB, env map[string]string) objectstore.Store {
	t.Helper()
	cfg := configtest.LoadEffectiveFixture(t, []string{"config", "valid.toml"}, env)
	store, err := appsupport.OpenObjectStore(context.Background(), cfg, env)
	if err != nil {
		t.Fatalf("open object store: %v", err)
	}
	return store
}

func CaptureParams(params recovery.CaptureBackupSetParams) recovery.CaptureBackupSetParams {
	if params.PostgresRestoreAnchorRetainedUntil.IsZero() {
		params.PostgresRestoreAnchorRetainedUntil = params.RetainedUntil
	}
	if params.ObjectStoreRestoreAnchorRetainedUntil.IsZero() {
		params.ObjectStoreRestoreAnchorRetainedUntil = params.RetainedUntil
	}
	return params
}

func RecoveryExtensionCatalog(t testing.TB) *recovery.ExtensionBackupCatalog {
	t.Helper()
	catalog, err := extensionassembly.GeneratedRecoveryCatalog()
	if err != nil {
		t.Fatalf("construct extension recovery catalog: %v", err)
	}
	return catalog
}

func TestPublicRouteInventoryAbsence_Process(t *testing.T) {
	postgresHarness, s3Harness := sharedProcessHarnesses(t)
	testDB := postgresHarness.PrepareIsolatedDatabaseT(t, "backup_restore-e-10-03-route-absence")

	bucket := BucketName("backup_restore-e-10-03")
	defer cleanupBucket(t, s3Harness, bucket)

	configPath := writeConfig(t, string(fixtures.MustRead("config", "valid.toml")))
	env := ServerEnv(t, testDB.Env(), s3Harness.Env(bucket), configPath, fixtures.Path("bootstrap-admin", "canonical.json"))
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
