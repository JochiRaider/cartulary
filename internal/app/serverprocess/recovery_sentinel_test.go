package serverprocess

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/app/extensionassembly"
	"github.com/JochiRaider/cartulary/internal/app/projectionassembly"
	"github.com/JochiRaider/cartulary/internal/app/recoveryassembly"
	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/evidence/recoveryprovider"
	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	recoverytestsupport "github.com/JochiRaider/cartulary/internal/modules/recovery/testsupport"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/configtest"
	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
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

	recoverytestsupport.RestoreLatest(t, ctx,
		recovery.NewRestoreRunner(fixture.SourceStore, fixture.BackupStorage, recoveryExtensionCatalog(t)),
		target.Postgres,
		recovery.RestoreTarget{
			Postgres:        target.Postgres,
			ObjectStore:     target.ObjectStore,
			EvidenceObjects: recoveryprovider.New(target.Postgres),
			Projections:     projectionRuntime.RecoveryPorts().Rebuilder,
		}, fixture.AsOf, recoverytestsupport.RestoreExpectation{
			BackupSetID:             fixture.LatestBackupSetID,
			ConsistencyPointAt:      fixture.ConsistencyPointAt,
			AuthoritativeRowsSHA256: fixture.AuthoritativeRowsSHA256,
			ChangeSetsSHA256:        fixture.ChangeSetsSHA256,
			BlobHashesSHA256:        fixture.BlobHashesSHA256,
			ChangeSetRowCount:       fixture.ChangeSetRowCount,
			BlobCount:               1,
			EvidenceRecordID:        fixture.EvidenceRecordID,
			EvidenceBlob:            fixture.EvidenceBlob,
			BlobSHA256:              fixture.BlobSHA256,
		})

	basis := recovery.RestoreVerificationBasis{
		MechanismID:                "backup_restore.process.restore.v1",
		DatabaseBindingSHA256:      recovery.SHA256String("backup_restore-i-10-02-database"),
		ObjectStoreBindingSHA256:   recovery.SHA256String("backup_restore-i-10-02-objects"),
		BackupStorageBindingSHA256: recovery.SHA256String("backup_restore-i-10-02-backups"),
		RecoveryStateCatalogSHA256: recovery.SHA256String("backup_restore-i-10-02-catalog"),
		CodecRegistrySHA256:        recovery.SHA256String("backup_restore-i-10-02-codecs"),
	}
	verificationTarget := prepareRestoreTarget(t, "backup_restore-i-10-02-verification-target")
	verificationRebuilder, verificationQuery, err := projectionassembly.NewRecoveryServices(verificationTarget.Postgres)
	if err != nil {
		t.Fatalf("compose restore projection services: %v", err)
	}
	verification, err := recovery.NewRestoreVerificationService(
		fixture.SourceStore,
		recovery.NewRestoreRunner(fixture.SourceStore, fixture.BackupStorage, recoveryExtensionCatalog(t)),
	).VerifyLatestSuccessfulRetained(ctx, recovery.RestoreVerificationTarget{
		RestoreTarget: recovery.RestoreTarget{
			Postgres:        verificationTarget.Postgres,
			ObjectStore:     verificationTarget.ObjectStore,
			EvidenceObjects: recoveryprovider.New(verificationTarget.Postgres),
			Projections:     verificationRebuilder,
		},
		Probe: recovery.RestoreVerificationWorkbookProbe{Executor: verificationQuery},
	}, fixture.AsOf, basis)
	if err != nil {
		t.Fatalf("verify restored process fixture: %v", err)
	}
	if verification.BackupSet.VerificationState != recovery.VerificationVerified || verification.Artifact.Result != "pass" {
		t.Fatalf("restore verification did not pass: %#v", verification)
	}
	verificationPath := recoverytestsupport.RequireVerificationArtifact(t, fixture.BackupStorage, verification, recoveryEvidenceLocation(t, "backup-restore"), recoverytestsupport.VerificationExpectation{
		BackupSetID:        fixture.LatestBackupSetID,
		ConsistencyPointAt: fixture.ConsistencyPointAt,
		IncidentID:         fixture.IncidentID,
		ObjectCount:        1,
		RegistrationID:     "timeline.base_restore_probe.v1",
		ViewSchemaID:       "cartulary.view.timeline.v2",
	})
	t.Logf("backup_restore_verification_artifact=%s", verificationPath)

	server := processtest.StartServer(t, processtest.ServerOptions{Env: target.Env})
	defer server.Stop(t)
	server.WaitForReady(t)

	login := loginLocalUserWithSecondFactor(t, server, authenticationBootstrapAdminEmail, authenticationBootstrapAdminPassword, flowtest.GenerateTOTPCode(t, fixture.AdminTOTPSecret))
	getIncident := doJSON(t, server, http.MethodGet, "/api/v1/incidents/"+fixture.IncidentID, nil, withCookies(login.SessionCookie))
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
	SourceStore             *recovery.Store
	BackupStorage           recovery.BackupStorage
	AsOf                    time.Time
	ConsistencyPointAt      time.Time
	LatestBackupSetID       uuid.UUID
	IncidentID              string
	TimelineRecordID        string
	TimelineRowVersion      int
	EvidenceRecordID        string
	BlobSHA256              string
	AuthoritativeRowsSHA256 string
	ChangeSetsSHA256        string
	BlobHashesSHA256        string
	ChangeSetRowCount       int
	EvidenceBlob            recoverytestsupport.EvidenceBlobConsistency
	AdminTOTPSecret         string
}

func captureRestoreSource(t testing.TB, prefix string) sourceBackupFixture {
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
	incident := createIncident(t, server, adminLogin.SessionCookie, adminLogin.CSRFCookie, map[string]any{
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
	}, withCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie), withHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value))
	blobData := httptestx.RequireSuccessEnvelope(t, blobCreate, http.StatusCreated)["data"].(map[string]any)
	uploadTarget := blobData["upload_target"].(map[string]any)
	objectBlobID := blobData["object_blob_id"].(string)
	putObject(t, server.BaseURL, uploadTarget["href"].(string), payload, "text/plain")

	attach := doJSON(t, server, http.MethodPost, "/api/v1/evidence-records/"+evidenceRecordID+"/attach-blob", map[string]any{
		"object_blob_id":   objectBlobID,
		"base_row_version": 1,
		"client_txn_id":    "txn-" + prefix + "-attach",
	}, withCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie), withHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value))
	httptestx.RequireSuccessEnvelope(t, attach, http.StatusOK)

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
	olderBackupSetID := uuid.MustParse("00000000-0000-0000-0000-000000100201")
	latestBackupSetID := uuid.MustParse("00000000-0000-0000-0000-000000100202")
	asOf := time.Now().UTC().Truncate(time.Second)
	olderConsistencyPointAt := asOf.Add(-9 * time.Minute)
	consistencyPointAt := asOf.Add(-time.Minute)
	backupRoot := t.TempDir()
	backupStorage := encryptedBackupStorage(t, backupRoot)
	sourceRecoveryStore := recovery.NewStore(sourcePool)
	captured := recoverytestsupport.CaptureSourceBackup(t, ctx, recoverytestsupport.CaptureInput{
		Prefix:                  prefix,
		AsOf:                    asOf,
		OlderBackupSetID:        olderBackupSetID,
		OlderConsistencyPointAt: olderConsistencyPointAt,
		BackupSetID:             latestBackupSetID,
		ConsistencyPointAt:      consistencyPointAt,
		Postgres:                sourcePool,
		ObjectStore:             sourceObjectStore,
		ObjectStoreBucket:       bucket,
		Store:                   sourceRecoveryStore,
		EvidenceObjects:         recoveryprovider.New(sourcePool),
		BackupStorage:           backupStorage,
		BackupStorageRoot:       backupRoot,
		ExtensionCatalog:        recoveryExtensionCatalog(t),
		EvidenceLocation:        recoveryEvidenceLocation(t, "backup-restore"),
		IncidentID:              incidentID,
		TimelineRecordID:        timelineRecordID,
		TimelineRowVersion:      linkedTimelineRowVersion,
		EvidenceRecordID:        evidenceRecordID,
		ObjectBlobID:            objectBlobID,
		BlobBody:                payload,
	})
	t.Logf("object_store_backup_manifest=%s", captured.ManifestEvidencePath)
	t.Logf("object_store_backup_summary=%s", captured.SummaryEvidencePath)

	return sourceBackupFixture{
		SourceStore:             sourceRecoveryStore,
		BackupStorage:           backupStorage,
		AsOf:                    captured.AsOf,
		ConsistencyPointAt:      captured.ConsistencyPointAt,
		LatestBackupSetID:       captured.BackupSetID,
		IncidentID:              captured.IncidentID,
		TimelineRecordID:        captured.TimelineRecordID,
		TimelineRowVersion:      captured.TimelineRowVersion,
		EvidenceRecordID:        captured.EvidenceRecordID,
		BlobSHA256:              captured.BlobSHA256,
		AuthoritativeRowsSHA256: captured.AuthoritativeRowsSHA256,
		ChangeSetsSHA256:        captured.ChangeSetsSHA256,
		BlobHashesSHA256:        captured.BlobHashesSHA256,
		ChangeSetRowCount:       captured.ChangeSetRowCount,
		EvidenceBlob:            captured.EvidenceBlob,
		AdminTOTPSecret:         adminSecret,
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

func prepareRestoreTarget(t testing.TB, prefix string) recoverytestsupport.TargetFixture {
	t.Helper()

	ctx := t.Context()
	postgresHarness, s3Harness := sharedProcessHarnesses(t)
	targetDB := postgresHarness.PrepareIsolatedDatabaseT(t, prefix)
	targetPool, err := pgxpool.New(ctx, targetDB.DSN)
	if err != nil {
		t.Fatalf("open target pgx pool: %v", err)
	}
	t.Cleanup(targetPool.Close)

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
	return target
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
