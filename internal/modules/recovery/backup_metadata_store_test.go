package recovery_test

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestBackupMetadataShapeAndRetentionFloors_Unit(t *testing.T) {
	db := pgtest.Start(t).BeginRollbackDBT(t, "backup_restore-u-10-01-shape")
	store := recovery.NewStore(db)
	capture := newCaptureService(t, store)
	ctx := context.Background()

	asOf := time.Date(2026, 5, 22, 12, 0, 0, 0, time.UTC)
	createdAt := asOf.Add(-2 * time.Hour)
	consistencyPointAt := asOf.Add(-90 * time.Minute)
	backupSetID := uuid.MustParse("00000000-0000-0000-0000-000000100101")
	retainedUntil := createdAt.Add(31 * 24 * time.Hour)

	created, err := capture.CaptureBackupSet(ctx, captureParams(recovery.CaptureBackupSetParams{
		BackupSetID:                           backupSetID,
		ConsistencyPointAt:                    consistencyPointAt,
		CreatedAt:                             createdAt,
		RetainedUntil:                         retainedUntil,
		PostgresRestoreAnchorRetainedUntil:    createdAt.Add(30 * 24 * time.Hour),
		ObjectStoreRestoreAnchorRetainedUntil: retainedUntil,
	}))
	if err != nil {
		t.Fatalf("capture successful backup metadata: %v", err)
	}
	if created.BackupSetID != backupSetID {
		t.Fatalf("backup_set_id not stable: got %s want %s", created.BackupSetID, backupSetID)
	}
	if !created.ConsistencyPointAt.Equal(consistencyPointAt) {
		t.Fatalf("unexpected consistency_point_at: got %s want %s", created.ConsistencyPointAt, consistencyPointAt)
	}
	if created.PostgresRestoreAnchor == "" || created.ObjectStoreRestoreAnchor == "" {
		t.Fatalf("successful backup metadata must include both restore anchors: %#v", created)
	}
	requireArtifactProof(t, created)
	if created.VerificationState != recovery.VerificationUnverified {
		t.Fatalf("new backup metadata must default to unverified, got %q", created.VerificationState)
	}
	if created.LastVerifiedRestoreAt != nil {
		t.Fatalf("unverified backup metadata must have null last_verified_restore_at, got %s", created.LastVerifiedRestoreAt)
	}
	requireRetentionFloor(t, created.CreatedAt, created.RetainedUntil, "retained_until")
	requireRetentionFloor(t, created.CreatedAt, created.PostgresRestoreAnchorRetainedUntil, "postgres_restore_anchor_retained_until")
	requireRetentionFloor(t, created.CreatedAt, created.ObjectStoreRestoreAnchorRetainedUntil, "object_store_restore_anchor_retained_until")

	reloaded, err := store.GetBackupSet(ctx, backupSetID)
	if err != nil {
		t.Fatalf("reload backup metadata: %v", err)
	}
	if reloaded != created {
		t.Fatalf("reloaded metadata changed stable identity or anchors:\ncreated=%#v\nreloaded=%#v", created, reloaded)
	}

	latest, err := store.LatestSuccessfulRetainedBackup(ctx, asOf)
	if err != nil {
		t.Fatalf("latest successful retained backup: %v", err)
	}
	if latest.BackupSetID != backupSetID || !latest.ConsistencyPointAt.Equal(consistencyPointAt) {
		t.Fatalf("latest lookup returned wrong backup metadata: %#v", latest)
	}

	retained, err := store.ListRetainedBackupSetMetadata(ctx, asOf)
	if err != nil {
		t.Fatalf("list retained backup metadata: %v", err)
	}
	if len(retained) != 1 || retained[0].BackupSetID != backupSetID {
		t.Fatalf("expected exactly one retained successful backup, got %#v", retained)
	}
}

func TestVerificationVocabularyAndTimestampRules_Unit(t *testing.T) {
	db := pgtest.Start(t).BeginRollbackDBT(t, "backup_restore-u-10-01-state")
	store := recovery.NewStore(db)
	ctx := context.Background()

	createdAt := time.Date(2026, 5, 22, 10, 0, 0, 0, time.UTC)
	backupSetID := uuid.MustParse("00000000-0000-0000-0000-000000100102")
	var (
		state        string
		verifiedTime pgtype.Timestamptz
	)
	if err := db.QueryRow(ctx, `
INSERT INTO backup_sets (
    backup_set_id,
    consistency_point_at,
    postgres_restore_anchor,
    object_store_restore_anchor,
    postgres_artifact_key,
    postgres_artifact_sha256,
    postgres_artifact_size_bytes,
    object_store_artifact_key,
    object_store_artifact_sha256,
    object_store_artifact_size_bytes,
    integrity_manifest_key,
    integrity_manifest_sha256,
    integrity_manifest_size_bytes,
    created_at,
    retained_until,
    postgres_restore_anchor_retained_until,
    object_store_restore_anchor_retained_until
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
RETURNING verification_state, last_verified_restore_at
`, validBackupSetInsertArgs(backupSetID, createdAt, createdAt.Add(30*24*time.Hour), createdAt.Add(30*24*time.Hour), createdAt.Add(30*24*time.Hour))...).Scan(&state, &verifiedTime); err != nil {
		t.Fatalf("insert backup metadata using DB defaults: %v", err)
	}
	if state != string(recovery.VerificationUnverified) {
		t.Fatalf("verification_state default got %q want %q", state, recovery.VerificationUnverified)
	}
	if verifiedTime.Valid {
		t.Fatalf("unverified default must leave last_verified_restore_at null, got %s", verifiedTime.Time)
	}

	requireCheckViolationExec(t, db, "backup_sets_verification_state_check", `
UPDATE backup_sets
   SET verification_state = 'pending',
       last_verified_restore_at = NULL
 WHERE backup_set_id = $1
`, backupSetID)

	requireCheckViolationExec(t, db, "backup_sets_verification_timestamp_check", `
UPDATE backup_sets
   SET verification_state = 'verified',
       last_verified_restore_at = NULL
 WHERE backup_set_id = $1
`, backupSetID)

	if _, err := store.UpdateVerificationState(ctx, backupSetID, recovery.VerificationState("pending"), nil); !errors.Is(err, recovery.ErrInvalidVerificationState) {
		t.Fatalf("invalid store vocabulary error got %v want %v", err, recovery.ErrInvalidVerificationState)
	}
	nonNullWhileUnverified := createdAt.Add(2 * time.Hour)
	if _, err := store.UpdateVerificationState(ctx, backupSetID, recovery.VerificationUnverified, &nonNullWhileUnverified); !errors.Is(err, recovery.ErrVerificationTimestampForbidden) {
		t.Fatalf("unverified non-null timestamp error got %v want %v", err, recovery.ErrVerificationTimestampForbidden)
	}
	if _, err := store.UpdateVerificationState(ctx, backupSetID, recovery.VerificationVerified, nil); !errors.Is(err, recovery.ErrVerificationTimestampRequired) {
		t.Fatalf("verified null timestamp error got %v want %v", err, recovery.ErrVerificationTimestampRequired)
	}

	verifiedAt := createdAt.Add(3 * time.Hour)
	verified, err := store.UpdateVerificationState(ctx, backupSetID, recovery.VerificationVerified, &verifiedAt)
	if err != nil {
		t.Fatalf("set verified state: %v", err)
	}
	if verified.VerificationState != recovery.VerificationVerified || verified.LastVerifiedRestoreAt == nil || !verified.LastVerifiedRestoreAt.Equal(verifiedAt) {
		t.Fatalf("verified transition did not persist timestamp: %#v", verified)
	}

	failedAt := createdAt.Add(4 * time.Hour)
	failed, err := store.UpdateVerificationState(ctx, backupSetID, recovery.VerificationFailed, &failedAt)
	if err != nil {
		t.Fatalf("set failed state: %v", err)
	}
	if failed.VerificationState != recovery.VerificationFailed || failed.LastVerifiedRestoreAt == nil || !failed.LastVerifiedRestoreAt.Equal(failedAt) {
		t.Fatalf("failed transition did not persist timestamp: %#v", failed)
	}
}

func TestRestoreVerificationDueSelectionAndAtomicCompletion(t *testing.T) {
	db := pgtest.Start(t).BeginRollbackDBT(t, "backup_restore-u-10-04-verification")
	store := recovery.NewStore(db)
	capture := newCaptureService(t, store)
	ctx := context.Background()

	asOf := time.Date(2026, 5, 22, 12, 0, 0, 0, time.UTC)
	createdAt := asOf.Add(-2 * time.Hour)
	backupSetID := uuid.MustParse("00000000-0000-0000-0000-000000100401")
	backupSet, err := capture.CaptureBackupSet(ctx, captureParams(recovery.CaptureBackupSetParams{
		BackupSetID:        backupSetID,
		ConsistencyPointAt: asOf.Add(-time.Hour),
		CreatedAt:          createdAt,
		RetainedUntil:      createdAt.Add(31 * 24 * time.Hour),
	}))
	if err != nil {
		t.Fatalf("capture backup set for restore verification due selection: %v", err)
	}
	basisA, err := recovery.RestoreVerificationBasisSHA256(map[string]string{
		"backup_mechanism": "test.v1",
		"source_root":      "root-a",
	})
	if err != nil {
		t.Fatalf("basis A: %v", err)
	}
	basisB, err := recovery.RestoreVerificationBasisSHA256(map[string]string{
		"backup_mechanism": "test.v1",
		"source_root":      "root-b",
	})
	if err != nil {
		t.Fatalf("basis B: %v", err)
	}

	due, err := store.ListBackupsDueForRestoreVerification(ctx, asOf, basisA)
	if err != nil {
		t.Fatalf("list initially due backups: %v", err)
	}
	if len(due) != 1 || due[0].BackupSetID != backupSetID {
		t.Fatalf("never-verified backup must be due, got %#v", due)
	}

	completedAt := asOf.Add(-6 * 24 * time.Hour)
	completed, run, err := store.RecordRestoreVerificationCompletion(ctx, recovery.CreateRestoreVerificationRunParams{
		BackupSetID:             backupSet.BackupSetID,
		StartedAt:               completedAt.Add(-2 * time.Minute),
		CompletedAt:             completedAt,
		VerificationState:       recovery.VerificationVerified,
		VerificationBasisSHA256: basisA,
		ConsistencyReport: recovery.RestoreConsistencyReport{
			AuthoritativeRowsSHA256: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			AuthoritativeRowCount:   5,
			ChangeSetsSHA256:        "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
			ChangeSetRowCount:       3,
			BlobHashesSHA256:        "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc",
			BlobCount:               1,
		},
	})
	if err != nil {
		t.Fatalf("record restore verification completion: %v", err)
	}
	if completed.VerificationState != recovery.VerificationVerified ||
		completed.LastVerifiedRestoreAt == nil ||
		!completed.LastVerifiedRestoreAt.Equal(completedAt) ||
		completed.LastVerificationBasisSHA256 != basisA {
		t.Fatalf("restore verification completion did not atomically update backup set: %#v", completed)
	}
	if run.BackupSetID != backupSetID ||
		run.VerificationState != recovery.VerificationVerified ||
		run.VerificationBasisSHA256 != basisA ||
		run.ConsistencyReport.BlobCount != 1 {
		t.Fatalf("restore verification run history not persisted: %#v", run)
	}

	due, err = store.ListBackupsDueForRestoreVerification(ctx, asOf, basisA)
	if err != nil {
		t.Fatalf("list backups due after fresh verification: %v", err)
	}
	if len(due) != 0 {
		t.Fatalf("fresh verification with same basis must not be due, got %#v", due)
	}
	due, err = store.ListBackupsDueForRestoreVerification(ctx, asOf, basisB)
	if err != nil {
		t.Fatalf("list backups due after basis change: %v", err)
	}
	if len(due) != 1 || due[0].BackupSetID != backupSetID {
		t.Fatalf("basis change must make backup due, got %#v", due)
	}

	oldVerifiedAt := asOf.Add(-8 * 24 * time.Hour)
	if _, err := store.UpdateVerificationState(ctx, backupSetID, recovery.VerificationVerified, &oldVerifiedAt, basisA); err != nil {
		t.Fatalf("age verified state: %v", err)
	}
	due, err = store.ListBackupsDueForRestoreVerification(ctx, asOf, basisA)
	if err != nil {
		t.Fatalf("list backups due after stale verification: %v", err)
	}
	if len(due) != 1 || due[0].BackupSetID != backupSetID {
		t.Fatalf("verification older than seven days must be due, got %#v", due)
	}
}

func TestLatestSuccessfulRetainedBackupRequiresTwentyFourHourFloor_Unit(t *testing.T) {
	db := pgtest.Start(t).BeginRollbackDBT(t, "backup_restore-u-10-01-latest")
	store := recovery.NewStore(db)
	capture := newCaptureService(t, store)
	ctx := context.Background()

	asOf := time.Date(2026, 5, 22, 12, 0, 0, 0, time.UTC)
	staleCreatedAt := asOf.Add(-26 * time.Hour)
	staleBackupSetID := uuid.MustParse("00000000-0000-0000-0000-000000100103")
	if _, err := capture.CaptureBackupSet(ctx, captureParams(recovery.CaptureBackupSetParams{
		BackupSetID:        staleBackupSetID,
		ConsistencyPointAt: asOf.Add(-25 * time.Hour),
		CreatedAt:          staleCreatedAt,
		RetainedUntil:      staleCreatedAt.Add(31 * 24 * time.Hour),
	})); err != nil {
		t.Fatalf("create stale retained backup metadata: %v", err)
	}

	if _, err := store.LatestSuccessfulRetainedBackup(ctx, asOf); !errors.Is(err, recovery.ErrLatestSuccessfulBackupStale) {
		t.Fatalf("stale latest backup error got %v want %v", err, recovery.ErrLatestSuccessfulBackupStale)
	} else {
		var staleErr *recovery.LatestSuccessfulBackupStaleError
		if !errors.As(err, &staleErr) || staleErr.BackupSet.BackupSetID != staleBackupSetID {
			t.Fatalf("stale error must identify the latest retained backup, got %#v", err)
		}
	}

	expiredCreatedAt := asOf.Add(-40 * 24 * time.Hour)
	if _, err := capture.CaptureBackupSet(ctx, captureParams(recovery.CaptureBackupSetParams{
		BackupSetID:        uuid.MustParse("00000000-0000-0000-0000-000000100104"),
		ConsistencyPointAt: asOf.Add(-time.Hour),
		CreatedAt:          expiredCreatedAt,
		RetainedUntil:      expiredCreatedAt.Add(30 * 24 * time.Hour),
	})); err != nil {
		t.Fatalf("create expired but otherwise fresh-looking backup metadata: %v", err)
	}
	if _, err := store.LatestSuccessfulRetainedBackup(ctx, asOf); !errors.Is(err, recovery.ErrLatestSuccessfulBackupStale) {
		t.Fatalf("expired fresh-looking backup must not satisfy latest retained lookup: %v", err)
	}

	freshCreatedAt := asOf.Add(-3 * time.Hour)
	freshBackupSetID := uuid.MustParse("00000000-0000-0000-0000-000000100105")
	fresh, err := capture.CaptureBackupSet(ctx, captureParams(recovery.CaptureBackupSetParams{
		BackupSetID:        freshBackupSetID,
		ConsistencyPointAt: asOf.Add(-2 * time.Hour),
		CreatedAt:          freshCreatedAt,
		RetainedUntil:      freshCreatedAt.Add(31 * 24 * time.Hour),
	}))
	if err != nil {
		t.Fatalf("create fresh retained backup metadata: %v", err)
	}
	latest, err := store.LatestSuccessfulRetainedBackup(ctx, asOf)
	if err != nil {
		t.Fatalf("fresh latest backup should satisfy 24-hour floor: %v", err)
	}
	if latest.BackupSetID != fresh.BackupSetID {
		t.Fatalf("latest lookup got %s want fresh backup %s", latest.BackupSetID, fresh.BackupSetID)
	}
}

func TestDurableCatalogSkipsMetadataWithMissingArtifacts_Unit(t *testing.T) {
	db := pgtest.Start(t).BeginRollbackDBT(t, "backup_restore-u-10-01-durable-catalog")
	store := recovery.NewStore(db)
	backupStorage := newEncryptedBackupStorage(t, t.TempDir())
	capture := recovery.NewCaptureService(store, backupStorage, testExtensionBackupCatalog(t))
	ctx := context.Background()

	asOf := time.Date(2026, 5, 22, 12, 0, 0, 0, time.UTC)
	older, err := capture.CaptureBackupSet(ctx, captureParams(recovery.CaptureBackupSetParams{
		BackupSetID:        uuid.MustParse("00000000-0000-0000-0000-000000100111"),
		ConsistencyPointAt: asOf.Add(-2 * time.Hour),
		CreatedAt:          asOf.Add(-3 * time.Hour),
		RetainedUntil:      asOf.Add(31 * 24 * time.Hour),
	}))
	if err != nil {
		t.Fatalf("capture older durable backup: %v", err)
	}
	newer, err := capture.CaptureBackupSet(ctx, captureParams(recovery.CaptureBackupSetParams{
		BackupSetID:        uuid.MustParse("00000000-0000-0000-0000-000000100112"),
		ConsistencyPointAt: asOf.Add(-time.Hour),
		CreatedAt:          asOf.Add(-90 * time.Minute),
		RetainedUntil:      asOf.Add(31 * 24 * time.Hour),
	}))
	if err != nil {
		t.Fatalf("capture newer durable backup: %v", err)
	}

	catalog := recovery.NewBackupCatalog(store, tamperedBackupStorage{
		Inner:   backupStorage,
		Missing: map[string]bool{newer.IntegrityManifestKey: true},
	}, testExtensionBackupCatalog(t))
	selection, err := catalog.RestoreCandidateBackupSelection(ctx, asOf)
	if err != nil {
		t.Fatalf("select latest durable backup: %v", err)
	}
	selected := selection.BackupSet
	if selected.BackupSetID != older.BackupSetID {
		t.Fatalf("catalog selected %s want older durable %s", selected.BackupSetID, older.BackupSetID)
	}
	if len(selection.DurabilityDiagnostics) != 1 {
		t.Fatalf("durability diagnostics count got %d want 1: %#v", len(selection.DurabilityDiagnostics), selection.DurabilityDiagnostics)
	}
	diagnostic := selection.DurabilityDiagnostics[0]
	if diagnostic.BackupSetID != newer.BackupSetID || diagnostic.Code != "artifact_missing" {
		t.Fatalf("unexpected redacted durability diagnostic: %#v", diagnostic)
	}
}

func TestRetentionFloorRejectsShortMetadataAndArtifacts_Unit(t *testing.T) {
	db := pgtest.Start(t).BeginRollbackDBT(t, "backup_restore-u-10-01-retention")
	store := recovery.NewStore(db)
	capture := newCaptureService(t, store)
	ctx := context.Background()

	createdAt := time.Date(2026, 5, 22, 10, 0, 0, 0, time.UTC)
	if _, err := capture.CaptureBackupSet(ctx, captureParams(recovery.CaptureBackupSetParams{
		BackupSetID:        uuid.MustParse("00000000-0000-0000-0000-000000100106"),
		ConsistencyPointAt: createdAt.Add(-time.Hour),
		CreatedAt:          createdAt,
		RetainedUntil:      createdAt.Add(29 * 24 * time.Hour),
	})); !errors.Is(err, recovery.ErrRetentionFloor) {
		t.Fatalf("short metadata retention error got %v want %v", err, recovery.ErrRetentionFloor)
	}

	requireCheckViolationExec(t, db, "backup_sets_retained_until_floor_check", `
INSERT INTO backup_sets (
    backup_set_id,
    consistency_point_at,
    postgres_restore_anchor,
    object_store_restore_anchor,
    postgres_artifact_key,
    postgres_artifact_sha256,
    postgres_artifact_size_bytes,
    object_store_artifact_key,
    object_store_artifact_sha256,
    object_store_artifact_size_bytes,
    integrity_manifest_key,
    integrity_manifest_sha256,
    integrity_manifest_size_bytes,
    created_at,
    retained_until,
    postgres_restore_anchor_retained_until,
    object_store_restore_anchor_retained_until
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
`, validBackupSetInsertArgs(uuid.MustParse("00000000-0000-0000-0000-000000100107"), createdAt, createdAt.Add(29*24*time.Hour), createdAt.Add(30*24*time.Hour), createdAt.Add(30*24*time.Hour))...)

	requireCheckViolationExec(t, db, "backup_sets_postgres_anchor_retained_until_floor_check", `
INSERT INTO backup_sets (
    backup_set_id,
    consistency_point_at,
    postgres_restore_anchor,
    object_store_restore_anchor,
    postgres_artifact_key,
    postgres_artifact_sha256,
    postgres_artifact_size_bytes,
    object_store_artifact_key,
    object_store_artifact_sha256,
    object_store_artifact_size_bytes,
    integrity_manifest_key,
    integrity_manifest_sha256,
    integrity_manifest_size_bytes,
    created_at,
    retained_until,
    postgres_restore_anchor_retained_until,
    object_store_restore_anchor_retained_until
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
`, validBackupSetInsertArgs(uuid.MustParse("00000000-0000-0000-0000-000000100108"), createdAt, createdAt.Add(30*24*time.Hour), createdAt.Add(29*24*time.Hour), createdAt.Add(30*24*time.Hour))...)
}

type tamperedBackupStorage struct {
	Inner        recovery.BackupStorage
	Missing      map[string]bool
	Replacements map[string][]byte
}

func (storage tamperedBackupStorage) WriteArtifact(ctx context.Context, key string, body []byte, contentType string) (recovery.BackupArtifactProof, error) {
	return storage.Inner.WriteArtifact(ctx, key, body, contentType)
}

func (storage tamperedBackupStorage) ReadArtifact(ctx context.Context, key string) ([]byte, error) {
	if storage.Missing[key] {
		return nil, os.ErrNotExist
	}
	if replacement, ok := storage.Replacements[key]; ok {
		return replacement, nil
	}
	return storage.Inner.ReadArtifact(ctx, key)
}

func TestCaptureRequiresArtifactProofs_Unit(t *testing.T) {
	db := pgtest.Start(t).BeginRollbackDBT(t, "recovery-metadata-artifacts")
	store := recovery.NewStore(db)
	capture := newCaptureService(t, store)

	createdAt := time.Date(2026, 5, 22, 10, 0, 0, 0, time.UTC)
	params := captureParams(recovery.CaptureBackupSetParams{
		BackupSetID:        uuid.MustParse("00000000-0000-0000-0000-000000100109"),
		ConsistencyPointAt: createdAt.Add(-time.Hour),
		CreatedAt:          createdAt,
		RetainedUntil:      createdAt.Add(30 * 24 * time.Hour),
	})
	params.PostgresArtifact.Body = nil
	if _, err := capture.CaptureBackupSet(context.Background(), params); !errors.Is(err, recovery.ErrInvalidBackupArtifact) {
		t.Fatalf("empty postgres artifact error got %v want %v", err, recovery.ErrInvalidBackupArtifact)
	}
}

func TestCaptureRequiresEncryptedBackupStorage_Unit(t *testing.T) {
	db := pgtest.Start(t).BeginRollbackDBT(t, "backup_restore-u-10-05-encrypted-storage")
	store := recovery.NewStore(db)
	rawStorage, err := recovery.NewFilesystemBackupStorage(t.TempDir())
	if err != nil {
		t.Fatalf("create raw backup storage: %v", err)
	}
	capture := recovery.NewCaptureService(store, rawStorage, testExtensionBackupCatalog(t))
	createdAt := time.Date(2026, 5, 22, 10, 0, 0, 0, time.UTC)
	if _, err := capture.CaptureBackupSet(context.Background(), captureParams(recovery.CaptureBackupSetParams{
		BackupSetID:        uuid.MustParse("00000000-0000-0000-0000-000000100110"),
		ConsistencyPointAt: createdAt.Add(-time.Hour),
		CreatedAt:          createdAt,
		RetainedUntil:      createdAt.Add(30 * 24 * time.Hour),
	})); !errors.Is(err, recovery.ErrEncryptedBackupStorage) {
		t.Fatalf("unencrypted backup storage error got %v want %v", err, recovery.ErrEncryptedBackupStorage)
	}
}

func TestEncryptedBackupStorageFailsClosedWithWrongKey_Unit(t *testing.T) {
	root := t.TempDir()
	storage := newEncryptedBackupStorage(t, root)
	ctx := context.Background()
	proof, err := storage.WriteArtifact(ctx, "backup_sets/wrong-key/proof.json", []byte(`{"marker":"secret incident data"}`), "application/json")
	if err != nil {
		t.Fatalf("write encrypted artifact: %v", err)
	}
	rawStorage, err := recovery.NewFilesystemBackupStorage(root)
	if err != nil {
		t.Fatalf("open raw storage: %v", err)
	}
	wrongKey, err := recovery.ParseRecoveryEncryptionKey("YWJjZGVmMDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODk=")
	if err != nil {
		t.Fatalf("parse wrong key: %v", err)
	}
	wrongStorage, err := recovery.NewEncryptedBackupStorage(rawStorage, wrongKey)
	if err != nil {
		t.Fatalf("create wrong-key storage: %v", err)
	}
	if _, err := recovery.VerifyArtifactProof(ctx, wrongStorage, proof); !errors.Is(err, recovery.ErrInvalidBackupArtifact) {
		t.Fatalf("wrong-key artifact proof error got %v want %v", err, recovery.ErrInvalidBackupArtifact)
	}
}

func newCaptureService(t *testing.T, store *recovery.Store) *recovery.CaptureService {
	t.Helper()
	storage := newEncryptedBackupStorage(t, t.TempDir())
	return recovery.NewCaptureService(store, storage, testExtensionBackupCatalog(t))
}

const RecoveryMasterKey = "MDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWY="

func newEncryptedBackupStorage(t testing.TB, root string) recovery.BackupStorage {
	t.Helper()
	rawStorage, err := recovery.NewFilesystemBackupStorage(root)
	if err != nil {
		t.Fatalf("create backup storage fixture: %v", err)
	}
	key, err := recovery.ParseRecoveryEncryptionKey(RecoveryMasterKey)
	if err != nil {
		t.Fatalf("parse recovery encryption key: %v", err)
	}
	storage, err := recovery.NewEncryptedBackupStorage(rawStorage, key)
	if err != nil {
		t.Fatalf("create encrypted backup storage fixture: %v", err)
	}
	return storage
}

func captureParams(params recovery.CaptureBackupSetParams) recovery.CaptureBackupSetParams {
	if params.PostgresArtifact.Body == nil {
		params.PostgresArtifact = recovery.BackupArtifact{
			Body:        emptyExtensionPostgresArtifact(),
			ContentType: "application/json",
		}
	}
	if params.ObjectStoreArtifact.Body == nil {
		params.ObjectStoreArtifact = recovery.BackupArtifact{
			Body:        []byte(`{"schema_id":"cartulary.object_store_snapshot_artifact.v2","objects":[]}`),
			ContentType: "application/json",
		}
	}
	if params.ObjectStoreBackupManifestArtifact.Body == nil && params.BackupSetID != uuid.Nil && !params.ConsistencyPointAt.IsZero() {
		snapshot, err := recovery.DecodeObjectStoreSnapshotArtifact(params.ObjectStoreArtifact.Body)
		if err != nil {
			panic(err)
		}
		manifest, manifestBody, err := recovery.BuildSeaweedFSS3ObjectStoreBackupManifest(snapshot, recovery.ObjectStoreBackupManifestParams{
			BackupSetID:        params.BackupSetID,
			ConsistencyPointAt: params.ConsistencyPointAt,
			Bucket:             "backup_restore-test-bucket",
		})
		if err != nil {
			panic(err)
		}
		_, summaryBody, err := recovery.BuildObjectStoreBackupSummary(manifest)
		if err != nil {
			panic(err)
		}
		params.ObjectStoreBackupManifestArtifact = recovery.BackupArtifact{Body: manifestBody, ContentType: "application/json"}
		params.ObjectStoreBackupSummaryArtifact = recovery.BackupArtifact{Body: summaryBody, ContentType: "application/json"}
	}
	return params
}

func requireArtifactProof(t *testing.T, backupSet recovery.BackupSet) {
	t.Helper()
	if backupSet.PostgresArtifactKey == "" || backupSet.ObjectStoreArtifactKey == "" || backupSet.IntegrityManifestKey == "" {
		t.Fatalf("backup metadata must include artifact keys: %#v", backupSet)
	}
	if len(backupSet.PostgresArtifactSHA256) != 64 || len(backupSet.ObjectStoreArtifactSHA256) != 64 || len(backupSet.IntegrityManifestSHA256) != 64 {
		t.Fatalf("backup metadata must include sha256 artifact proofs: %#v", backupSet)
	}
	if backupSet.PostgresArtifactSizeBytes <= 0 || backupSet.ObjectStoreArtifactSizeBytes <= 0 || backupSet.IntegrityManifestSizeBytes <= 0 {
		t.Fatalf("backup metadata must include positive artifact sizes: %#v", backupSet)
	}
}

func validBackupSetInsertArgs(backupSetID uuid.UUID, createdAt time.Time, retainedUntil time.Time, postgresRetainedUntil time.Time, objectRetainedUntil time.Time) []any {
	const (
		postgresSHA  = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
		objectSHA    = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
		manifestSHA  = "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"
		postgresSize = int64(17)
		objectSize   = int64(19)
		manifestSize = int64(23)
	)
	return []any{
		backupSetID,
		createdAt.Add(-time.Hour),
		"backup-storage://backup_restore/postgres/" + backupSetID.String(),
		"backup-storage://backup_restore/object-store/" + backupSetID.String(),
		"backup_restore/postgres/" + backupSetID.String() + ".json",
		postgresSHA,
		postgresSize,
		"backup_restore/object-store/" + backupSetID.String() + ".json",
		objectSHA,
		objectSize,
		"backup_restore/manifests/" + backupSetID.String() + ".json",
		manifestSHA,
		manifestSize,
		createdAt,
		retainedUntil,
		postgresRetainedUntil,
		objectRetainedUntil,
	}
}

func requireRetentionFloor(t *testing.T, createdAt time.Time, retainedUntil time.Time, label string) {
	t.Helper()
	floor := createdAt.Add(recovery.MinimumRetentionDuration)
	if retainedUntil.Before(floor) {
		t.Fatalf("%s got %s, before floor %s", label, retainedUntil, floor)
	}
}

func requireCheckViolation(t *testing.T, err error, constraintName string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected check violation %s, got nil", constraintName)
	}
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		t.Fatalf("expected postgres check violation %s, got %T: %v", constraintName, err, err)
	}
	if pgErr.Code != "23514" || pgErr.ConstraintName != constraintName {
		t.Fatalf("unexpected postgres error: code=%s constraint=%s message=%s; want check violation %s", pgErr.Code, pgErr.ConstraintName, pgErr.Message, constraintName)
	}
}

type rollbackExecer interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
}

func requireCheckViolationExec(t *testing.T, db rollbackExecer, constraintName string, sql string, args ...any) {
	t.Helper()
	ctx := context.Background()
	if _, err := db.Exec(ctx, "SAVEPOINT backup_restore_expected_check_violation"); err != nil {
		t.Fatalf("create check-violation savepoint: %v", err)
	}
	_, err := db.Exec(ctx, sql, args...)
	requireCheckViolation(t, err, constraintName)
	if _, rollbackErr := db.Exec(ctx, "ROLLBACK TO SAVEPOINT backup_restore_expected_check_violation"); rollbackErr != nil {
		t.Fatalf("rollback check-violation savepoint: %v", rollbackErr)
	}
	if _, releaseErr := db.Exec(ctx, "RELEASE SAVEPOINT backup_restore_expected_check_violation"); releaseErr != nil {
		t.Fatalf("release check-violation savepoint: %v", releaseErr)
	}
}
