package appsupport

import (
	verificationtest "github.com/JochiRaider/cartulary/internal/testutil/referencetest"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/app/configassembly"
	"github.com/JochiRaider/cartulary/internal/app/extensionassembly"
	"github.com/JochiRaider/cartulary/internal/app/projectionassembly"
	"github.com/JochiRaider/cartulary/internal/app/recoveryassembly"
	"github.com/JochiRaider/cartulary/internal/app/referenceassembly"
	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/configtest"
	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
	"github.com/JochiRaider/cartulary/internal/testutil/recoverytest"
	"github.com/JochiRaider/cartulary/internal/testutil/s3test"
	"github.com/google/uuid"
)

// CaptureRestoreServer composes shipped capture, fenced structural restore and
// server entry points against a fresh owned database and object-store namespace.
// It neither encodes snapshots nor resets an existing target. The caller stops
// source workers first and selects the claims for the restored application.
// Operator process behavior must be qualified separately.
func (runtime *Runtime) CaptureRestoreServer(t testing.TB, source *ServerHarness, options ServerOptions) *ServerHarness {
	t.Helper()
	ctx := t.Context()
	open := func(database *pgtest.TestDatabase) postgres.AdmittedPool {
		t.Helper()
		dsn, err := database.DSNForPurpose(postgres.PurposeRecovery)
		if err != nil {
			t.Fatal(err)
		}
		pool, err := postgres.Setup(ctx, postgres.Settings{BindingKind: "managed_service", DSN: dsn, Purpose: postgres.PurposeRecovery, ExpectedRole: "cartulary_recovery"})
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(pool.Close)
		return pool
	}
	sourcePool := open(source.Database)
	state, err := recoveryassembly.CurrentRecoveryStateCatalog()
	if err != nil {
		t.Fatal(err)
	}
	extensions, err := extensionassembly.GeneratedRecoveryCatalog()
	if err != nil {
		t.Fatal(err)
	}
	sourceRoots := source.Server.Config.Roots
	sourcePacks, err := referenceassembly.NewRecoveryStorage(sourceRoots.TemporaryWork.Path, sourceRoots.ReferencePackStorage.Path, reference_data.DefaultLimits(), verificationtest.Lifetime(t))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(sourcePacks.Close)
	inventory, err := recoveryassembly.CurrentVNextObjectInventoryCatalog(recoveryassembly.NewVNextObjectSource(source.ObjectStore), sourcePacks, nil)
	if err != nil {
		t.Fatal(err)
	}
	rawStorage, err := recoveryassembly.NewFilesystemStorage(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	key, err := recovery.ParseRecoveryEncryptionKey("MDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWY=")
	if err != nil {
		t.Fatal(err)
	}
	storage, err := recovery.NewEncryptedBackupStorage(rawStorage, key)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := recovery.CloseBackupStorage(storage); err != nil {
			t.Error(err)
		}
	})
	streaming, err := recovery.RequireStreamingBackupStorage(storage)
	if err != nil {
		t.Fatal(err)
	}
	capture, err := recovery.NewVNextCaptureService(recoveryassembly.NewVNextSnapshotRepository(sourcePool), streaming, state, inventory, extensions, recoverytest.ProofScratch(t))
	if err != nil {
		t.Fatal(err)
	}
	sourceBindings, err := configassembly.RecoveryTargetBindings(source.Server.Config, source.configurationEnv)
	if err != nil {
		t.Fatalf("bind current capture source: %v", err)
	}
	now := time.Now().UTC().Truncate(time.Second)
	captured, err := capture.Capture(ctx, recovery.VNextCaptureParams{SourceBindings: sourceBindings, BackupSetID: uuid.New(), ConsistencyPointAt: now, CreatedAt: now, RetainedUntil: now.Add(31 * 24 * time.Hour)})
	if err != nil {
		t.Fatalf("capture current backup: %v", err)
	}
	store := recovery.NewStore(sourcePool)
	backup, err := store.PublishVNextCapturedBackup(ctx, captured)
	if err != nil {
		t.Fatal(err)
	}

	targetDB := runtime.PrepareIsolatedDatabase(t, options.Prefix+"-target")
	targetPool := open(targetDB)
	targetObjects, objectEnv := prepareApplicationObjectStore(t, options.Prefix+"-target")
	env := configtest.SetupTempRoots(t).Paths
	env["CARTULARY__DEPLOYMENT_PROFILE"] = "on_prem"
	for k, v := range targetDB.EnvForServiceRef("restored_pg", postgres.PurposeRuntime) {
		env[k] = v
	}
	for k, v := range s3test.Start(t).EnvForServiceRef("restored_objects", objectEnv[objectstore.BucketEnv]) {
		env[k] = v
	}
	for k, v := range options.Env {
		env[k] = v
	}
	env["CARTULARY__ROOTS__DATABASE_STORAGE__BINDING_KIND"] = "managed_service"
	env["CARTULARY__ROOTS__DATABASE_STORAGE__PATH"] = ""
	env["CARTULARY__ROOTS__DATABASE_STORAGE__SERVICE_REF"] = "restored_pg"
	env["CARTULARY__ROOTS__OBJECT_STORAGE__BINDING_KIND"] = "managed_service"
	env["CARTULARY__ROOTS__OBJECT_STORAGE__PATH"] = ""
	env["CARTULARY__ROOTS__OBJECT_STORAGE__SERVICE_REF"] = "restored_objects"
	env["CARTULARY__REFERENCE_PACKS__TRUST_BOOTSTRAP_PATH"] = fixtures.Path("reference-packs", "trust-bootstrap.json")
	env["CARTULARY__BOOTSTRAP__FIRST_ADMIN_MANIFEST_PATH"] = fixtures.Path("bootstrap-admin", "canonical.json")
	loaded := configtest.LoadFixture(t, []string{"config", "valid.toml"}, env)
	bindings, err := configassembly.RecoveryTargetBindings(loaded.Deployment(), configtest.EffectiveConfigEnv([]string{"config", "valid.toml"}, env))
	if err != nil {
		t.Fatalf("bind current restore target: %v", err)
	}
	roots := loaded.Deployment().Roots
	targetPacks, err := referenceassembly.NewRecoveryStorage(roots.TemporaryWork.Path, roots.ReferencePackStorage.Path, reference_data.DefaultLimits(), verificationtest.Lifetime(t))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(targetPacks.Close)
	core, err := projectionassembly.Build(targetPool)
	if err != nil {
		t.Fatal(err)
	}
	target := recovery.RestoreTarget{Postgres: targetPool, ObjectStore: targetObjects, ReferencePacks: targetPacks, ProofScratch: recoverytest.ProofScratch(t), EvidenceObjects: evidence.NewRecoveryProvider(targetPool), Projections: core.RecoveryPorts().Rebuilder}
	_, err = RestoreCurrentBackup(ctx, CurrentRestoreParams{
		Source: sourcePool, Target: target,
		Runner:  recovery.NewVersionedRestoreRunner(store, storage, extensions, state),
		Catalog: recovery.NewBackupCatalog(store, storage, extensions, state),
		Backup:  backup, Key: key, Bindings: bindings,
	})
	if err != nil {
		t.Fatalf("restore current backup: %v", err)
	}
	options.Database, options.ObjectStore, options.Env, options.loaded = targetDB, targetObjects, env, &loaded
	return runtime.StartServer(t, options)
}
