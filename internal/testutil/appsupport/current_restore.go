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
	recoverystate "github.com/JochiRaider/cartulary/internal/platform/recoverystate"
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
func (runtime *Runtime) CaptureRestoreServer(t testing.TB, source *ServerHarness, options ServerOptions, probes ...CurrentRestoreProbe) *ServerHarness {
	t.Helper()
	return runtime.CaptureCurrent(t, source).RestoreServer(t, options, probes...)
}

// CurrentCapture is one published current backup of a source server. Any
// number of fresh, isolated targets restore from it, so callers that need
// several targets of the same source state capture once. Its source resources
// live until the capturing test ends.
type CurrentCapture struct {
	runtime    *Runtime
	sourcePool postgres.AdmittedPool
	store      *recovery.Store
	storage    recovery.BackupStorage
	extensions *recovery.ExtensionBackupCatalog
	state      *recoverystate.Catalog
	backup     recovery.BackupSet
	key        recovery.RecoveryEncryptionKey
}

func openRecoveryPool(t testing.TB, database *pgtest.TestDatabase) postgres.AdmittedPool {
	t.Helper()
	dsn, err := database.DSNForPurpose(postgres.PurposeRecovery)
	if err != nil {
		t.Fatal(err)
	}
	pool, err := postgres.Setup(t.Context(), postgres.Settings{BindingKind: "managed_service", DSN: dsn, Purpose: postgres.PurposeRecovery, ExpectedRole: "cartulary_recovery"})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	return pool
}

// CaptureCurrent captures and publishes the source server's current backup.
func (runtime *Runtime) CaptureCurrent(t testing.TB, source *ServerHarness) *CurrentCapture {
	t.Helper()
	ctx := t.Context()
	sourcePool := openRecoveryPool(t, source.Database)
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
	return &CurrentCapture{runtime: runtime, sourcePool: sourcePool, store: store, storage: storage, extensions: extensions, state: state, backup: backup, key: key}
}

// RestoreServer restores the capture into a fresh owned database and
// object-store namespace and starts a server on it. Each probe first receives
// its own fresh target of the same captured bytes.
func (capture *CurrentCapture) RestoreServer(t testing.TB, options ServerOptions, probes ...CurrentRestoreProbe) *ServerHarness {
	t.Helper()
	runtime := capture.runtime
	prepareTarget := func(selected ServerOptions) (ServerOptions, CurrentRestoreParams) {
		targetDB := runtime.PrepareIsolatedDatabase(t, selected.Prefix+"-target")
		targetPool := openRecoveryPool(t, targetDB)
		targetObjects, objectEnv := prepareApplicationObjectStore(t, selected.Prefix+"-target")
		env := configtest.SetupTempRoots(t).Paths
		env["CARTULARY__DEPLOYMENT_PROFILE"] = "on_prem"
		for k, v := range targetDB.EnvForServiceRef("restored_pg", postgres.PurposeRuntime) {
			env[k] = v
		}
		for k, v := range s3test.Start(t).EnvForServiceRef("restored_objects", objectEnv[objectstore.BucketEnv]) {
			env[k] = v
		}
		for k, v := range selected.Env {
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
		params := CurrentRestoreParams{
			Source: capture.sourcePool, Target: target,
			Runner:  recovery.NewVersionedRestoreRunner(capture.store, capture.storage, capture.extensions, capture.state),
			Catalog: recovery.NewBackupCatalog(capture.store, capture.storage, capture.extensions, capture.state),
			Backup:  capture.backup, Key: capture.key, Bindings: bindings,
		}
		selected.Database, selected.ObjectStore, selected.Env, selected.loaded = targetDB, targetObjects, env, &loaded
		return selected, params
	}
	for _, probe := range probes {
		selected := options
		selected.Prefix += "-" + probe.Name
		_, params := prepareTarget(selected)
		probe.Run(t, params)
	}
	options, params := prepareTarget(options)
	if _, err := RestoreCurrentBackup(t.Context(), params); err != nil {
		t.Fatalf("restore current backup: %v", err)
	}
	return runtime.StartServer(t, options)
}

// CurrentRestoreProbe receives a separate fresh target of the same captured
// bytes. Probes can inject faults at participant boundaries, never reuse/reset
// the successful target or manufacture recovery evidence.
type CurrentRestoreProbe struct {
	Name string
	Run  func(testing.TB, CurrentRestoreParams)
}
