package reference_data_test

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/app/extensionassembly"
	"github.com/JochiRaider/cartulary/internal/app/projectionassembly"
	"github.com/JochiRaider/cartulary/internal/app/recoveryassembly"
	"github.com/JochiRaider/cartulary/internal/app/referenceassembly"
	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	referencetest "github.com/JochiRaider/cartulary/internal/modules/reference_data/testsupport"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/platform/pagination"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestCanonicalBasePersistencePinsAndImmutableProvenance_Integration(t *testing.T) {
	ctx := context.Background()
	db := pgtest.Start(t).PrepareIsolatedDatabaseT(t, "reference-pack-immutable-base")
	pool, err := pgxpool.New(ctx, db.DSN)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	for _, table := range []string{"reference_pack_portable_preparations", "reference_pack_portable_selections"} {
		for _, role := range []string{"cartulary_runtime", "cartulary_recovery"} {
			for _, privilege := range []string{"SELECT", "INSERT", "UPDATE", "DELETE", "TRUNCATE"} {
				var granted bool
				if err := pool.QueryRow(ctx, `SELECT has_table_privilege($1,$2,$3)`, role, "public."+table, privilege).Scan(&granted); err != nil {
					t.Fatal(err)
				}
				want := role == "cartulary_recovery" || privilege == "SELECT" || privilege == "INSERT"
				if granted != want {
					t.Fatalf("preparation privilege %s/%s/%s=%v want %v", table, role, privilege, granted, want)
				}
			}
		}
	}

	temporaryRoot, publishedRoot := t.TempDir(), t.TempDir()
	storage, err := referenceassembly.NewRootStorage(temporaryRoot, publishedRoot)
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()
	var freshRows int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM reference_pack_current_set`).Scan(&freshRows); err != nil || freshRows != 0 {
		t.Fatal("fresh migration fabricated retained pack state", freshRows, err)
	}
	at := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	if err := reference_data.ReconcileBaseRelease(ctx, pool, storage, reference_data.BaseReleaseOptions{Limits: reference_data.DefaultLimits()}, at); err != nil {
		t.Fatal(err)
	}
	if err := reference_data.ValidateRequiredState(ctx, pool, storage, reference_data.DefaultLimits()); err != nil {
		t.Fatal("Base readiness validation", err)
	}
	if err := reference_data.RestoreHistoricalState(ctx, pool, storage, reference_data.DefaultLimits()); err != nil {
		t.Fatal("historical Base index rebuild", err)
	}
	t.Run("publication and backup retention guards", func(t *testing.T) { testPublicationCollectionAndBackupGuards(t, pool, storage) })
	consumer, err := reference_data.NewConsumer(pool, storage, pagination.NewCodec([32]byte([]byte(strings.Repeat("k", 32)))), func() time.Time { return at }, consumerIntegrityOptions(t))
	if err != nil {
		t.Fatal(err)
	}
	set := consumer.ResolveCurrentPackSet(ctx)
	if set.Error != nil || set.Value == nil || len(set.Value.Members) != 3 {
		t.Fatalf("Base set: %#v", set)
	}
	request := reference_data.GetPackProvenanceRequest{PackSetID: set.Value.ID, PackKey: "type_registry.host"}
	before := consumer.GetPackProvenance(ctx, request)
	if before.Error != nil {
		t.Fatal(before.Error)
	}
	var objects, envelopes int
	if err := pool.QueryRow(ctx, `SELECT (SELECT count(*) FROM reference_pack_objects),(SELECT count(*) FROM reference_pack_envelopes)`).Scan(&objects, &envelopes); err != nil {
		t.Fatal(err)
	}
	if err := reference_data.ReconcileBaseRelease(ctx, pool, storage, reference_data.BaseReleaseOptions{ProfileClaimed: true, Limits: reference_data.DefaultLimits()}, at.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	after := consumer.GetPackProvenance(ctx, request)
	if after.Error != nil || !reflect.DeepEqual(before.Value, after.Value) {
		t.Fatal("reconciliation rewrote first-success provenance")
	}
	var newObjects, newEnvelopes int
	if err := pool.QueryRow(ctx, `SELECT (SELECT count(*) FROM reference_pack_objects),(SELECT count(*) FROM reference_pack_envelopes)`).Scan(&newObjects, &newEnvelopes); err != nil {
		t.Fatal(err)
	}
	if objects != newObjects || envelopes != newEnvelopes {
		t.Fatal("restart duplicated immutable content or envelopes")
	}
	entry := consumer.GetPackEntry(ctx, reference_data.GetPackEntryRequest{PackSetID: set.Value.ID, PackKey: "type_registry.host", EntryID: "unknown"})
	if entry.Error != nil || entry.Value == nil {
		t.Fatal("Base entry not consumable")
	}
	t.Run("lost index is unavailable rather than a missing entry", func(t *testing.T) {
		var index uuid.UUID
		if err := pool.QueryRow(ctx, `SELECT current_index_id FROM reference_pack_candidates WHERE pack_key='type_registry.host'`).Scan(&index); err != nil {
			t.Fatal(err)
		}
		if _, err := pool.Exec(ctx, `UPDATE reference_pack_candidates SET current_index_id=NULL WHERE pack_key='type_registry.host'`); err != nil {
			t.Fatal(err)
		}
		defer func() {
			if _, err := pool.Exec(ctx, `UPDATE reference_pack_candidates SET current_index_id=$1 WHERE pack_key='type_registry.host'`, index); err != nil {
				t.Fatal(err)
			}
		}()
		entry := consumer.GetPackEntry(ctx, reference_data.GetPackEntryRequest{PackSetID: set.Value.ID, PackKey: "type_registry.host", EntryID: "unknown"})
		page := consumer.LookupPackEntries(ctx, reference_data.LookupPackEntriesRequest{PackSetID: set.Value.ID, PackKey: "type_registry.host", LookupKind: "entry_id", LookupValue: "unknown"})
		for _, err := range []*reference_data.ConsumerError{entry.Error, page.Error, consumer.GetPackProvenance(ctx, request).Error} {
			if err == nil || err.Code != "pack_unavailable" {
				t.Fatal("lost index did not fail explicitly", err)
			}
		}
	})
	t.Run("registry usage shares publication guards and transaction visibility", func(t *testing.T) {
		assignments, err := reference_data.NewRegistryAssignments(pool, storage, pagination.NewCodec([32]byte([]byte(strings.Repeat("r", 32)))), func() time.Time { return at }, consumerIntegrityOptions(t))
		if err != nil {
			t.Fatal(err)
		}
		readRevision := func(key string) int64 {
			t.Helper()
			var revision int64
			if err := pool.QueryRow(ctx, `SELECT revision FROM reference_pack_registry_usage WHERE pack_key=$1`, key).Scan(&revision); err != nil {
				t.Fatal(err)
			}
			return revision
		}
		beforeIndicator, beforeHost := readRevision("type_registry.indicator"), readRevision("type_registry.host")
		transactions, err := referenceassembly.NewRegistryTransactions(pool, assignments)
		if err != nil {
			t.Fatal(err)
		}
		write, err := transactions.BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			t.Fatal(err)
		}
		defer write.Rollback(ctx)
		// A competing publisher must wait for the same current-set guard. The
		// database timeout makes this assertion independent of goroutine timing.
		competing, err := pool.Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		defer competing.Rollback(ctx)
		if _, err := competing.Exec(ctx, `SET LOCAL lock_timeout='100ms'`); err != nil {
			t.Fatal(err)
		}
		_, err = competing.Exec(ctx, `SELECT 1 FROM reference_pack_current_set WHERE singleton FOR UPDATE`)
		var locked *pgconn.PgError
		if !errors.As(err, &locked) || locked.Code != "55P03" {
			t.Fatal("publication guard not held", err)
		}
		if err := competing.Rollback(ctx); err != nil {
			t.Fatal(err)
		}
		session, err := assignments.BeginTx(ctx, write)
		if err != nil {
			t.Fatal(err)
		}
		bound := session.ResolveCurrentPackSet(ctx)
		if bound.Error != nil || bound.Value.ID != set.Value.ID {
			t.Fatal("assignment did not resolve guarded set", bound.Error)
		}
		value := session.EvaluateIndicatorValue(ctx, reference_data.EvaluateIndicatorRequest{PackSetID: bound.Value.ID, IndicatorTypeID: "url", ValueKind: "atomic", RawValue: "https://EXAMPLE.COM/a/%2e%2e/b"})
		if value.Error != nil || !value.Value.Valid || *value.Value.Normalized != "https://example.com/b" {
			t.Fatal("transaction consumer evaluation", value.Error)
		}
		if err := session.RecordUsage(ctx, "type_registry.indicator"); err != nil {
			t.Fatal(err)
		}
		if err := session.RecordUsage(ctx, "type_registry.indicator"); err != nil {
			t.Fatal(err)
		}
		if session.RecordUsage(ctx, "undeclared.registry") == nil {
			t.Fatal("unknown registry usage admitted")
		}
		if readRevision("type_registry.indicator") != beforeIndicator {
			t.Fatal("uncommitted usage escaped")
		}
		if err := write.Rollback(ctx); err != nil {
			t.Fatal(err)
		}
		if readRevision("type_registry.indicator") != beforeIndicator {
			t.Fatal("rollback retained usage")
		}
		write, err = pool.Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		defer write.Rollback(ctx)
		session, err = assignments.BeginTx(ctx, write)
		if err != nil {
			t.Fatal(err)
		}
		for range 2 {
			if err := session.RecordUsage(ctx, "type_registry.indicator"); err != nil {
				t.Fatal(err)
			}
		}
		if err := write.Commit(ctx); err != nil {
			t.Fatal(err)
		}
		if readRevision("type_registry.indicator") != beforeIndicator+1 || readRevision("type_registry.host") != beforeHost {
			t.Fatal("usage advanced more than the affected registry")
		}
	})
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	pinned, err := reference_data.PinCurrentTx(ctx, tx, "test_snapshot", uuid.NewString(), uuid.New())
	if err != nil || pinned.ID != set.Value.ID {
		t.Fatal("snapshot admission did not capture exact set", err)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	retention, err := reference_data.NewRetention(pool, storage, func() time.Time { return at }, consumerIntegrityOptions(t))
	if err != nil {
		t.Fatal(err)
	}
	pinTx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	binding, err := retention.CaptureCurrentTx(ctx, pinTx, "test_report", uuid.NewString(), uuid.New())
	if err != nil {
		pinTx.Rollback(ctx)
		t.Fatal(err)
	}
	if err := pinTx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	if binding.SetID != pinned.ID || len(binding.Provenance) != 3 || retention.ValidateBinding(ctx, binding) != nil {
		t.Fatal("invalid immutable report binding")
	}
	t.Run("current backup restores exact historical pins and rebuilds indexes", func(t *testing.T) {
		sourcePacks, err := referenceassembly.NewRecoveryStorage(temporaryRoot, publishedRoot, reference_data.DefaultLimits())
		if err != nil {
			t.Fatal(err)
		}
		defer sourcePacks.Close()
		sourceObjects, err := objectstore.NewFilesystemStore(t.TempDir())
		if err != nil {
			t.Fatal(err)
		}
		defer sourceObjects.Close()
		rawBackup, err := recoveryassembly.NewFilesystemStorage(t.TempDir())
		if err != nil {
			t.Fatal(err)
		}
		key, err := recovery.ParseRecoveryEncryptionKey("MDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWY=")
		if err != nil {
			t.Fatal(err)
		}
		backup, err := recovery.NewEncryptedBackupStorage(rawBackup, key)
		if err != nil {
			t.Fatal(err)
		}
		streaming, err := recovery.RequireStreamingBackupStorage(backup)
		if err != nil {
			t.Fatal(err)
		}
		catalog, err := recoveryassembly.CurrentRecoveryStateCatalog()
		if err != nil {
			t.Fatal(err)
		}
		inventories, err := recoveryassembly.CurrentVNextObjectInventoryCatalog(recoveryassembly.NewVNextObjectSource(sourceObjects), sourcePacks, nil)
		if err != nil {
			t.Fatal(err)
		}
		capture, err := recovery.NewVNextCaptureService(recoveryassembly.NewVNextSnapshotRepository(pool), streaming, catalog, inventories)
		if err != nil {
			t.Fatal(err)
		}
		captured, err := capture.Capture(ctx, recovery.VNextCaptureParams{BackupSetID: uuid.New(), ConsistencyPointAt: at, CreatedAt: at, RetainedUntil: at.Add(31 * 24 * time.Hour)})
		if err != nil {
			t.Fatal(err)
		}
		backupStore := recovery.NewStore(pool)
		saved, err := backupStore.PublishVNextCapturedBackup(ctx, captured)
		if err != nil {
			t.Fatal(err)
		}
		targetDB := pgtest.Start(t).PrepareIsolatedDatabaseT(t, "reference-pack-restored-base")
		targetPool, err := pgxpool.New(ctx, targetDB.DSN)
		if err != nil {
			t.Fatal(err)
		}
		defer targetPool.Close()
		targetTemporary, targetPublished := t.TempDir(), t.TempDir()
		targetPacks, err := referenceassembly.NewRecoveryStorage(targetTemporary, targetPublished, reference_data.DefaultLimits())
		if err != nil {
			t.Fatal(err)
		}
		defer targetPacks.Close()
		targetObjects, err := objectstore.NewFilesystemStore(t.TempDir())
		if err != nil {
			t.Fatal(err)
		}
		defer targetObjects.Close()
		extensionCatalog, err := extensionassembly.GeneratedRecoveryCatalog()
		if err != nil {
			t.Fatal(err)
		}
		graph, err := recoveryassembly.NewGraphProjectionRestoreParticipant(targetPool)
		if err != nil {
			t.Fatal(err)
		}
		projections, _, err := projectionassembly.NewRecoveryServices(targetPool)
		if err != nil {
			t.Fatal(err)
		}
		_, err = recovery.NewVersionedRestoreRunner(backupStore, backup, extensionCatalog, catalog).RestoreBackupSet(ctx, recovery.RestoreTarget{RestoreOperationID: uuid.New(), TargetGenerationID: uuid.New(), Postgres: targetPool, ObjectStore: targetObjects, ReferencePacks: targetPacks, EvidenceObjects: evidence.NewRecoveryProvider(targetPool), GraphProjection: graph, Projections: projections}, saved)
		if err != nil {
			t.Fatal("restore retained Base", err)
		}
		restoredStorage, err := referenceassembly.NewRootStorage(targetTemporary, targetPublished)
		if err != nil {
			t.Fatal(err)
		}
		defer restoredStorage.Close()
		retained, err := reference_data.NewRetention(targetPool, restoredStorage, func() time.Time { return at }, consumerIntegrityOptions(t))
		if err != nil || retained.ValidateBinding(ctx, binding) != nil {
			t.Fatal("restored provenance/pin unavailable", err)
		}
		if err := reference_data.ValidateRequiredState(ctx, targetPool, restoredStorage, reference_data.DefaultLimits()); err != nil {
			t.Fatal(err)
		}
	})
	tampered := binding
	tampered.SHA256 = strings.Repeat("0", 64)
	if retention.ValidateBinding(ctx, tampered) == nil {
		t.Fatal("tampered binding admitted")
	}
	// A missing notice is as unavailable as a missing payload; counting remaining
	// members alone must not let a partial retained inventory pass.
	lossTx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := lossTx.Exec(ctx, `DELETE FROM reference_pack_object_refs WHERE owner_kind='version' AND logical_path='notices/LICENSE.txt'`); err != nil {
		lossTx.Rollback(ctx)
		t.Fatal(err)
	}
	if err := lossTx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	expected := referencetest.OwnerFixture(t, "historical_readiness")
	var unavailable *reference_data.ConsumerError
	if err := retention.ValidateBinding(ctx, binding); !errors.As(err, &unavailable) || unavailable.Code != expected.Error.Code {
		t.Fatal("incomplete retained inventory admitted", err)
	}
	var retainedSet *string
	if err := pool.QueryRow(ctx, `SELECT pack_set_id FROM reference_pack_current_set WHERE singleton`).Scan(&retainedSet); err != nil || retainedSet != nil {
		t.Fatal("definitive required loss retained an unusable selection", retainedSet, err)
	}
	var historicalDigest string
	if err := pool.QueryRow(ctx, `SELECT pack_set_sha256 FROM reference_pack_sets WHERE pack_set_id=$1`, "rpset_"+*expected.Set).Scan(&historicalDigest); err != nil || historicalDigest != *expected.Set {
		t.Fatal("required loss rewrote historical identity", historicalDigest, err)
	}
	if reference_data.ValidateRequiredState(ctx, pool, storage, reference_data.DefaultLimits()) == nil {
		t.Fatal("required missing notice passed readiness")
	}
	if _, err := pool.Exec(ctx, `UPDATE reference_pack_versions SET payload_sha256=repeat('0',64)`); err == nil {
		t.Fatal("immutable content update succeeded")
	}
	if _, err := pool.Exec(ctx, `DELETE FROM reference_pack_sets`); err == nil {
		t.Fatal("retained provenance deletion succeeded")
	}
}

func TestCanonicalCutoverRejectsLegacyStateBeforeMutation_Integration(t *testing.T) {
	t.Run("retired table removal refuses retained rows", testRetiredReferenceTableCutover)
	t.Run("unclassified fallback attribution refuses retained history", testFallbackAttributionCutover)
	ctx := context.Background()
	scratch := pgtest.Start(t).MigrationDatabaseThroughT(t, 45)
	db := scratch.SQL()
	_, err := db.ExecContext(ctx, `INSERT INTO reference_packs(pack_key,version,pack_kind,manifest_sha256,payload_sha256,pack_contract_version,verification_method,status,imported_at,verification_result,bundle_sha256,bundle_storage_ref,metadata) VALUES('type_registry.host','old','type_registry',repeat('a',64),repeat('b',64),'reference_pack.v1','manifest_sha256_v1','available',now(),'passed',repeat('c',64),'reference-packs/retained.bundle','{}'::jsonb)`)
	if err != nil {
		t.Fatal(err)
	}
	preflight := reference_data.PreflightCutover(ctx, db)
	if preflight == nil || !strings.Contains(preflight.Error(), "reference_pack_cutover_required") {
		t.Fatalf("missing safe cutover preflight report: %v", preflight)
	}
	if err := scratch.ApplyThrough(ctx, 46); err == nil {
		t.Fatal("legacy state passed cutover preflight")
	}
	var count, head int
	var relation *string
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM reference_packs WHERE version='old'`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT to_regclass('public.reference_pack_versions')::text`).Scan(&relation); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT max(version_id) FROM goose_db_version WHERE is_applied`).Scan(&head); err != nil {
		t.Fatal(err)
	}
	if count != 1 || head != 45 || relation != nil {
		t.Fatalf("preflight mutated retained state: count=%d head=%d relation=%v", count, head, relation)
	}
}

func testFallbackAttributionCutover(t *testing.T) {
	ctx := context.Background()
	scratch := pgtest.Start(t).MigrationDatabaseThroughT(t, 60)
	db := scratch.SQL()
	if err := reference_data.PreflightCutover(ctx, db); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `
INSERT INTO reference_pack_key_state(pack_key) VALUES('type_registry.host');
INSERT INTO reference_pack_candidates(pack_key,pack_version,distribution_kind,health,admitted_at) VALUES('type_registry.host','old-base','packaged_builtin','staged',now());
INSERT INTO reference_pack_operations(operation_id,kind,actor_kind,admitted_at,terminal_at,frozen_input,final_outcome) VALUES('63000000-0000-4000-8000-000000000001','reconcile','system',now(),now(),'{}','{}');
INSERT INTO reference_pack_events(attestation_id,operation_id,event_kind,pack_key,pack_version,canonical_attestation) VALUES('rpa_'||repeat('a',64),'63000000-0000-4000-8000-000000000001','safety_fallback','type_registry.host','old-base','{}');`); err != nil {
		t.Fatal(err)
	}
	var required *reference_data.CutoverRequiredError
	if err := reference_data.PreflightCutover(ctx, db); !errors.As(err, &required) || !strings.Contains(required.RemediationReportJSON(), `"category":"unclassified_safety_fallbacks","count":1`) {
		t.Fatal("missing fallback cutover category", err)
	}
	if err := scratch.ApplyThrough(ctx, 61); err == nil {
		t.Fatal("migration guessed current fallback attribution")
	}
	var head, count int
	var changed bool
	if err := db.QueryRowContext(ctx, `SELECT (SELECT max(version_id) FROM goose_db_version WHERE is_applied),(SELECT count(*) FROM reference_pack_events),EXISTS(SELECT 1 FROM pg_attribute WHERE attrelid='reference_pack_candidates'::regclass AND attname='fallback_from_version' AND NOT attisdropped)`).Scan(&head, &count, &changed); err != nil || head != 60 || count != 1 || changed {
		t.Fatal("rejected fallback cutover mutated retained state", head, count, changed, err)
	}
}

func TestCanonicalCutoverRejectsIntermediateIndicatorStateBeforeMutation_Integration(t *testing.T) {
	ctx := context.Background()
	scratch := pgtest.Start(t).MigrationDatabaseThroughT(t, 47)
	db := scratch.SQL()
	if err := reference_data.PreflightCutover(ctx, db); err != nil {
		t.Fatal("empty intermediate schema rejected", err)
	}
	_, err := db.ExecContext(ctx, `
INSERT INTO users(id,email,display_name,password_hash) VALUES ('57000000-0000-4000-8000-000000000001','cutover@example.test','Cutover','not-used');
INSERT INTO incidents(id,incident_key,incident_key_canonical,title,status,created_by_user_id,updated_by_user_id)
VALUES ('57000000-0000-4000-8000-000000000002','IR-CUTOVER','ir-cutover','Cutover','active','57000000-0000-4000-8000-000000000001','57000000-0000-4000-8000-000000000001');
INSERT INTO records(record_id,incident_id,record_type,created_by_user_id,updated_by_user_id,row_version)
VALUES ('57000000-0000-4000-8000-000000000003','57000000-0000-4000-8000-000000000002','indicator','57000000-0000-4000-8000-000000000001','57000000-0000-4000-8000-000000000001',1);
INSERT INTO indicators(record_id,incident_id,indicator_type,value_kind,display_value,normalized_value,dedupe_key)
VALUES ('57000000-0000-4000-8000-000000000003','57000000-0000-4000-8000-000000000002','domain_name','atomic','cutover.example','cutover.example',repeat('a',64));`)
	if err != nil {
		t.Fatal(err)
	}
	var required *reference_data.CutoverRequiredError
	if err := reference_data.PreflightCutover(ctx, db); !errors.As(err, &required) {
		t.Fatal("intermediate identities passed preflight", err)
	}
	if !strings.Contains(required.RemediationReportJSON(), `"category":"indicator_identities","count":1`) || strings.Contains(required.RemediationReportJSON(), "cutover.example") {
		t.Fatal("preflight must expose category counts only")
	}
	if err := scratch.ApplyThrough(ctx, 48); err == nil {
		t.Fatal("intermediate identities passed transactional guard")
	}
	var count, head int
	var current bool
	if err := db.QueryRowContext(ctx, `SELECT (SELECT count(*) FROM indicators),
(SELECT max(version_id) FROM goose_db_version WHERE is_applied),
EXISTS (SELECT 1 FROM pg_attribute WHERE attrelid='public.indicator_active_identities'::regclass AND attname='dedupe_sha256' AND NOT attisdropped)`).Scan(&count, &head, &current); err != nil {
		t.Fatal(err)
	}
	if count != 1 || head != 47 || current {
		t.Fatalf("rejected cutover mutated state: count=%d head=%d current=%v", count, head, current)
	}
}

func TestCanonicalCutoverRejectsUnboundImportedReferencesBeforeMutation_Integration(t *testing.T) {
	ctx := context.Background()
	scratch := pgtest.Start(t).MigrationDatabaseThroughT(t, 50)
	db := scratch.SQL()
	if err := reference_data.PreflightCutover(ctx, db); err != nil {
		t.Fatal("empty portability cutover rejected", err)
	}
	_, err := db.ExecContext(ctx, `
INSERT INTO users(id,email,display_name,password_hash) VALUES('58000000-0000-4000-8000-000000000001','reference-cutover@example.test','Cutover','not-used');
INSERT INTO incidents(id,incident_key,incident_key_canonical,title,status,created_by_user_id,updated_by_user_id)
VALUES('58000000-0000-4000-8000-000000000002','REFERENCE-CUTOVER','reference-cutover','Cutover','active','58000000-0000-4000-8000-000000000001','58000000-0000-4000-8000-000000000001');
INSERT INTO incident_bundle_imported_actors(incident_id,source_actor_id,local_user_id,display_name,email_hint)
VALUES('58000000-0000-4000-8000-000000000002','58000000-0000-4000-8000-000000000003','58000000-0000-4000-8000-000000000001','Historical actor','private@example.test');`)
	if err != nil {
		t.Fatal(err)
	}
	var required *reference_data.CutoverRequiredError
	if err := reference_data.PreflightCutover(ctx, db); !errors.As(err, &required) {
		t.Fatal("lost reference history passed preflight", err)
	}
	report := required.RemediationReportJSON()
	if !strings.Contains(report, `"category":"historical_incident_reference_bindings","count":1`) || strings.Contains(report, "private@example.test") {
		t.Fatal("preflight failed category/privacy contract")
	}
	if err := scratch.ApplyThrough(ctx, 51); err == nil {
		t.Fatal("migration fabricated imported reference history")
	}
	var count, head int
	var created bool
	if err := db.QueryRowContext(ctx, `SELECT (SELECT count(*) FROM incident_bundle_imported_actors),(SELECT max(version_id) FROM goose_db_version WHERE is_applied),to_regclass('public.reference_pack_portable_catalogs') IS NOT NULL`).Scan(&count, &head, &created); err != nil {
		t.Fatal(err)
	}
	if count != 1 || head != 50 || created {
		t.Fatalf("rejected cutover changed state: count=%d head=%d created=%v", count, head, created)
	}
}

func TestCanonicalCutoverRejectsUnboundDiagnosticsBeforeMutation_Integration(t *testing.T) {
	ctx := context.Background()
	scratch := pgtest.Start(t).MigrationDatabaseThroughT(t, 51)
	db := scratch.SQL()
	if err := reference_data.PreflightCutover(ctx, db); err != nil {
		t.Fatal(err)
	}
	_, err := db.ExecContext(ctx, `
INSERT INTO reference_pack_operations(operation_id,kind,actor_kind,admitted_at,frozen_input)
VALUES('59000000-0000-4000-8000-000000000001','import','local_operator',now(),'{}');
INSERT INTO reference_pack_attempts(attempt_id,operation_id,started_at)
VALUES('59000000-0000-4000-8000-000000000002','59000000-0000-4000-8000-000000000001',now());
INSERT INTO reference_pack_attempt_members(attempt_id,ordinal,verdict,failure_code,check_id)
VALUES('59000000-0000-4000-8000-000000000002',1,'content_rejected','content_schema_invalid','content_schema');`)
	if err != nil {
		t.Fatal(err)
	}
	var required *reference_data.CutoverRequiredError
	if err := reference_data.PreflightCutover(ctx, db); !errors.As(err, &required) {
		t.Fatal("diagnostic history passed preflight", err)
	}
	if !strings.Contains(required.RemediationReportJSON(), `"category":"historical_validation_summaries","count":1`) {
		t.Fatal("missing diagnostic category")
	}
	if err := scratch.ApplyThrough(ctx, 52); err == nil {
		t.Fatal("migration fabricated historical diagnostics")
	}
	var count, head int
	var created bool
	if err := db.QueryRowContext(ctx, `SELECT (SELECT count(*) FROM reference_pack_attempt_members),
(SELECT max(version_id) FROM goose_db_version WHERE is_applied),
EXISTS(SELECT 1 FROM pg_attribute WHERE attrelid='public.reference_pack_attempt_members'::regclass AND attname='canonical_validation_summary' AND NOT attisdropped)`).Scan(&count, &head, &created); err != nil {
		t.Fatal(err)
	}
	if count != 1 || head != 51 || created {
		t.Fatal("rejected diagnostic cutover mutated history")
	}
}

func TestCanonicalCutoverRejectsRetiredOperatorAttributionBeforeMutation_Integration(t *testing.T) {
	ctx := context.Background()
	scratch := pgtest.Start(t).MigrationDatabaseThroughT(t, 52)
	db := scratch.SQL()
	if err := reference_data.PreflightCutover(ctx, db); err != nil {
		t.Fatal(err)
	}
	_, err := db.ExecContext(ctx, `
 INSERT INTO users(id,email,display_name,password_hash) VALUES('60000000-0000-4000-8000-000000000001','operator-cutover@example.test','Cutover','not-used');
 INSERT INTO jobs(scope_kind,status,cancelable,submitted_by_user_id,submitted_at,updated_at,progress_completed,auth_policy,handler_name,job_kind,progress_unit_id)
 VALUES('deployment','queued',true,'60000000-0000-4000-8000-000000000001',now(),now(),0,'deployment_admin','reference_pack.lifecycle_v2','reference_pack.import_v2','reference_pack.import.request.v1');`)
	if err != nil {
		t.Fatal(err)
	}
	var required *reference_data.CutoverRequiredError
	if err := reference_data.PreflightCutover(ctx, db); !errors.As(err, &required) {
		t.Fatal("retired attribution passed preflight", err)
	}
	if !strings.Contains(required.RemediationReportJSON(), `"category":"retired_import_attribution","count":1`) {
		t.Fatal("missing attribution category")
	}
	if err := scratch.ApplyThrough(ctx, 53); err == nil {
		t.Fatal("migration fabricated local attribution")
	}
	var count, head int
	var changed bool
	if err := db.QueryRowContext(ctx, `SELECT (SELECT count(*) FROM jobs),(SELECT max(version_id) FROM goose_db_version WHERE is_applied), EXISTS(SELECT 1 FROM pg_attribute WHERE attrelid='public.jobs'::regclass AND attname='submitting_operator_operation_id' AND NOT attisdropped)`).Scan(&count, &head, &changed); err != nil {
		t.Fatal(err)
	}
	if count != 1 || head != 52 || changed {
		t.Fatal("preflight mutated schema or retained job")
	}
}

func TestCanonicalCutoverRejectsUnprovenDependencyStateBeforeMutation_Integration(t *testing.T) {
	t.Run("portable requirement classification", testPortableContentCutover)
	ctx := context.Background()
	for _, category := range []string{"pending_dependency_captures", "unproven_dependency_history"} {
		t.Run(category, func(t *testing.T) {
			scratch := pgtest.Start(t).MigrationDatabaseThroughT(t, 53)
			db := scratch.SQL()
			if err := reference_data.PreflightCutover(ctx, db); err != nil {
				t.Fatal(err)
			}
			query := `INSERT INTO reference_pack_operations(operation_id,kind,actor_kind,admitted_at,frozen_input) VALUES('61000000-0000-4000-8000-000000000001','refresh','system',now(),'{}')`
			if category == "unproven_dependency_history" {
				query = `INSERT INTO reference_pack_key_state(pack_key) VALUES('enrichment.lolbas');
 INSERT INTO reference_pack_candidates(pack_key,pack_version,distribution_kind,health,admitted_at) VALUES('enrichment.lolbas','old','operator_imported','failed',now());
 INSERT INTO reference_pack_versions(pack_key,pack_version,pack_release_sequence,manifest_sha256,payload_sha256,manifest_bytes) VALUES('enrichment.lolbas','old',1,repeat('a',64),repeat('b',64),'{"dependencies":[{"pack_key":"enrichment.tor","pack_version":"old","payload_sha256":"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"}]}');`
			}
			if _, err := db.ExecContext(ctx, query); err != nil {
				t.Fatal(err)
			}
			var required *reference_data.CutoverRequiredError
			if err := reference_data.PreflightCutover(ctx, db); !errors.As(err, &required) {
				t.Fatal("unproven dependency state passed preflight", err)
			}
			if !strings.Contains(required.RemediationReportJSON(), `"category":"`+category+`","count":1`) {
				t.Fatal("missing dependency category")
			}
			if err := scratch.ApplyThrough(ctx, 54); err == nil {
				t.Fatal("migration invented dependency evidence")
			}
			var head, count int
			var created bool
			if err := db.QueryRowContext(ctx, `SELECT (SELECT max(version_id) FROM goose_db_version WHERE is_applied),to_regclass('public.reference_pack_operation_dependency_keys') IS NOT NULL,(SELECT count(*) FROM reference_pack_operations)+(SELECT count(*) FROM reference_pack_versions)`).Scan(&head, &created, &count); err != nil {
				t.Fatal(err)
			}
			if head != 53 || created || count != 1 {
				t.Fatal("rejected preflight changed retained state")
			}
		})
	}
}

func testPortableContentCutover(t *testing.T) {
	ctx := context.Background()
	scratch := pgtest.Start(t).MigrationDatabaseThroughT(t, 54)
	db := scratch.SQL()
	if err := reference_data.PreflightCutover(ctx, db); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO reference_pack_operations(operation_id,kind,actor_kind,admitted_at,terminal_at,frozen_input,final_outcome) VALUES('62000000-0000-4000-8000-000000000001','portable_retention','system',now(),now(),'{}','{}')`); err != nil {
		t.Fatal(err)
	}
	var required *reference_data.CutoverRequiredError
	if err := reference_data.PreflightCutover(ctx, db); !errors.As(err, &required) {
		t.Fatal("missing portable content classification passed", err)
	}
	if !strings.Contains(required.RemediationReportJSON(), `"category":"unclassified_portable_content","count":1`) {
		t.Fatal("missing portable cutover category")
	}
	if err := scratch.ApplyThrough(ctx, 55); err == nil {
		t.Fatal("migration invented required-member classification")
	}
	var head, count int
	var changed bool
	if err := db.QueryRowContext(ctx, `SELECT (SELECT max(version_id) FROM goose_db_version WHERE is_applied),(SELECT count(*) FROM reference_pack_operations),EXISTS(SELECT 1 FROM pg_constraint WHERE conname='rp_portable_content_input_ck')`).Scan(&head, &count, &changed); err != nil {
		t.Fatal(err)
	}
	if head != 54 || count != 1 || changed {
		t.Fatal("rejected portable cutover mutated state")
	}
}
