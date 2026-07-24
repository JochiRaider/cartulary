package recovery_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/JochiRaider/cartulary/internal/modules/evidence/blobref"
	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	"github.com/JochiRaider/cartulary/internal/platform/config"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
	"github.com/JochiRaider/cartulary/internal/testutil/s3test"
)

const migrationPreservationIdentity = "local_os_execution"

func TestSupportSeaweedFSMigrationPreservationPassEvidence(t *testing.T) {
	t.Parallel()
	fixture := newMigrationPreservationFixture(t, "pass")
	beforeRefs := fixture.evidenceRefs(t)
	result := runMigrationPreservation(t, fixture, uuid.MustParse("00000000-0000-0000-0000-000000130111"), migrationPreservationArtifactsDir(t, "pass"))
	if result.BlockingFailure || result.Run.CurrentState != recovery.ObjectStoreMigrationStateCutoverReady {
		t.Fatalf("migration pass did not reach cutover readiness: blocking=%v state=%s", result.BlockingFailure, result.Run.CurrentState)
	}
	requireMigrationPreservationArtifacts(t, result, true)
	if afterRefs := fixture.evidenceRefs(t); !reflect.DeepEqual(afterRefs, beforeRefs) {
		t.Fatalf("migration mutated database storage_ref values: before=%#v after=%#v", beforeRefs, afterRefs)
	}
	for _, blob := range fixture.Blobs {
		reader, _, err := fixture.TargetStore.ReadObject(fixture.Context, blob.StorageKey, objectstore.ReadOptions{})
		if err != nil {
			t.Fatalf("read migrated object %s: %v", blob.StorageKey, err)
		}
		body, err := ioReadAllAndClose(reader)
		if err != nil {
			t.Fatalf("read migrated object body %s: %v", blob.StorageKey, err)
		}
		if !bytes.Equal(body, blob.Body) {
			t.Fatalf("migrated object bytes changed for %s: got %q want %q", blob.StorageKey, body, blob.Body)
		}
	}
}

func TestSupportSeaweedFSMigrationPreservationMismatchEvidence(t *testing.T) {
	t.Parallel()
	fixture := newMigrationPreservationFixture(t, "mismatch")
	if err := fixture.TargetStore.PutObject(fixture.Context, fixture.Blobs[0].StorageKey, bytes.NewReader([]byte("target-side mismatch")), int64(len("target-side mismatch")), "application/octet-stream"); err != nil {
		t.Fatalf("seed target mismatch: %v", err)
	}
	result := runMigrationPreservation(t, fixture, uuid.MustParse("00000000-0000-0000-0000-000000130112"), migrationPreservationArtifactsDir(t, "mismatch"))
	if !result.BlockingFailure || result.Run.CurrentState != recovery.ObjectStoreMigrationStateFailed {
		t.Fatalf("migration mismatch did not fail closed: blocking=%v state=%s", result.BlockingFailure, result.Run.CurrentState)
	}
	requireMigrationPreservationArtifacts(t, result, false)
}

type migrationPreservationBlob struct {
	ObjectBlobID uuid.UUID
	RecordID     uuid.UUID
	StorageKey   string
	StorageRef   string
	Body         []byte
}

type migrationPreservationFixture struct {
	Context       context.Context
	Pool          *pgxpool.Pool
	DSN           string
	SourceStore   objectstore.Store
	TargetStore   objectstore.Store
	BackupStorage recovery.BackupStorage
	BackupSet     recovery.BackupSet
	SourceConfig  config.Config
	TargetConfig  config.Config
	Environment   map[string]string
	Bucket        string
	AsOf          time.Time
	Blobs         []migrationPreservationBlob
}

func newMigrationPreservationFixture(t testing.TB, name string) migrationPreservationFixture {
	t.Helper()
	ctx := context.Background()
	sourceHarness, err := s3test.StartOwnedWithLabels(ctx, map[string]string{"cartulary.fixture": "migration-preservation-source-" + name})
	if err != nil {
		t.Fatalf("start source S3 fixture: %v", err)
	}
	t.Cleanup(func() { _ = sourceHarness.Close(context.Background()) })
	targetHarness, err := s3test.StartOwnedWithLabels(ctx, map[string]string{"cartulary.fixture": "migration-preservation-target-" + name})
	if err != nil {
		t.Fatalf("start target S3 fixture: %v", err)
	}
	t.Cleanup(func() { _ = targetHarness.Close(context.Background()) })

	database := pgtest.Start(t).PrepareIsolatedDatabaseT(t, "seaweedfs-migration-preservation-"+name)
	pool, err := pgxpool.New(ctx, database.DSN)
	if err != nil {
		t.Fatalf("open migration source pool: %v", err)
	}
	t.Cleanup(pool.Close)
	actorID := seedMigrationPreservationUser(t, database.DSN)
	blobs := []migrationPreservationBlob{
		seedMigrationPreservationBlob(t, database.DSN, actorID, []byte("object-store migration object")),
		seedMigrationPreservationBlob(t, database.DSN, actorID, []byte{}),
	}

	bucket := "object-store-migration-" + name
	if err := sourceHarness.CreateBucket(ctx, bucket); err != nil {
		t.Fatalf("create source bucket: %v", err)
	}
	if err := targetHarness.CreateBucket(ctx, bucket); err != nil {
		t.Fatalf("create target bucket: %v", err)
	}
	for _, blob := range blobs {
		if _, err := sourceHarness.RoundTrip(ctx, bucket, blob.StorageKey, blob.Body); err != nil {
			t.Fatalf("seed source object %s: %v", blob.StorageKey, err)
		}
	}

	sourceConfig := config.Config{Roots: config.RootBindings{ObjectStorage: config.RootBinding{BindingKind: "managed_service", ServiceRef: "migration-source"}}}
	targetConfig := config.Config{Roots: config.RootBindings{ObjectStorage: config.RootBinding{BindingKind: "managed_service", ServiceRef: "migration-target"}}}
	environment := mergeMigrationPreservationEnv(sourceHarness.EnvForServiceRef("migration-source", bucket), targetHarness.EnvForServiceRef("migration-target", bucket))
	sourceStore, err := objectstore.SetupWithEnv(ctx, sourceConfig, environment)
	if err != nil {
		t.Fatalf("open source object store: %v", err)
	}
	t.Cleanup(func() { _ = sourceStore.Close() })
	targetStore, err := objectstore.SetupWithEnv(ctx, targetConfig, environment)
	if err != nil {
		t.Fatalf("open target object store: %v", err)
	}
	t.Cleanup(func() { _ = targetStore.Close() })

	asOf := time.Date(2026, 6, 4, 12, 0, 0, 0, time.UTC)
	backupSetID := map[string]uuid.UUID{
		"pass":     uuid.MustParse("00000000-0000-0000-0000-000000130101"),
		"mismatch": uuid.MustParse("00000000-0000-0000-0000-000000130102"),
	}[name]
	backupStorage := newEncryptedBackupStorage(t, t.TempDir())
	postgresArtifact, err := recovery.CapturePostgresSnapshotArtifact(ctx, pool)
	if err != nil {
		t.Fatalf("capture migration postgres artifact: %v", err)
	}
	blobIndex := map[string]uuid.UUID{}
	for _, blob := range blobs {
		blobIndex[blob.StorageKey] = blob.ObjectBlobID
	}
	objectArtifacts, err := recovery.CaptureSeaweedFSS3ObjectStoreBackupArtifacts(ctx, sourceStore, recovery.ObjectStoreBackupCaptureParams{
		BackupSetID:               backupSetID,
		ConsistencyPointAt:        asOf.Add(-2 * time.Minute),
		Bucket:                    bucket,
		BlobObjectIDsByStorageRef: blobIndex,
	})
	if err != nil {
		t.Fatalf("capture migration object-store artifacts: %v", err)
	}
	backupSet, err := recovery.NewCaptureService(recovery.NewStore(pool), backupStorage, testExtensionBackupCatalog(t)).CaptureBackupSet(ctx, captureParams(recovery.CaptureBackupSetParams{
		BackupSetID:                       backupSetID,
		ConsistencyPointAt:                asOf.Add(-2 * time.Minute),
		CreatedAt:                         asOf.Add(-3 * time.Minute),
		RetainedUntil:                     asOf.Add(31 * 24 * time.Hour),
		PostgresArtifact:                  recovery.BackupArtifact{Body: postgresArtifact, ContentType: "application/json"},
		ObjectStoreArtifact:               recovery.BackupArtifact{Body: objectArtifacts.SnapshotBody, ContentType: "application/json"},
		ObjectStoreBackupManifestArtifact: recovery.BackupArtifact{Body: objectArtifacts.ManifestBody, ContentType: "application/json"},
		ObjectStoreBackupSummaryArtifact:  recovery.BackupArtifact{Body: objectArtifacts.SummaryBody, ContentType: "application/json"},
	}))
	if err != nil {
		t.Fatalf("capture migration backup set: %v", err)
	}

	return migrationPreservationFixture{
		Context: ctx, Pool: pool, DSN: database.DSN, SourceStore: sourceStore, TargetStore: targetStore,
		BackupStorage: backupStorage, BackupSet: backupSet, SourceConfig: sourceConfig, TargetConfig: targetConfig,
		Environment: environment, Bucket: bucket, AsOf: asOf, Blobs: blobs,
	}
}

type migrationPreservationResult struct {
	Run             recovery.ObjectStoreMigrationRun
	CopyLedger      recovery.ObjectStoreMigrationCopyLedger
	Validation      recovery.ObjectStoreMigrationValidation
	BlockingFailure bool
	Artifacts       map[string]string
}

func runMigrationPreservation(t testing.TB, fixture migrationPreservationFixture, runID uuid.UUID, artifactsDir string) migrationPreservationResult {
	t.Helper()
	sourceSettings, err := objectstore.ResolveSettings(fixture.SourceConfig, fixture.Environment)
	if err != nil {
		t.Fatalf("resolve source settings: %v", err)
	}
	targetSettings, err := objectstore.ResolveSettings(fixture.TargetConfig, fixture.Environment)
	if err != nil {
		t.Fatalf("resolve target settings: %v", err)
	}
	if sourceSettings.Endpoint == targetSettings.Endpoint {
		t.Fatal("migration source and target bindings must differ")
	}

	proof := recovery.ObjectStoreMigrationWriteQuiescenceProof{
		SchemaID: recovery.ObjectStoreMigrationProofSchemaID, ProofKind: "process_stopped",
		CheckedAt: fixture.AsOf.Add(-time.Minute), ProcessState: "absent", HTTPListenerClosed: true, WebSocketListenerClosed: true,
	}
	if err := recovery.ValidateObjectStoreMigrationWriteQuiescenceProof(proof); err != nil {
		t.Fatalf("validate write-quiescence proof: %v", err)
	}
	backupRef, err := recovery.LoadObjectStoreMigrationBackupRefs(fixture.Context, fixture.BackupStorage, fixture.BackupSet)
	if err != nil {
		t.Fatalf("load migration backup refs: %v", err)
	}
	run, err := recovery.NewObjectStoreMigrationRun(runID, fixture.AsOf, migrationPreservationIdentity, sourceSettings.Endpoint, targetSettings.Endpoint, fixture.Bucket, fixture.Bucket)
	if err != nil {
		t.Fatalf("create migration run: %v", err)
	}
	applyMigrationEvent(t, &run, recovery.ObjectStoreMigrationEventPreflightPassed, fixture.AsOf, nil)
	applyMigrationEvent(t, &run, recovery.ObjectStoreMigrationEventWriteQuiescenceVerified, proof.CheckedAt, map[string]string{"proof_kind": proof.ProofKind})
	run.BackupRefs = append(run.BackupRefs, backupRef)
	applyMigrationEvent(t, &run, recovery.ObjectStoreMigrationEventBackupCaptured, fixture.BackupSet.ConsistencyPointAt, map[string]string{"backup_set_id": fixture.BackupSet.BackupSetID.String()})

	artifacts := map[string]string{}
	probe, probeBody, err := recovery.ProbeObjectStoreMigrationTarget(fixture.Context, runID, fixture.Bucket, fixture.TargetStore, fixture.AsOf.Add(time.Millisecond))
	if err != nil {
		t.Fatalf("probe migration target: %v", err)
	}
	probePath, probeRef := writeMigrationPreservationArtifact(t, artifactsDir, "target-probe.json", probeBody)
	artifacts["target-probe.json"] = probePath
	run.ProbeRef = &probeRef
	applyMigrationEvent(t, &run, recovery.ObjectStoreMigrationEventTargetPrepared, probe.CompletedAt, nil)

	_, rollbackBody, err := recovery.BuildObjectStoreMigrationRollbackEvidence(runID, fixture.AsOf.Add(2*time.Millisecond))
	if err != nil {
		t.Fatalf("build rollback evidence: %v", err)
	}
	rollbackPath, rollbackRef := writeMigrationPreservationArtifact(t, artifactsDir, "rollback-evidence.json", rollbackBody)
	artifacts["rollback-evidence.json"] = rollbackPath
	run.RollbackRef = &rollbackRef

	objects, err := recovery.ListObjectStoreMigrationBlobs(fixture.Context, fixture.Pool)
	if err != nil {
		t.Fatalf("list migration blobs: %v", err)
	}
	applyMigrationEvent(t, &run, recovery.ObjectStoreMigrationEventCopyStarted, fixture.AsOf.Add(3*time.Millisecond), nil)
	copyLedger, copyBody, err := recovery.CopyObjectStoreMigrationObjects(fixture.Context, recovery.ObjectStoreMigrationCopyParams{
		RunID: runID, SourceBackend: recovery.ObjectStoreBackendMinIOS3, TargetBackend: recovery.ObjectStoreBackendSeaweedFSS3,
		SourceBucket: fixture.Bucket, TargetBucket: fixture.Bucket, SourceStore: fixture.SourceStore, TargetStore: fixture.TargetStore, Objects: objects,
	})
	if err != nil {
		t.Fatalf("copy migration objects: %v", err)
	}
	copyPath, copyRef := writeMigrationPreservationArtifact(t, artifactsDir, "copy-ledger.json", copyBody)
	artifacts["copy-ledger.json"] = copyPath
	run.CopyLedgerRef = &copyRef

	validation, validationBody, err := recovery.ValidateObjectStoreMigration(fixture.Context, recovery.ObjectStoreMigrationValidationParams{
		RunID: runID, StartedAt: fixture.AsOf.Add(4 * time.Millisecond), CompletedAt: fixture.AsOf.Add(5 * time.Millisecond),
		SourceBackend: recovery.ObjectStoreBackendMinIOS3, TargetBackend: recovery.ObjectStoreBackendSeaweedFSS3,
		SourceBucket: fixture.Bucket, TargetBucket: fixture.Bucket, SourceStore: fixture.SourceStore, TargetStore: fixture.TargetStore, Objects: objects,
	})
	if err != nil {
		t.Fatalf("validate migration objects: %v", err)
	}
	validationPath, validationRef := writeMigrationPreservationArtifact(t, artifactsDir, "validation.json", validationBody)
	artifacts["validation.json"] = validationPath
	run.ValidationRef = &validationRef

	blocking := copyLedger.Result != "pass" || validation.Result != "pass"
	if blocking {
		applyMigrationEvent(t, &run, recovery.ObjectStoreMigrationEventBlockingFailure, fixture.AsOf.Add(6*time.Millisecond), nil)
	} else {
		applyMigrationEvent(t, &run, recovery.ObjectStoreMigrationEventCopyCompleted, fixture.AsOf.Add(5*time.Millisecond), nil)
		applyMigrationEvent(t, &run, recovery.ObjectStoreMigrationEventValidationStarted, fixture.AsOf.Add(6*time.Millisecond), nil)
		applyMigrationEvent(t, &run, recovery.ObjectStoreMigrationEventValidationPassed, fixture.AsOf.Add(7*time.Millisecond), nil)
	}
	runBody, err := recovery.EncodeObjectStoreMigrationRun(run)
	if err != nil {
		t.Fatalf("encode migration run: %v", err)
	}
	runPath, _ := writeMigrationPreservationArtifact(t, artifactsDir, "migration-run.json", runBody)
	artifacts["migration-run.json"] = runPath
	return migrationPreservationResult{Run: run, CopyLedger: copyLedger, Validation: validation, BlockingFailure: blocking, Artifacts: artifacts}
}

func applyMigrationEvent(t testing.TB, run *recovery.ObjectStoreMigrationRun, event recovery.ObjectStoreMigrationEventName, at time.Time, detail map[string]string) {
	t.Helper()
	if err := recovery.ApplyObjectStoreMigrationEvent(run, event, at, detail); err != nil {
		t.Fatalf("apply migration event %s: %v", event, err)
	}
}

func requireMigrationPreservationArtifacts(t testing.TB, result migrationPreservationResult, wantPass bool) {
	t.Helper()
	for _, name := range []string{"migration-run.json", "copy-ledger.json", "validation.json", "target-probe.json", "rollback-evidence.json"} {
		if _, err := os.Stat(result.Artifacts[name]); err != nil {
			t.Fatalf("missing migration artifact %s: %v", name, err)
		}
	}
	if wantPass {
		if result.CopyLedger.Result != "pass" || result.Validation.Result != "pass" || result.Run.CurrentState != recovery.ObjectStoreMigrationStateCutoverReady {
			t.Fatalf("pass artifacts are not cutover-ready: %#v", result)
		}
		return
	}
	if result.CopyLedger.Result != "fail" || result.Validation.Result != "fail" || result.Run.CurrentState != recovery.ObjectStoreMigrationStateFailed {
		t.Fatalf("mismatch artifacts are not fail-closed: %#v", result)
	}
	foundMismatch := false
	for _, item := range result.CopyLedger.Items {
		foundMismatch = foundMismatch || item.Status == recovery.ObjectStoreMigrationCopyTargetMismatch
	}
	if !foundMismatch {
		t.Fatalf("mismatch ledger lacks target_mismatch: %#v", result.CopyLedger.Items)
	}
}

func migrationPreservationArtifactsDir(t testing.TB, scenario string) string {
	t.Helper()
	root := strings.TrimSpace(os.Getenv("CARTULARY_TEST_RESULTS_DIR"))
	runID := strings.TrimSpace(os.Getenv("CARTULARY_TEST_RUN_ID"))
	if root == "" || runID == "" {
		root = t.TempDir()
		runID = "local"
	}
	dir := filepath.Join(root, runID, "seaweedfs-migration-preservation", "object-store-migration", scenario)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatalf("create migration preservation artifact directory: %v", err)
	}
	return dir
}

func writeMigrationPreservationArtifact(t testing.TB, dir string, name string, body []byte) (string, recovery.ObjectStoreMigrationArtifactRef) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, body, 0o600); err != nil {
		t.Fatalf("write migration artifact %s: %v", name, err)
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		t.Fatalf("resolve migration artifact path: %v", err)
	}
	return abs, recovery.ArtifactRefForBody(abs, body, "application/json")
}

func seedMigrationPreservationUser(t testing.TB, dsn string) uuid.UUID {
	t.Helper()
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("open user seed database: %v", err)
	}
	defer db.Close()
	var userID uuid.UUID
	email := "migration-preservation-" + uuid.NewString() + "@example.test"
	if err := db.QueryRow(`
INSERT INTO users (email, display_name, password_hash, mfa_required, is_active, is_deployment_admin)
VALUES ($1, $2, 'not-used-by-migration-preservation', false, true, true)
RETURNING id
`, email, email).Scan(&userID); err != nil {
		t.Fatalf("seed migration user: %v", err)
	}
	return userID
}

func seedMigrationPreservationBlob(t testing.TB, dsn string, actorID uuid.UUID, body []byte) migrationPreservationBlob {
	t.Helper()
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("open blob seed database: %v", err)
	}
	defer db.Close()
	now := time.Now().UTC()
	incidentID, recordID, blobID := uuid.New(), uuid.New(), uuid.New()
	storageKey := blobref.MustObjectBlobStorageKey(incidentID, blobID)
	sha := fmt.Sprintf("%x", sha256.Sum256(body))
	if _, err := db.Exec(`INSERT INTO incidents (id, incident_key, incident_key_canonical, title, status, created_by_user_id, updated_by_user_id, created_at, updated_at) VALUES ($1, $2, $2, 'Migration preservation fixture', 'active', $3, $3, $4, $4)`, incidentID, "object-store-migration-"+blobID.String(), actorID, now); err != nil {
		t.Fatalf("insert migration incident: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO records (record_id, incident_id, record_type, created_by_user_id, created_at, updated_by_user_id, updated_at, row_version) VALUES ($1, $2, 'evidence', $3, $4, $3, $4, 1)`, recordID, incidentID, actorID, now); err != nil {
		t.Fatalf("insert migration record: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO object_blobs (object_blob_id, incident_id, created_by_user_id, storage_key, upload_state, byte_size, observed_size, observed_content_type, observed_sha256_hex, target_expires_at, pending_expires_at, finalized_at, created_at, updated_at) VALUES ($1, $2, $3, $4, 'available', $5, $5, 'application/octet-stream', $6, $7, $7, $8, $8, $8)`, blobID, incidentID, actorID, storageKey, int64(len(body)), sha, now.Add(time.Hour), now); err != nil {
		t.Fatalf("insert migration object blob: %v", err)
	}
	storageRef := "object://" + blobID.String()
	if _, err := db.Exec(`INSERT INTO evidence (record_id, incident_id, title, lifecycle_state, upload_state, storage_ref, blob_hash, object_blob_id, requested_at, received_at, created_at, updated_at) VALUES ($1, $2, 'Migration preservation evidence', 'available', 'available', $3, $4, $5, $6, $6, $6, $6)`, recordID, incidentID, storageRef, sha, blobID, now); err != nil {
		t.Fatalf("insert migration evidence: %v", err)
	}
	return migrationPreservationBlob{ObjectBlobID: blobID, RecordID: recordID, StorageKey: storageKey, StorageRef: storageRef, Body: append([]byte(nil), body...)}
}

func (fixture migrationPreservationFixture) evidenceRefs(t testing.TB) map[string]string {
	t.Helper()
	rows, err := fixture.Pool.Query(fixture.Context, `SELECT record_id::text, COALESCE(storage_ref, '') FROM evidence ORDER BY record_id::text`)
	if err != nil {
		t.Fatalf("query evidence refs: %v", err)
	}
	defer rows.Close()
	refs := map[string]string{}
	for rows.Next() {
		var recordID, storageRef string
		if err := rows.Scan(&recordID, &storageRef); err != nil {
			t.Fatalf("scan evidence ref: %v", err)
		}
		refs[recordID] = storageRef
	}
	return refs
}

func mergeMigrationPreservationEnv(parts ...map[string]string) map[string]string {
	merged := map[string]string{}
	for _, part := range parts {
		for key, value := range part {
			merged[key] = value
		}
	}
	return merged
}
