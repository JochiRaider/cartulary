package recovery_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/app/projectionassembly"
	"github.com/JochiRaider/cartulary/internal/app/recoveryassembly"
	"github.com/JochiRaider/cartulary/internal/app/referenceassembly"
	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/scenariotest"
	"github.com/JochiRaider/cartulary/internal/modules/networkflow"
	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/platform/recoverystate"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport/extensionstate"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/recoverytest"
	"github.com/JochiRaider/cartulary/internal/testutil/s3test"
)

func TestRealBackingStorageMetadataPersistsAndLatestLookup_Integration(t *testing.T) {
	runtimeHarness := appsupport.StartRuntime(t)
	objects := s3test.Start(t)
	harness := runtimeHarness.StartServer(t, appsupport.ServerOptions{Prefix: "backup_restore-i-10-01-metadata", TestRouteMode: httptestx.TestRouteModeDisabled})
	ctx := t.Context()
	recoveryDSN, err := harness.Database.DSNForPurpose(postgres.PurposeRecovery)
	if err != nil {
		t.Fatalf("resolve source Recovery DSN: %v", err)
	}
	recoveryAdmission, err := postgres.Setup(ctx, postgres.Settings{
		BindingKind:  "managed_service",
		DSN:          recoveryDSN,
		Purpose:      postgres.PurposeRecovery,
		ExpectedRole: "cartulary_recovery",
	})
	if err != nil {
		t.Fatalf("open admitted source Recovery pool: %v", err)
	}
	t.Cleanup(recoveryAdmission.Close)
	recoveryPool := recoveryAdmission.Pool()
	store := recovery.NewStore(recoveryPool)
	backupStorage, err := recoveryassembly.NewBackupStorage(
		harness.Server.Config.Roots.BackupStorage.BindingKind,
		harness.Server.Config.Roots.BackupStorage.Path,
		map[string]string{
			recovery.RecoveryMasterKeyEnv: RecoveryMasterKey,
		},
	)
	if err != nil {
		t.Fatalf("create backup storage from runtime config: %v", err)
	}

	adminLogin, adminUserID := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, harness.Server.HTTP.URL)
	incident := appsupport.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-backup_restore-i-10-01-incident",
		"incident_key":  "backup_restore-i-10-01",
		"title":         "Recovery and coordination recovery-metadata Backup Metadata",
	})
	incidentID := uuid.MustParse(incident["incident_id"].(string))
	// These retained declarations are deliberately incompatible with current
	// profile semantics. Recovery must preserve their exact logical entries;
	// only a later claim may reject them. The large JSON integer must never pass
	// through a floating-point database decoding path.
	for _, id := range []string{"nfgv_ffffffffffffffffffffffffffffffff", "nfgv_00000000000000000000000000000001"} {
		if _, err := recoveryPool.Exec(ctx, `INSERT INTO network_flow_graph_views
		(graph_view_id,incident_id,display_name,normalized_display_name,declaration_state,
		 refresh_intent_json,semantic_query_json,semantic_query_sha256,desired_source_snapshot_id,
		 graph_view_version,materialization_generation,created_by_user_id,created_at,updated_at,retired_at)
		VALUES ($1,$2,'Retained Ω','retained ω','retired','{}',
		 '{"schema_id":"incompatible.profile.semantic.v1","counter":9007199254740993}',
		 repeat('a',64),'retained-source',1,1,$3,now(),now(),now())`, id, incidentID, adminUserID); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := recoveryPool.Exec(ctx, `INSERT INTO extension_state_metadata
	(profile_id,migration_lineage_id,state_version,metadata_version,created_at,updated_at)
VALUES ('network_flow_activity','network_flow_activity.state_v1',7,1,now(),now())
	ON CONFLICT (profile_id) DO NOTHING`); err != nil {
		t.Fatal(err)
	}

	historyUser := flowtest.SeedActor(t, harness.DB, flowtest.ActorAttributes{Email: "recovery-history@example.test", DisplayName: "Removed member"})
	scenariotest.CreateMembership(t, harness.Server, adminLogin, incidentID.String(), map[string]any{"client_txn_id": "recovery-history-create", "user_id": historyUser.ID.String(), "role": "viewer"})
	removed := httptestx.DoJSON(t, http.DefaultClient, http.MethodDelete, harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/memberships/"+historyUser.ID.String(), map[string]any{"base_membership_version": 1}, httptestx.WithCookies(adminLogin.Session, adminLogin.CSRF), httptestx.WithHeader("X-CSRF-Token", adminLogin.CSRF.Value))
	if removed.StatusCode != http.StatusNoContent {
		t.Fatalf("remove recovery history membership: status=%d", removed.StatusCode)
	}
	_ = removed.Body.Close()
	objectKey := "backup_restore/i-10-01/" + incident["incident_id"].(string) + "/proof.txt"
	objectPayload := []byte("backup_restore backup proof")
	sourceBucket := "backup-restore-i-10-01-source-" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if err := objects.CreateBucket(ctx, sourceBucket); err != nil {
		t.Fatalf("create source SeaweedFS bucket: %v", err)
	}
	sourceObjectStore := S3StoreForBucket(t, harness, objects, sourceBucket)
	if err := sourceObjectStore.PutObject(ctx, objectKey, bytes.NewReader(objectPayload), int64(len(objectPayload)), "text/plain"); err != nil {
		t.Fatalf("write object-store proof before capture: %v", err)
	}
	objectSHA := SHA256Hex(objectPayload)
	objectBlobID := uuid.MustParse("00000000-0000-0000-0000-000000101003")
	if _, err := harness.Pool.Exec(ctx, `
INSERT INTO object_blobs (
    object_blob_id, incident_id, created_by_user_id, storage_key, upload_state,
    byte_size, expected_sha256_hex, observed_size, observed_content_type, observed_sha256_hex,
    target_expires_at, pending_expires_at, finalized_at, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, 'available',
    $5, $6, $5, 'text/plain', $6,
    $7, $7, $8, $8, $8
)
`, objectBlobID, incidentID, adminUserID, objectKey, int64(len(objectPayload)), objectSHA, asTime(t, "2026-05-22T13:00:00Z"), asTime(t, "2026-05-22T12:00:00Z")); err != nil {
		t.Fatalf("insert source durable object blob row: %v", err)
	}
	stateCatalog, err := recoveryassembly.CurrentRecoveryStateCatalog()
	if err != nil {
		t.Fatal(err)
	}
	sourcePacks, err := referenceassembly.NewRecoveryStorage(harness.Server.Config.Roots.TemporaryWork.Path, harness.Server.Config.Roots.ReferencePackStorage.Path, reference_data.DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(sourcePacks.Close)
	inventories, err := recoveryassembly.CurrentVNextObjectInventoryCatalog(recoveryassembly.NewVNextObjectSource(sourceObjectStore), sourcePacks, nil)
	if err != nil {
		t.Fatal(err)
	}
	streaming, err := recovery.RequireStreamingBackupStorage(backupStorage)
	if err != nil {
		t.Fatal(err)
	}
	snapshotCuts := &bindingSnapshotCut{VNextSnapshotRepository: recoveryassembly.NewVNextSnapshotRepository(recoveryPool), mutate: func(ctx context.Context) error {
		_, err := recoveryPool.Exec(ctx, `UPDATE network_flow_graph_views SET display_name='Later source Ω' WHERE graph_view_id='nfgv_00000000000000000000000000000001'`)
		return err
	}}
	capture, err := recovery.NewVNextCaptureService(snapshotCuts, streaming, stateCatalog, inventories, testExtensionBackupCatalog(t), recoverytest.ProofScratch(t))
	if err != nil {
		t.Fatal(err)
	}
	asOf := time.Now().UTC()
	captureAt := func(id uuid.UUID, created, point time.Time) (recovery.VNextCapturedBackup, recovery.BackupSet) {
		t.Helper()
		captured, err := capture.Capture(ctx, recovery.VNextCaptureParams{BackupSetID: id, ConsistencyPointAt: point, CreatedAt: created, RetainedUntil: created.Add(31 * 24 * time.Hour)})
		if err != nil {
			t.Fatal(err)
		}
		published, err := store.PublishVNextCapturedBackup(ctx, captured)
		if err != nil {
			t.Fatal(err)
		}
		return captured, published
	}
	olderID := uuid.MustParse("00000000-0000-0000-0000-000000101001")
	captureAt(olderID, asOf.Add(-6*time.Hour), asOf.Add(-5*time.Hour))
	latestID := uuid.MustParse("00000000-0000-0000-0000-000000101002")
	latestCaptured, latestCreated := captureAt(latestID, asOf.Add(-2*time.Hour), asOf.Add(-time.Hour))
	if snapshotCuts.snapshots != 2 || snapshotCuts.mutations != 1 {
		t.Fatalf("capture snapshot boundaries: snapshots=%d mutations=%d", snapshotCuts.snapshots, snapshotCuts.mutations)
	}
	var changedSource int
	if err := recoveryPool.QueryRow(ctx, `SELECT count(*) FROM network_flow_graph_views WHERE display_name='Later source Ω'`).Scan(&changedSource); err != nil || changedSource != 1 {
		t.Fatalf("concurrent source write did not occur: %d %v", changedSource, err)
	}

	for _, proof := range []recovery.BackupArtifactStreamProof{latestCaptured.PostgresProof, latestCaptured.ObjectManifestProof, latestCaptured.IntegrityProof} {
		var body bytes.Buffer
		if err := streaming.ReadArtifactStream(ctx, proof, &body); err != nil {
			t.Fatal(err)
		}
		recovery.AssertCurrentArtifactSchema(t, body.Bytes())
		if proof.LogicalRef == latestCaptured.PostgresProof.LogicalRef {
			var snapshot struct {
				SchemaID string           `json:"schema_id"`
				Proofs   []map[string]any `json:"extension_binding_proofs"`
			}
			if err := json.Unmarshal(body.Bytes(), &snapshot); err != nil {
				t.Fatal(err)
			}
			bindings := testExtensionBackupCatalog(t).Bindings()
			if snapshot.SchemaID != "cartulary.postgres_snapshot_artifact.v3" || len(snapshot.Proofs) != len(bindings) {
				t.Fatalf("current capture must emit snapshot v3 with every binding proof: schema=%s proofs=%d bindings=%d", snapshot.SchemaID, len(snapshot.Proofs), len(bindings))
			}
			for i, binding := range bindings {
				proof := snapshot.Proofs[i]
				wantCount := float64(0)
				if binding.CurrentCodec.Postgres.Mapping().TableName == "network_flow_graph_views" {
					wantCount = 2
				}
				if proof["item_count"] != wantCount || (wantCount > 0 && proof["content_bytes"].(float64) <= 0) {
					t.Fatalf("logical binding count/bytes = %#v", proof)
				}

				if proof["schema_id"] != "cartulary.extension_postgres_binding_proof.v1" ||
					proof["profile_id"] != binding.ProfileID || proof["binding_id"] != binding.BindingID ||
					proof["physical_state_binding_sha256"] != binding.PhysicalStateBindingSHA256 ||
					proof["implementation_binding_sha256"] != binding.ImplementationBindingSHA256 ||
					proof["backup_codec_id"] != binding.CurrentCodec.CodecID || proof["backup_codec_sha256"] != binding.CurrentCodec.CodecSHA256 {
					t.Fatal("captured binding proof does not identify the admitted current binding")
				}
			}
		}
	}
	var objectManifestBody bytes.Buffer
	if err := streaming.ReadArtifactStream(ctx, latestCaptured.ObjectManifestProof, &objectManifestBody); err != nil {
		t.Fatal(err)
	}
	var objectManifest recovery.VNextObjectStoreBackupManifest
	if err := json.Unmarshal(objectManifestBody.Bytes(), &objectManifest); err != nil {
		t.Fatal(err)
	}
	foundBlob := false
	for _, item := range objectManifest.Objects {
		if item.OwnerID == "module.evidence" && item.StorageKey == objectKey {
			if item.PlaintextSHA256 != objectSHA || item.PlaintextBytes != int64(len(objectPayload)) {
				t.Fatalf("backup blob identity changed: %#v", item)
			}
			foundBlob = true
		}
	}
	if !foundBlob || len(objectManifest.Objects) <= 1 {
		t.Fatal("backup omitted evidence blob or Reference Pack assets")
	}

	reopenedStore := recovery.NewStore(recoveryPool)
	targetDB := runtimeHarness.PrepareIsolatedDatabase(t, "backup_restore-i-10-01-service-backed-target")
	targetAdmissionPool := openCurrentRestorePool(t, targetDB)
	targetPool := targetAdmissionPool.Pool()
	targetBucket := "backup-restore-i-10-01-target-" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if err := objects.CreateBucket(ctx, targetBucket); err != nil {
		t.Fatalf("create target SeaweedFS bucket: %v", err)
	}
	targetObjectStore := S3StoreForBucket(t, harness, objects, targetBucket)
	if objects, err := targetObjectStore.ListObjects(ctx, ""); err != nil {
		t.Fatalf("list fresh target SeaweedFS bucket: %v", err)
	} else if len(objects) != 0 {
		t.Fatalf("fresh target SeaweedFS bucket is not empty before restore: %#v", objects)
	}

	targetPacks, err := referenceassembly.NewRecoveryStorage(t.TempDir(), t.TempDir(), reference_data.DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(targetPacks.Close)

	algorithms, err := recovery.NewVNextRestoreAlgorithmCatalog(stateCatalog, recovery.RequiredVNextRestoreAlgorithmIDs(stateCatalog)...)
	if err != nil {
		t.Fatal(err)
	}
	structural, err := recovery.NewVNextRestoreService(streaming, stateCatalog, algorithms, testExtensionBackupCatalog(t))
	if err != nil {
		t.Fatal(err)
	}
	restoreSource := capturedBackupContractFixture{Store: reopenedStore, BackupStorage: backupStorage, BackupSet: latestCreated, sourcePool: recoveryPool, state: stateCatalog, AsOf: asOf}
	composeTarget := func(t testing.TB, pool postgres.AdmittedPool) recovery.RestoreTarget {
		t.Helper()
		projection, err := projectionassembly.Build(pool)
		if err != nil {
			t.Fatal(err)
		}
		return recovery.RestoreTarget{Postgres: pool, ObjectStore: targetObjectStore, ReferencePacks: targetPacks, ProofScratch: recoverytest.ProofScratch(t), Projections: projection.RecoveryPorts().Rebuilder, EvidenceObjects: evidence.NewRecoveryProvider(pool)}
	}
	for _, defect := range []string{"missing", "extra", "identity", "content"} {
		t.Run("binding readback rejects "+defect, func(t *testing.T) {
			failedDB := runtimeHarness.PrepareIsolatedDatabase(t, "binding-readback-"+defect)
			failedPool := openCurrentRestorePool(t, failedDB)
			runner := recovery.NewVersionedRestoreRunner(reopenedStore, backupStorage, testExtensionBackupCatalog(t), stateCatalog)
			admitted := admitCurrentRestore(t, restoreSource, runner, composeTarget(t, failedPool))
			trace := &observedBindingRestore{VNextRestoreTarget: recovery.AtomicRestoreTargetForTest(admitted.Fence), defect: defect}
			if err := structural.Restore(ctx, trace, latestCaptured.IntegrityProof); !errors.Is(err, recovery.ErrExtensionBindingInvalid) {
				t.Fatalf("corrupt logical binding admitted: %v", err)
			}
			var want []string
			for _, binding := range testExtensionBackupCatalog(t).Bindings() {
				want = append(want, "table:"+binding.CurrentCodec.Postgres.Mapping().TableName, "validate:"+binding.BindingID)
			}
			if !slices.Equal(trace.steps, want) || trace.objects != 0 {
				t.Fatalf("restore advanced before valid readback: steps=%v objects=%d", trace.steps, trace.objects)
			}
			var retained int
			if err := failedPool.QueryRow(ctx, `SELECT count(*) FROM network_flow_graph_views`).Scan(&retained); err != nil || retained != 0 {
				t.Fatalf("failed binding committed: %d %v", retained, err)
			}
		})
	}
	runner := recovery.NewVersionedRestoreRunner(reopenedStore, backupStorage, testExtensionBackupCatalog(t), stateCatalog)
	admitted := admitCurrentRestore(t, restoreSource, runner, composeTarget(t, targetAdmissionPool))
	serviceBackedRestore, err := runner.RestoreLatestSuccessfulRetained(ctx, admitted.Fence.Target(), asOf)
	if err != nil {
		t.Fatalf("restore latest retained backup into fresh SeaweedFS-backed target: %v", err)
	}
	admitted.Complete(t, serviceBackedRestore)
	var retainedCount int
	if err := targetPool.QueryRow(ctx, `SELECT count(*) FROM network_flow_graph_views
	 WHERE semantic_query_json->>'counter'='9007199254740993' AND declaration_state='retired' AND display_name='Retained Ω'`).Scan(&retainedCount); err != nil || retainedCount != 2 {
		t.Fatalf("logical restored bytes: count=%d error=%v", retainedCount, err)
	}
	assertRestoredInactiveStateRejectedOnlyByClaim(t, targetPool)
	if serviceBackedRestore.BackupSet.BackupSetID != latestID || serviceBackedRestore.ConsistencyReport.BlobCount != 1 {
		t.Fatalf("service-backed restore selected wrong backup or missed blob lifecycle: %#v", serviceBackedRestore)
	}
	var restoredVersion int64
	if err := targetPool.QueryRow(ctx, `SELECT last_version FROM incident_membership_versions WHERE incident_id=$1 AND user_id=$2`, incidentID, historyUser.ID).Scan(&restoredVersion); err != nil || restoredVersion != 1 {
		t.Fatalf("restored deleted-membership history=%d err=%v", restoredVersion, err)
	}
	var restoredMemberships int
	if err := targetPool.QueryRow(ctx, `SELECT count(*) FROM incident_memberships WHERE incident_id=$1 AND user_id=$2`, incidentID, historyUser.ID).Scan(&restoredMemberships); err != nil || restoredMemberships != 0 {
		t.Fatalf("restore recreated deleted membership: %d err=%v", restoredMemberships, err)
	}
	restoredObject, _, err := targetObjectStore.ReadObject(ctx, objectKey, objectstore.ReadOptions{})
	if err != nil {
		t.Fatalf("read restored SeaweedFS object: %v", err)
	}
	restoredBytes, readErr := io.ReadAll(restoredObject)
	closeErr := restoredObject.Close()
	if readErr != nil {
		t.Fatalf("read restored SeaweedFS object body: %v", readErr)
	}
	if closeErr != nil {
		t.Fatalf("close restored SeaweedFS object body: %v", closeErr)
	}
	if !bytes.Equal(restoredBytes, objectPayload) {
		t.Fatalf("restored SeaweedFS object bytes changed: got %q want %q", restoredBytes, objectPayload)
	}
	var restoredBlobCount int
	if err := targetPool.QueryRow(ctx, `
SELECT count(*)
  FROM object_blobs
 WHERE object_blob_id = $1
   AND storage_key = $2
   AND upload_state = 'available'
   AND observed_sha256_hex = $3
`, objectBlobID, objectKey, objectSHA).Scan(&restoredBlobCount); err != nil {
		t.Fatalf("query restored object blob row: %v", err)
	}
	if restoredBlobCount != 1 {
		t.Fatalf("restored object blob lifecycle row count got %d want 1", restoredBlobCount)
	}

	reloaded, err := reopenedStore.GetBackupSet(ctx, latestID)
	if err != nil {
		t.Fatalf("reload committed latest backup metadata: %v", err)
	}
	if reloaded.BackupSetID != latestCreated.BackupSetID ||
		!reloaded.ConsistencyPointAt.Equal(latestCreated.ConsistencyPointAt) ||
		reloaded.PostgresRestoreAnchor != latestCreated.PostgresRestoreAnchor ||
		reloaded.ObjectStoreRestoreAnchor != latestCreated.ObjectStoreRestoreAnchor ||
		reloaded.IntegrityManifestSHA256 != latestCreated.IntegrityManifestSHA256 {
		t.Fatalf("committed metadata did not persist stable identity, point, and anchors:\ncreated=%#v\nreloaded=%#v", latestCreated, reloaded)
	}
	var incidentProof recovery.BackupArtifactStreamProof
	for _, artifact := range latestCaptured.IntegrityManifest.Artifacts {
		if strings.HasSuffix(artifact.LogicalRef, "/postgres/incidents.ndjson") {
			incidentProof = recovery.BackupArtifactStreamProof{LogicalRef: artifact.LogicalRef, ContentType: artifact.ContentType, PlaintextBytes: artifact.PlaintextBytes, PlaintextSHA256: artifact.PlaintextSHA256, EnvelopeRef: artifact.EnvelopeRef, EnvelopeSHA256: artifact.EnvelopeSHA256}
		}
	}
	if incidentProof.LogicalRef == "" {
		t.Fatal("backup omitted incident table unit")
	}
	rawBody, err := os.ReadFile(filepath.Join(harness.Server.Config.Roots.BackupStorage.Path, filepath.FromSlash(incidentProof.EnvelopeRef)))
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(rawBody, []byte("backup_restore-i-10-01")) {
		t.Fatal("encrypted backup envelope contains incident plaintext")
	}
	var incidentRows bytes.Buffer
	if err := streaming.ReadArtifactStream(ctx, incidentProof, &incidentRows); err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(incidentRows.Bytes(), []byte("backup_restore-i-10-01")) {
		t.Fatal("authenticated backup omitted incident content")
	}
	var manifestBody bytes.Buffer
	if err := streaming.ReadArtifactStream(ctx, latestCaptured.IntegrityProof, &manifestBody); err != nil {
		t.Fatal(err)
	}
	var manifest recovery.VNextBackupIntegrityManifest
	if err := json.Unmarshal(manifestBody.Bytes(), &manifest); err != nil {
		t.Fatal(err)
	}
	if manifest.SchemaID != recovery.BackupIntegrityManifestV4SchemaID || manifest.BackupSetID != latestID.String() || reloaded.IntegrityManifestSHA256 != latestCaptured.IntegrityProof.PlaintextSHA256 {
		t.Fatalf("persisted integrity manifest does not match reloaded metadata: %#v", manifest)
	}
	if err := recovery.NewBackupCatalog(reopenedStore, backupStorage, testExtensionBackupCatalog(t), stateCatalog).VerifyBackupSetDurability(ctx, reloaded); err != nil {
		t.Fatal(err)
	}

	if reloaded.VerificationState != recovery.VerificationUnverified || reloaded.LastVerifiedRestoreAt != nil {
		t.Fatalf("committed backup metadata must remain unverified with null restore timestamp until verification: %#v", reloaded)
	}
	requireRetentionFloor(t, reloaded.CreatedAt, reloaded.RetainedUntil, "retained_until")
	requireRetentionFloor(t, reloaded.CreatedAt, reloaded.PostgresRestoreAnchorRetainedUntil, "postgres_restore_anchor_retained_until")
	requireRetentionFloor(t, reloaded.CreatedAt, reloaded.ObjectStoreRestoreAnchorRetainedUntil, "object_store_restore_anchor_retained_until")

	latest, err := reopenedStore.LatestSuccessfulRetainedBackup(ctx, asOf)
	if err != nil {
		t.Fatalf("latest successful retained lookup against real backing Postgres: %v", err)
	}
	if latest.BackupSetID != latestID {
		t.Fatalf("latest lookup got %s want %s", latest.BackupSetID, latestID)
	}
}

func S3StoreForBucket(t testing.TB, harness *appsupport.ServerHarness, objects *s3test.Harness, bucket string) objectstore.Store {
	t.Helper()
	const serviceRef = "object_primary"
	cfg := harness.Server.Config
	cfg.Roots.ObjectStorage.BindingKind = "managed_service"
	cfg.Roots.ObjectStorage.Path = ""
	cfg.Roots.ObjectStorage.ServiceRef = serviceRef
	env := objects.EnvForServiceRef(serviceRef, bucket)
	store, err := appsupport.OpenObjectStore(context.Background(), cfg, env)
	if err != nil {
		t.Fatalf("open SeaweedFS object-store adapter for bucket %s: %v", bucket, err)
	}
	t.Cleanup(func() {
		_ = store.Close()
	})
	return store
}

func SHA256Hex(body []byte) string {
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:])
}

func asTime(t testing.TB, value string) time.Time {
	t.Helper()
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		t.Fatalf("parse time fixture %q: %v", value, err)
	}
	return parsed
}

func assertRestoredInactiveStateRejectedOnlyByClaim(t testing.TB, pool *pgxpool.Pool) {
	t.Helper()
	calls := 0
	claim := extensionstate.NetworkFlowClaim(t, pool, networkflow.ExtensionStateFamilyCounters(), func(ctx context.Context, reader extensionstate.Reader) error {
		calls++
		return networkflow.ValidateExtensionState(ctx, reader)
	})
	if err := claim(); err == nil || calls != 1 {
		t.Fatalf("later owner claim: calls=%d error=%v", calls, err)
	}
	var count int
	if err := pool.QueryRow(t.Context(), `SELECT count(*) FROM network_flow_graph_views`).Scan(&count); err != nil || count != 2 {
		t.Fatalf("failed claim changed restored entries: %d %v", count, err)
	}
}

// Faults are introduced at the target storage boundary. Capture, authenticated
// input admission, transaction handling and logical readback are production code.
type observedBindingRestore struct {
	recovery.VNextRestoreTarget
	defect  string
	steps   []string
	objects int
}

func (target *observedBindingRestore) WithAtomicRestore(ctx context.Context, catalog *recoverystate.Catalog, run func(recovery.VNextRestoreMutation) error) error {
	return target.VNextRestoreTarget.WithAtomicRestore(ctx, catalog, func(mutation recovery.VNextRestoreMutation) error {
		return run(&observedBindingMutation{VNextRestoreMutation: mutation, target: target})
	})
}

type observedBindingMutation struct {
	recovery.VNextRestoreMutation
	target  *observedBindingRestore
	changed bool
}

func (m *observedBindingMutation) InsertPostgresRow(ctx context.Context, table string, row json.RawMessage) error {
	if table != "network_flow_graph_views" || m.changed {
		return m.VNextRestoreMutation.InsertPostgresRow(ctx, table, row)
	}
	m.changed = true
	if m.target.defect == "missing" {
		return nil
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(row, &fields); err != nil {
		return err
	}
	if m.target.defect == "extra" {
		if err := m.VNextRestoreMutation.InsertPostgresRow(ctx, table, row); err != nil {
			return err
		}
	}
	switch m.target.defect {
	case "extra", "identity":
		fields["graph_view_id"] = json.RawMessage(`"nfgv_11111111111111111111111111111111"`)
	case "content":
		fields["display_name"] = json.RawMessage(`"Changed Ω"`)
	default:
		return fmt.Errorf("unknown target defect")
	}
	changed, err := json.Marshal(fields)
	if err != nil {
		return err
	}
	return m.VNextRestoreMutation.InsertPostgresRow(ctx, table, changed)
}
func (m *observedBindingMutation) FinishPostgresTable(ctx context.Context, table string) error {
	m.target.steps = append(m.target.steps, "table:"+table)
	return m.VNextRestoreMutation.FinishPostgresTable(ctx, table)
}
func (m *observedBindingMutation) ValidatePostgresBinding(ctx context.Context, binding recovery.ExtensionBackupBinding, proof recovery.PostgresBindingProof) error {
	m.target.steps = append(m.target.steps, "validate:"+binding.BindingID)
	return m.VNextRestoreMutation.ValidatePostgresBinding(ctx, binding, proof)
}
func (m *observedBindingMutation) RestoreObject(ctx context.Context, object recovery.VNextObjectManifestEntry, reader io.Reader) error {
	m.target.objects++
	return m.VNextRestoreMutation.RestoreObject(ctx, object, reader)
}

// The writer commits after the table unit is streamed but before binding-proof
// queries. Proofs and table bytes must continue to observe the same real snapshot.
type bindingSnapshotCut struct {
	recovery.VNextSnapshotRepository
	snapshots, mutations int
	mutate               func(context.Context) error
}

func (r *bindingSnapshotCut) WithinRepeatableReadReadOnly(ctx context.Context, run func(recovery.VNextSnapshot, recovery.TargetControl) error) error {
	r.snapshots++
	return r.VNextSnapshotRepository.WithinRepeatableReadReadOnly(ctx, func(snapshot recovery.VNextSnapshot, control recovery.TargetControl) error {
		return run(bindingSnapshotView{VNextSnapshot: snapshot, repository: r}, control)
	})
}

type bindingSnapshotView struct {
	recovery.VNextSnapshot
	repository *bindingSnapshotCut
}

func (s bindingSnapshotView) StreamCanonicalTableRows(ctx context.Context, table string, visit func(json.RawMessage) error) error {
	if err := s.VNextSnapshot.StreamCanonicalTableRows(ctx, table, visit); err != nil {
		return err
	}
	if table == "network_flow_graph_views" && s.repository.snapshots == 2 {
		if err := s.repository.mutate(ctx); err != nil {
			return err
		}
		s.repository.mutations++
	}
	return nil
}
