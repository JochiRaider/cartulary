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
	t.Run("historical binding in consumer read transaction", func(t *testing.T) {
		read, err := pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
		if err != nil {
			t.Fatal(err)
		}
		defer read.Rollback(ctx)
		if err := retention.ValidateBindingTx(ctx, read, binding); err != nil {
			t.Fatal(err)
		}
		altered := binding
		altered.SHA256 = strings.Repeat("0", 64)
		if retention.ValidateBindingTx(ctx, read, altered) == nil {
			t.Fatal("altered historical binding admitted")
		}
		if retention.ValidateBindingTx(ctx, nil, binding) == nil {
			t.Fatal("missing read transaction admitted")
		}
		if err := read.Commit(ctx); err != nil {
			t.Fatal("validation changed or aborted read-only transaction", err)
		}
	})
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
		restoredRead, err := targetPool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
		if err != nil {
			t.Fatal(err)
		}
		if err := retained.ValidateBindingTx(ctx, restoredRead, binding); err != nil {
			_ = restoredRead.Rollback(ctx)
			t.Fatal("restored historical binding rejected", err)
		}
		if err := restoredRead.Commit(ctx); err != nil {
			t.Fatal(err)
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
