package serverprocess

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"github.com/JochiRaider/cartulary/internal/testutil/recoverytest"
	verificationtest "github.com/JochiRaider/cartulary/internal/testutil/referencetest"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/app/configassembly"
	"github.com/JochiRaider/cartulary/internal/app/extensionassembly"
	"github.com/JochiRaider/cartulary/internal/app/incidentportabilityassembly"
	"github.com/JochiRaider/cartulary/internal/app/projectionassembly"
	"github.com/JochiRaider/cartulary/internal/app/recoveryassembly"
	"github.com/JochiRaider/cartulary/internal/app/referenceassembly"
	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	evidencemodule "github.com/JochiRaider/cartulary/internal/modules/evidence"
	evidenceflow "github.com/JochiRaider/cartulary/internal/modules/evidence/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	"github.com/JochiRaider/cartulary/internal/modules/recovery/application"
	recoverytestsupport "github.com/JochiRaider/cartulary/internal/modules/recovery/testsupport"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/authcookietest"
	"github.com/JochiRaider/cartulary/internal/testutil/configtest"
	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
	"github.com/JochiRaider/cartulary/internal/testutil/processtest"
	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices"
)

func TestFreshEnvironmentRestoreWorkbookConsistency_Integration(t *testing.T) {
	ctx := t.Context()
	fixture := captureRestoreSource(t, "backup_restore-i-10-02-source")
	target := prepareRestoreTarget(t, "backup_restore-i-10-02-target")
	projectionRuntime, err := projectionassembly.Build(target.Postgres)
	if err != nil {
		t.Fatalf("compose restore projection runtime: %v", err)
	}

	stateCatalog, err := recoveryassembly.CurrentRecoveryStateCatalog()
	if err != nil {
		t.Fatal(err)
	}
	packs := openRecoveryPackStorage(t, target.Env)
	runner := recovery.NewVersionedRestoreRunner(fixture.SourceStore, fixture.BackupStorage, recoveryExtensionCatalog(t), stateCatalog)
	admitted := admitProcessRestore(t, fixture, target, runner, recovery.RestoreTarget{ProofScratch: recoverytest.ProofScratch(t), Postgres: target.RecoveryPool, ObjectStore: target.ObjectStore, ReferencePacks: packs, EvidenceObjects: evidencemodule.NewRecoveryProvider(target.Postgres), Projections: projectionRuntime.RecoveryPorts().Rebuilder}, application.OperationRestoreLatest)
	restored, err := runner.RestoreLatestSuccessfulRetained(ctx, admitted.Fence.Target(), fixture.AsOf)
	if err != nil {
		t.Fatalf("restore current process fixture: %v", err)
	}
	if restored.BackupSet.BackupSetID != fixture.LatestBackupSetID || restored.ConsistencyReport.BlobCount != 1 || restored.RestoredObjectCount != fixture.ObjectCount || restored.RestoreGeneration == nil {
		t.Fatalf("incomplete current restore: %#v", restored)
	}
	admitted.Complete(t, restored)
	var restoredProofs int
	if err := target.Postgres.QueryRow(ctx, `SELECT count(*) FROM evidence_blob_verifications`).Scan(&restoredProofs); err != nil {
		t.Fatal(err)
	}
	if restoredProofs != 0 {
		t.Fatalf("restore retained %d security-ephemeral verification proofs", restoredProofs)
	}

	basis := recovery.RestoreVerificationBasis{
		ApplicationCryptoFormat:           recovery.ApplicationCryptoFormatID,
		MechanismID:                       "backup_restore.process.restore.v1",
		DatabaseBindingSHA256:             recovery.SHA256String("backup_restore-i-10-02-database"),
		ObjectStoreBindingSHA256:          recovery.SHA256String("backup_restore-i-10-02-objects"),
		ReferencePackStorageBindingSHA256: recovery.SHA256String("backup_restore-i-10-02-reference-packs"),
		ExportOutputsBindingSHA256:        recovery.SHA256String("test-export-root"),
		BackupStorageBindingSHA256:        recovery.SHA256String("backup_restore-i-10-02-backups"),
		RecoveryStateCatalogSHA256:        recovery.SHA256String("backup_restore-i-10-02-catalog"),
		CodecRegistrySHA256:               recovery.SHA256String("backup_restore-i-10-02-codecs"),
	}
	verificationTarget := prepareRestoreTarget(t, "backup_restore-i-10-02-verification-target")
	verificationRebuilder, verificationQuery, err := projectionassembly.NewRecoveryServices(verificationTarget.Postgres)
	if err != nil {
		t.Fatalf("compose restore projection services: %v", err)
	}
	verificationPacks := openRecoveryPackStorage(t, verificationTarget.Env)
	verificationAdmission := admitProcessRestore(t, fixture, verificationTarget, runner, recovery.RestoreTarget{ProofScratch: recoverytest.ProofScratch(t), Postgres: verificationTarget.RecoveryPool, ObjectStore: verificationTarget.ObjectStore, ReferencePacks: verificationPacks, EvidenceObjects: evidencemodule.NewRecoveryProvider(verificationTarget.Postgres), Projections: verificationRebuilder}, application.OperationRestoreVerifyLatest)
	verification, err := recovery.NewRestoreVerificationService(fixture.SourceStore, runner).VerifyLatestSuccessfulRetained(ctx, recovery.RestoreVerificationTarget{RestoreTarget: verificationAdmission.Fence.Target(), Probe: recovery.RestoreVerificationWorkbookProbe{Executor: verificationQuery}}, fixture.AsOf, basis)
	if err != nil {
		t.Fatalf("verify restored process fixture: %v", err)
	}
	if verification.BackupSet.VerificationState != recovery.VerificationVerified || verification.Artifact.Result != "pass" {
		t.Fatalf("restore verification did not pass: %#v", verification)
	}
	verificationAdmission.Complete(t, verification.RestoreResult)
	verificationPath := recoverytestsupport.RequireVerificationArtifact(t, fixture.BackupStorage, verification, recoveryEvidenceLocation(t, "backup-restore"), recoverytestsupport.VerificationExpectation{
		BackupSetID:        fixture.LatestBackupSetID,
		ConsistencyPointAt: fixture.ConsistencyPointAt,
		IncidentID:         fixture.IncidentID,
		ObjectCount:        fixture.ObjectCount,
		RegistrationID:     "timeline.base_restore_probe.v1",
		ViewSchemaID:       "cartulary.view.timeline.v2",
	})
	t.Logf("backup_restore_verification_artifact=%s", verificationPath)

	settle, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	if err := admitted.Admission.Release(settle); err != nil {
		cancel()
		t.Fatal(err)
	}
	cancel()
	delete(target.Env, recovery.RecoveryMasterKeyEnv)
	server := processtest.StartServer(t, processtest.ServerOptions{Env: target.Env})
	defer server.Stop(t)
	server.WaitForReady(t)

	login := flowtest.LoginLocalUserWithSecondFactor(t, server.Client, server.BaseURL, authenticationBootstrapAdminEmail, authenticationBootstrapAdminPassword, flowtest.GenerateTOTPCode(t, fixture.AdminTOTPSecret))
	getIncident := doJSON(t, server, http.MethodGet, "/api/v1/incidents/"+fixture.IncidentID, nil, httptestx.WithCookies(login.Session))
	incidentData := httptestx.RequireSuccessEnvelope(t, getIncident, http.StatusOK)["data"].(map[string]any)
	if incidentData["incident_id"] != fixture.IncidentID {
		t.Fatalf("restored incident identity got %v want %s", incidentData["incident_id"], fixture.IncidentID)
	}
	restoredTimeline := requireTimelineEvidenceCount(t, server, login, fixture.IncidentID, fixture.TimelineRecordID, 1, true)
	if got := int(restoredTimeline["row_version"].(float64)); got != fixture.TimelineRowVersion {
		t.Fatalf("restored Timeline row_version got %d want %d", got, fixture.TimelineRowVersion)
	}
}

type sourceBackupFixture struct {
	sourcePool         *pgxpool.Pool
	SourceStore        *recovery.Store
	BackupStorage      recovery.BackupStorage
	AsOf               time.Time
	ConsistencyPointAt time.Time
	LatestBackupSetID  uuid.UUID
	IncidentID         string
	TimelineRecordID   string
	TimelineRowVersion int
	ObjectCount        int64
	AdminTOTPSecret    string
}

func captureRestoreSource(t testing.TB, prefix string) sourceBackupFixture {
	return captureRestoreSourceWithNetworkFlow(t, prefix, nil)
}

func captureRestoreSourceWithNetworkFlow(t testing.TB, prefix string, prepare func(*processtest.Server, *pgxpool.Pool, authcookietest.AuthCookies, string)) sourceBackupFixture {
	t.Helper()

	ctx := t.Context()
	postgresHarness, s3Harness := sharedProcessHarnesses(t)
	sourceDB := postgresHarness.PrepareIsolatedDatabaseT(t, prefix)
	sourcePool, err := pgxpool.New(ctx, sourceDB.DSN)
	if err != nil {
		t.Fatalf("open source pgx pool: %v", err)
	}
	t.Cleanup(sourcePool.Close)

	bucket := bucketName(prefix)
	t.Cleanup(func() {
		cleanupBucket(t, s3Harness, bucket)
	})
	configPath := writeConfig(t, string(fixtures.MustRead("config", "valid.toml")))
	sourceEnv := newProcessEnv(t, processEnvOptions{Database: sourceDB.Env(), ObjectStore: s3Harness.Env(bucket), ConfigPath: configPath, BootstrapPath: fixtures.Path("bootstrap-admin", "canonical.json")})
	if prepare != nil {
		for key, value := range restoredNetworkFlowClaimEnv() {
			sourceEnv[key] = value
		}
	}
	sourceObjectStore := openObjectStore(t, sourceEnv)
	t.Cleanup(func() {
		if err := sourceObjectStore.Close(); err != nil {
			t.Errorf("close source object store: %v", err)
		}
	})

	server := processtest.StartServer(t, processtest.ServerOptions{Env: sourceEnv})
	defer server.Stop(t)
	server.WaitForReady(t)

	adminLogin, adminSecret := provisionBootstrapAdmin(t, server)
	incident := createIncident(t, server, adminLogin.Session, adminLogin.CSRF, map[string]any{
		"client_txn_id": "txn-" + prefix + "-incident",
		"incident_key":  "IR-BACKUP-RESTORE-RESTORE",
		"title":         "Recovery and coordination restore source",
	})
	incidentID := incident["incident_id"].(string)
	timeline := createViewRow(t, server, adminLogin, incidentID, "cartulary.view.timeline.v2", map[string]any{
		"client_txn_id":                   "txn-" + prefix + "-timeline",
		"timeline.activity_synopsis_text": "restore source timeline",
	})
	timelineRow := timeline["row"].(map[string]any)
	timelineRecordID := timelineRow["record_id"].(string)
	timelineRowVersion := int(timelineRow["row_version"].(float64))

	evidence := createViewRow(t, server, adminLogin, incidentID, "cartulary.view.evidence.v1", map[string]any{
		"client_txn_id":  "txn-" + prefix + "-evidence",
		"evidence.title": "restore source evidence",
	})
	evidenceRecordID := evidence["row"].(map[string]any)["record_id"].(string)
	payload := []byte("backup_restore restore coherent blob payload")
	blobCreate := doJSON(t, server, http.MethodPost, "/api/v1/object-blobs", map[string]any{
		"incident_id":       incidentID,
		"client_txn_id":     "txn-" + prefix + "-blob",
		"byte_size":         len(payload),
		"filename_hint":     "restore.txt",
		"content_type_hint": "text/plain",
	}, httptestx.WithCookies(adminLogin.Session, adminLogin.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRF.Value))
	blobData := httptestx.RequireSuccessEnvelope(t, blobCreate, http.StatusCreated)["data"].(map[string]any)
	uploadTarget := blobData["upload_target"].(map[string]any)
	objectBlobID := blobData["object_blob_id"].(string)
	putObject(t, server, uploadTarget, payload, adminLogin)
	evidenceflow.VerifyUploadedBlob(t, server.Client, server.BaseURL, adminLogin.Session, adminLogin.CSRF, blobData["object_blob_id"].(string), "txn-verify-uploaded-"+blobData["object_blob_id"].(string))

	attach := doJSON(t, server, http.MethodPost, "/api/v1/evidence-records/"+evidenceRecordID+"/attach-blob", map[string]any{
		"object_blob_id":   objectBlobID,
		"base_row_version": 1,
		"client_txn_id":    "txn-" + prefix + "-attach",
	}, httptestx.WithCookies(adminLogin.Session, adminLogin.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRF.Value))
	attachData := httptestx.RequireSuccessEnvelope(t, attach, http.StatusOK)["data"].(map[string]any)
	attachedRow := attachData["row"].(map[string]any)
	patchRecord(t, server, adminLogin, evidenceRecordID, map[string]any{
		"view_schema_id":   "cartulary.view.evidence.v1",
		"base_row_version": attachedRow["row_version"],
		"client_txn_id":    "txn-" + prefix + "-available",
		"changes": []map[string]any{{
			"field_key": "evidence.lifecycle_state",
			"value":     "available",
		}},
	})

	patchRecord(t, server, adminLogin, timelineRecordID, map[string]any{
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
	linkedTimeline := requireTimelineEvidenceCount(t, server, adminLogin, incidentID, timelineRecordID, 1, true)
	linkedTimelineRowVersion := int(linkedTimeline["row_version"].(float64))
	if prepare != nil {
		prepare(server, sourcePool, adminLogin, incidentID)
	}
	latestBackupSetID := uuid.New()
	asOf := time.Now().UTC().Truncate(time.Second)
	consistencyPointAt := asOf.Add(-time.Minute)
	backupStorage := encryptedBackupStorage(t, t.TempDir())
	sourceRecoveryStore := recovery.NewStore(sourcePool)
	catalog, err := recoveryassembly.CurrentRecoveryStateCatalog()
	if err != nil {
		t.Fatal(err)
	}
	packs := openRecoveryPackStorage(t, sourceEnv)
	inventories, err := recoveryassembly.CurrentVNextObjectInventoryCatalog(recoveryassembly.NewVNextObjectSource(sourceObjectStore), packs, nil)
	if err != nil {
		t.Fatal(err)
	}
	streaming, err := recovery.RequireStreamingBackupStorage(backupStorage)
	if err != nil {
		t.Fatal(err)
	}
	capture, err := recovery.NewVNextCaptureService(recoveryassembly.NewVNextSnapshotRepository(sourcePool), streaming, catalog, inventories, recoveryExtensionCatalog(t), recoverytest.ProofScratch(t))
	if err != nil {
		t.Fatal(err)
	}
	var sourceProofs int
	if err := sourcePool.QueryRow(ctx, `SELECT count(*) FROM evidence_blob_verifications`).Scan(&sourceProofs); err != nil {
		t.Fatal(err)
	}
	if sourceProofs != 1 {
		t.Fatalf("source fixture requires one real verification proof, got %d", sourceProofs)
	}
	captured, err := capture.Capture(ctx, recovery.VNextCaptureParams{
		BackupSetID: latestBackupSetID, ConsistencyPointAt: consistencyPointAt,
		CreatedAt: consistencyPointAt.Add(-time.Minute), RetainedUntil: asOf.Add(31 * 24 * time.Hour),
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := sourceRecoveryStore.PublishVNextCapturedBackup(ctx, captured); err != nil {
		t.Fatal(err)
	}
	var packCount int64
	if err := sourcePool.QueryRow(ctx, `SELECT count(*) FROM reference_pack_objects o WHERE EXISTS(SELECT 1 FROM reference_pack_object_refs r WHERE r.object_id=o.object_id)`).Scan(&packCount); err != nil {
		t.Fatal(err)
	}
	if packCount < 9 {
		t.Fatal("Base bytes were absent from process backup")
	}
	return sourceBackupFixture{
		SourceStore: sourceRecoveryStore, BackupStorage: backupStorage, AsOf: asOf, sourcePool: sourcePool,
		ConsistencyPointAt: consistencyPointAt, LatestBackupSetID: latestBackupSetID,
		IncidentID: incidentID, TimelineRecordID: timelineRecordID,
		TimelineRowVersion: linkedTimelineRowVersion, ObjectCount: packCount + 1, AdminTOTPSecret: adminSecret,
	}

}

const recoveryMasterKey = "MDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWY="

func recoveryEvidenceLocation(t testing.TB, group string) recoverytestsupport.EvidenceLocation {
	t.Helper()

	resultsDir, err := suiteservices.ResolveResultsRoot(nil)
	if err != nil {
		t.Fatalf("resolve Recovery results root: %v", err)
	}
	runID := suiteservices.ResolveRunID(nil)
	target := strings.TrimSpace(suiteservices.LookupEnvValue(nil, suiteservices.TargetEnv))
	if target == "" {
		target = "backend-process"
	}
	return recoverytestsupport.EvidenceLocation{
		ResultsRoot: resultsDir,
		RunID:       runID,
		Target:      target,
		Group:       group,
	}
}

func encryptedBackupStorage(t testing.TB, root string) recovery.BackupStorage {
	t.Helper()
	rawStorage, err := recoveryassembly.NewFilesystemStorage(root)
	if err != nil {
		t.Fatalf("create backup storage: %v", err)
	}
	key, err := recovery.ParseRecoveryEncryptionKey(recoveryMasterKey)
	if err != nil {
		t.Fatalf("parse recovery encryption key: %v", err)
	}
	storage, err := recovery.NewEncryptedBackupStorage(rawStorage, key)
	if err != nil {
		t.Fatalf("create encrypted backup storage: %v", err)
	}
	return storage
}

type processRestoreTarget struct {
	recoverytestsupport.TargetFixture
	RecoveryPool postgres.AdmittedPool
	database     *pgtest.TestDatabase
}

func prepareRestoreTarget(t testing.TB, prefix string) processRestoreTarget {
	t.Helper()

	ctx := t.Context()
	postgresHarness, s3Harness := sharedProcessHarnesses(t)
	targetDB := postgresHarness.PrepareIsolatedDatabaseT(t, prefix)
	dsn, err := targetDB.DSNForPurpose(postgres.PurposeRecovery)
	if err != nil {
		t.Fatal(err)
	}
	pool, err := postgres.Setup(ctx, postgres.Settings{BindingKind: "managed_service", DSN: dsn, Purpose: postgres.PurposeRecovery, ExpectedRole: "cartulary_recovery"})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	targetPool := pool.Pool()

	bucket := bucketName(prefix)
	t.Cleanup(func() {
		cleanupBucket(t, s3Harness, bucket)
	})
	configPath := writeConfig(t, string(fixtures.MustRead("config", "valid.toml")))
	env := newProcessEnv(t, processEnvOptions{Database: targetDB.Env(), ObjectStore: s3Harness.Env(bucket), ConfigPath: configPath, BootstrapPath: fixtures.Path("bootstrap-admin", "canonical.json")})
	store := openObjectStore(t, env)
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Errorf("close target object store: %v", err)
		}
	})
	target := recoverytestsupport.NewTargetFixture(env, targetPool, store)
	t.Cleanup(target.Cleanup)
	return processRestoreTarget{TargetFixture: target, RecoveryPool: pool, database: targetDB}
}

func openRecoveryPackStorage(t testing.TB, env map[string]string) *referenceassembly.RecoveryStorage {
	t.Helper()
	cfg := configtest.LoadFixture(t, []string{"config", "valid.toml"}, env).Deployment()
	packs, err := referenceassembly.NewRecoveryStorage(cfg.Roots.TemporaryWork.Path, cfg.Roots.ReferencePackStorage.Path, reference_data.DefaultLimits(), verificationtest.Lifetime(t))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(packs.Close)
	return packs
}

func openObjectStore(t testing.TB, env map[string]string) objectstore.Store {
	t.Helper()
	cfg := configtest.LoadFixture(t, []string{"config", "valid.toml"}, env).Deployment()
	store, err := appsupport.OpenObjectStore(t.Context(), cfg, env)
	if err != nil {
		t.Fatalf("open object store: %v", err)
	}
	return store
}

func recoveryExtensionCatalog(t testing.TB) *recovery.ExtensionBackupCatalog {
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

	bucket := bucketName("backup_restore-e-10-03")
	defer cleanupBucket(t, s3Harness, bucket)

	configPath := writeConfig(t, string(fixtures.MustRead("config", "valid.toml")))
	env := newProcessEnv(t, processEnvOptions{Database: testDB.Env(), ObjectStore: s3Harness.Env(bucket), ConfigPath: configPath, BootstrapPath: fixtures.Path("bootstrap-admin", "canonical.json")})
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

// Current capture/restore entry points and durable evidence precede the actual
// application process. The fixture never writes a complete control row itself.
type processRestoreAdmission struct {
	Fence      *recovery.FencedRestore
	Admission  application.TargetServingAdmission
	Identity   recovery.RestoreTargetIdentity
	Repository application.RecoveryEvidenceRepository
	StartedAt  time.Time
}

func admitProcessRestore(t testing.TB, source sourceBackupFixture, target processRestoreTarget, runner *recovery.RestoreRunner, resources recovery.RestoreTarget, operation application.Operation) processRestoreAdmission {
	t.Helper()
	ctx := t.Context()
	catalog, err := recoveryassembly.CurrentRecoveryStateCatalog()
	if err != nil {
		t.Fatal(err)
	}
	backups := recovery.NewBackupCatalog(source.SourceStore, source.BackupStorage, recoveryExtensionCatalog(t), catalog)
	backup, err := backups.RestoreCandidateBackup(ctx, source.AsOf)
	if err != nil {
		t.Fatal(err)
	}
	generation, err := backups.RecoveryGenerationIdentity(ctx, backup)
	if err != nil {
		t.Fatal(err)
	}
	cfg := configtest.LoadFixture(t, []string{"config", "valid.toml"}, target.Env).Deployment()
	bindings, err := configassembly.RecoveryTargetBindings(cfg, target.Env)
	if err != nil {
		t.Fatal(err)
	}
	admission, err := recoveryassembly.AcquireTargetServingAdmission(ctx, target.RecoveryPool, 5*time.Second, 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		settle, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := admission.Release(settle); err != nil {
			t.Error(err)
		}
	})
	purpose := application.RestoreTargetPurpose
	if operation == application.OperationRestoreVerifyLatest {
		purpose = application.RestoreVerificationTargetPurpose
	}
	identity := recovery.RestoreTargetIdentity{TargetGenerationID: uuid.New(), Purpose: purpose, Operation: string(operation), OperationID: uuid.New(), BackupSetID: backup.BackupSetID, RecoveryStateCatalogSHA256: generation.RecoveryStateCatalogSHA256, CodecRegistrySHA256: generation.CodecRegistrySHA256, BindingDigests: bindings}
	key, err := recovery.ParseRecoveryEncryptionKey(recoveryMasterKey)
	if err != nil {
		t.Fatal(err)
	}
	repository, err := recoveryassembly.NewRecoveryEvidenceRepository(source.sourcePool, func() (recovery.RecoveryEncryptionKey, error) { return key, nil })
	if err != nil {
		t.Fatal(err)
	}
	started := time.Now().UTC()
	exports, err := incidentportabilityassembly.NewRecoveryStorage(cfg.Roots.ExportOutputs.Path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(exports.Close)
	resources.ExportOutputs = exports
	fence, err := runner.AdmitRestore(ctx, admission, resources, backup, identity, func(work context.Context) error {
		return repository.AppendAdmission(work, application.RecoveryAdmissionRecord{OperationID: identity.OperationID, Operation: operation, StartedAt: started, BackupSetID: &identity.BackupSetID, ArtifactKinds: []string{}})
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := fence.Close(); err != nil {
			t.Error(err)
		}
	})
	return processRestoreAdmission{fence, admission, identity, repository, started}
}

func (admitted processRestoreAdmission) Complete(t testing.TB, result recovery.RestoreResult) {
	t.Helper()
	point := result.BackupSet.ConsistencyPointAt
	kind := "restore_operation"
	if admitted.Identity.Operation == string(application.OperationRestoreVerifyLatest) {
		kind = "restore_verification"
	}
	record := application.RecoveryCompletionRecord{OperationID: admitted.Identity.OperationID, Operation: application.Operation(admitted.Identity.Operation), StartedAt: admitted.StartedAt, CompletedAt: time.Now().UTC(), Result: application.ResultSucceeded, BackupSetID: &admitted.Identity.BackupSetID, ConsistencyPointAt: &point, ArtifactCounts: []application.ArtifactCount{{Kind: kind, Count: 1}}, RestoreGeneration: result.RestoreGeneration, TargetBindings: &admitted.Identity.BindingDigests}
	if err := admitted.Repository.AppendCompletion(t.Context(), record); err != nil {
		t.Fatal(err)
	}
	body, err := application.EncodeRecoveryCompletion(record)
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(body)
	if err := admitted.Fence.Complete(t.Context(), recovery.TargetCompletionProof{Identity: admitted.Identity, EvidenceSHA256: hex.EncodeToString(digest[:])}); err != nil {
		t.Fatal(err)
	}
}
