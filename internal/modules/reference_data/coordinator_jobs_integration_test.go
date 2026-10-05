package reference_data

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packformat"
	"github.com/JochiRaider/cartulary/internal/platform/extensionstore"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type coordinatorTestFinalizer struct {
	owner *extensionstore.OwnerFinalizer
}

func (f coordinatorTestFinalizer) FinalizeReferencePackAction(ctx context.Context, tx pgx.Tx, proof func(context.Context) (bool, error)) error {
	return f.owner.CommitOwnerMutation(ctx, tx, proof)
}

func (f coordinatorTestFinalizer) FinalizeReferencePackJobTimeout(ctx context.Context, r JobFailureFinalization) (jobs.Resource, error) {
	return f.owner.FinalizeTimeout(ctx, extensionstore.JobFailureFinalizationRequest{Execution: r.Execution, Completion: r.Completion, Mutate: extensionstore.OwnerMutation(r.Mutate)})
}

func (f coordinatorTestFinalizer) FinalizeReferencePackJobSuccess(ctx context.Context, r JobSuccessFinalization) (jobs.Resource, error) {
	return f.owner.FinalizeSuccess(ctx, extensionstore.JobFinalizationRequest{Execution: r.Execution, Completion: r.Completion, FinalCommitID: r.FinalCommitID, Mutate: extensionstore.OwnerMutation(r.Mutate)})
}
func (f coordinatorTestFinalizer) FinalizeReferencePackJobFailure(ctx context.Context, r JobFailureFinalization) (jobs.Resource, error) {
	return f.owner.FinalizeFailure(ctx, extensionstore.JobFailureFinalizationRequest{Execution: r.Execution, Completion: r.Completion, Mutate: extensionstore.OwnerMutation(r.Mutate)})
}
func (f coordinatorTestFinalizer) FinalizeReferencePackJobCancellation(ctx context.Context, r JobCancellationFinalization) (jobs.Resource, error) {
	return f.owner.FinalizeCancellation(ctx, extensionstore.JobCancellationFinalizationRequest{Execution: r.Execution, Completion: r.Completion, Mutate: extensionstore.OwnerMutation(r.Mutate)})
}

func TestCanonicalCoordinatorLeasedJobsAndTerminalReplay_Integration(t *testing.T) {
	f := newCanonicalCoordinatorFixture(t, "reference-pack-coordinator-jobs")
	ctx := context.Background()
	c, pool, manager, storage, actor := f.coordinator, f.pool, f.manager, f.storage, f.actor
	catalog, transactions, definitions := f.catalog, f.transactions, f.definitions
	vector, container, clock := f.vector, f.container, f.now
	t.Run("retained terminal Job startup repair", func(t *testing.T) {
		testRetainedTerminalJobRepair(t, c, pool, catalog, transactions, definitions, actor, container)
	})
	t.Run("initial Job recovery", func(t *testing.T) {
		testJobTerminalOwnerRecovery(t, c, pool, catalog, transactions, definitions, actor, container, false)
	})
	initial, err := c.Import(ctx, actor, "initial-canceled", bytes.NewReader(container))
	if err != nil {
		t.Fatal(err)
	}
	initialID := uuid.MustParse(initial.Job.JobID)
	if _, err := manager.Cancel(ctx, jobs.CancelParams{JobID: initialID, ActorUserID: actor, ClientTxnID: "cancel-initial", NormalizedRequest: []byte(`{"client_txn_id":"cancel-initial"}`)}); err != nil {
		t.Fatal(err)
	}
	initialExecution, claimed, err := manager.Claim(ctx, initialID)
	if err != nil || !claimed {
		t.Fatal("claim initial cancellation", claimed, err)
	}
	if err := c.Execute(ctx, initialExecution); err != nil {
		t.Fatal(err)
	}
	var initialHealth string
	var initialEnvelope *string
	if err := pool.QueryRow(ctx, `SELECT health,current_envelope_id FROM reference_pack_candidates WHERE pack_key='type_registry.host' AND pack_version='signed-fixture.1'`).Scan(&initialHealth, &initialEnvelope); err != nil || initialHealth != "failed" || initialEnvelope != nil {
		t.Fatal("initial abort fabricated success or remained staged", initialHealth, initialEnvelope, err)
	}
	accepted, err := c.Import(ctx, actor, "signed-import", bytes.NewReader(container))
	if err != nil {
		t.Fatal(err)
	}
	jobID := uuid.MustParse(accepted.Job.JobID)
	var retainedInputs int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM reference_pack_object_refs r JOIN reference_pack_objects o USING(object_id) JOIN reference_pack_operations p ON r.owner_id=p.operation_id::text WHERE p.job_id=$1 AND r.owner_kind='operation' AND r.logical_path='input/container' AND o.size_bytes=$2`, jobID, len(container)).Scan(&retainedInputs); err != nil || retainedInputs != 1 || len(storage.staged) != 0 {
		t.Fatal("queued input missing from retained recovery inventory", retainedInputs, err)
	}
	queuedVersion, err := c.GetVersion(ctx, "type_registry.host", "signed-fixture.1")
	if err != nil || !queuedVersion.PendingWork || queuedVersion.Dependencies != nil {
		t.Fatal("queued work missing from resource", queuedVersion, err)
	}
	execution, claimed, err := manager.Claim(ctx, jobID)
	if err != nil || !claimed {
		t.Fatal("claim", claimed, err)
	}
	if err := c.Execute(ctx, execution); err != nil {
		t.Fatal(err)
	}
	job, err := manager.Get(ctx, jobID)
	if err != nil || job.Status != jobs.StatusSucceeded {
		t.Fatal("signed import Job", job.Status, job.ErrorSummary, err)
	}
	if job.ResultSummary == nil || len(job.ResultSummary.ResourceRefs) != 1 || job.ResultSummary.ResourceRefs[0].ID != "/api/v1/reference-packs/type_registry.host/signed-fixture.1" {
		t.Fatal("successful import lost exact navigation resource", job.ResultSummary)
	}
	version, err := c.GetVersion(ctx, "type_registry.host", "signed-fixture.1")
	if err != nil || version.Health != "verified_available" || version.PendingWork || version.ManifestSHA256 == nil || version.LastVerifiedAt == nil || len(version.VerifiedSignerKeyIDs) != 1 || version.Dependencies == nil || len(version.Dependencies) != 0 {
		t.Fatal("administrative successful envelope projection", version, err)
	}
	action, apiErr := DecodeActionRequest(strings.NewReader(`{"client_txn_id":"disable-signed"}`))
	if apiErr != nil {
		t.Fatal(apiErr)
	}
	params := ActionParams{ActorUserID: actor, PackKey: "type_registry.host", PackVersion: "signed-fixture.1", Request: action, Now: *clock}
	disabled, err := c.Disable(ctx, params)
	if err != nil || disabled.Version.Condition != "disabled" || disabled.Version.Health != "verified_available" {
		t.Fatal("disable did not preserve verification health", disabled, err)
	}
	again, err := c.Disable(ctx, params)
	if err != nil || !again.Replayed {
		t.Fatal("disable replay re-executed state conflict", again, err)
	}
	t.Run("established Job recovery", func(t *testing.T) {
		testJobTerminalOwnerRecovery(t, c, pool, catalog, transactions, definitions, actor, container, true)
	})
	t.Run("retained attempt integrity", func(t *testing.T) { testHistoricalAttemptCorruption(t, c) })
	t.Run("local operator import shares verification finalization and preserves disablement", func(t *testing.T) {
		operation := uuid.New()
		pending, err := c.PrepareImport(ctx, bytes.NewReader(container))
		if err != nil {
			t.Fatal(err)
		}
		defer pending.Close()
		accepted, err := pending.AcceptLocalOperator(ctx, operation)
		if err != nil || accepted.Job.SubmittedByUserID != nil {
			t.Fatal("local admission", accepted, err)
		}
		id := uuid.MustParse(accepted.Job.JobID)
		execution, claimed, err := manager.Claim(ctx, id)
		if err != nil || !claimed {
			t.Fatal("local claim", claimed, err)
		}
		if err := c.Execute(ctx, execution); err != nil {
			t.Fatal(err)
		}
		job, err := manager.Get(ctx, id)
		if err != nil || job.Status != jobs.StatusSucceeded || job.SubmittedByUserID != nil {
			t.Fatal("local finalization", job, err)
		}
		version, err := c.GetVersion(ctx, "type_registry.host", "signed-fixture.1")
		if err != nil || !version.AdministrativelyDisabled || version.Active {
			t.Fatal("local renewal changed administration", version, err)
		}
		var count int
		if err := pool.QueryRow(ctx, `SELECT count(*) FROM reference_pack_operations o JOIN reference_pack_events e USING(operation_id)
WHERE o.operation_id=$1 AND o.job_id=$2 AND o.actor_kind='local_operator' AND o.actor_user_id IS NULL
AND convert_from(e.canonical_attestation,'UTF8')::jsonb->>'operator_operation_id'=$1::text`, operation, id).Scan(&count); err != nil || count != 1 {
			t.Fatal("local attestation", count, err)
		}
		if err := pool.QueryRow(ctx, `SELECT count(*) FROM administrative_audit_projections WHERE target_id=$1 AND actor_kind='operator' AND actor_user_id IS NULL AND source='operator'`, operation.String()).Scan(&count); err != nil || count != 3 {
			t.Fatal("local audit", count, err)
		}
		if err := pool.QueryRow(ctx, `SELECT count(*) FROM extension_job_commit_proofs WHERE job_id=$1`, id).Scan(&count); err != nil || count != 1 {
			t.Fatal("local success proof", count, err)
		}
	})
	t.Run("historical retention preserves activation and serializes removal", func(t *testing.T) {
		repository := canonicalRepository{pool: pool, storage: storage}
		current, err := repository.CurrentSet(ctx)
		if err != nil {
			t.Fatal(err)
		}
		var revision int64
		if err := pool.QueryRow(ctx, `SELECT revision FROM reference_pack_current_set WHERE singleton`).Scan(&revision); err != nil {
			t.Fatal(err)
		}
		members := append([]PackSetMember(nil), current.Members...)
		for index := range members {
			if members[index].Key == version.PackKey {
				members[index].Version = version.PackVersion
				members[index].ManifestSHA256 = *version.ManifestSHA256
				members[index].PayloadSHA256 = *version.PayloadSHA256
			}
		}
		var operation uuid.UUID
		if err := pool.QueryRow(ctx, `SELECT operation_id FROM reference_pack_operations WHERE job_id=$1`, jobID).Scan(&operation); err != nil {
			t.Fatal(err)
		}
		tx, err := pool.Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		defer tx.Rollback(ctx)
		for _, member := range members {
			if _, err := tx.Exec(ctx, `SELECT 1 FROM reference_pack_key_state WHERE pack_key=$1 FOR UPDATE`, member.Key); err != nil {
				t.Fatal(err)
			}
		}
		historical, err := retainSetTx(ctx, tx, members, operation)
		if err != nil || historical.ID == current.ID {
			t.Fatal("retain disabled historical set", err)
		}
		if err := tx.Commit(ctx); err != nil {
			t.Fatal(err)
		}
		var unchanged bool
		if err := pool.QueryRow(ctx, `SELECT pack_set_id=$1 AND revision=$2 FROM reference_pack_current_set WHERE singleton`, current.ID, revision).Scan(&unchanged); err != nil || !unchanged {
			t.Fatal("retaining historical set changed activation", err)
		}
		integrity, err := newIntegrityService(c.pool, c.storage, c.now, IntegrityOptions{Finalizer: c.actionFinalizer, Limits: c.limits})
		if err != nil {
			t.Fatal(err)
		}
		retained := &retention{repository: &repository, integrity: integrity}
		tx, err = pool.Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		defer tx.Rollback(ctx)
		binding, err := retained.CaptureRetainedTx(ctx, tx, historical.ID, "portable_snapshot", uuid.NewString(), uuid.New())
		if err != nil || binding.SetID != historical.ID || len(binding.Provenance) != 3 {
			t.Fatal("pin historical disabled set", err)
		}
		pinnedResource, err := scanAdministrativeVersion(tx.QueryRow(ctx, administrativeVersionSelect+` WHERE c.pack_key=$1 AND c.pack_version=$2`, version.PackKey, version.PackVersion))
		if err != nil || !pinnedResource.ReproducibilityPinned {
			t.Fatal("pin missing from administrative projection", pinnedResource, err)
		}
		catalog, err := retained.ExportReferencesTx(ctx, tx, []string{historical.ID, current.ID, historical.ID})
		if err != nil {
			t.Fatal(err)
		}
		refs, err := DecodeIncidentBundleReferences(catalog)
		if err != nil || len(refs.Sets) != 2 || len(refs.Versions) != 4 {
			t.Fatal("historical catalog lost exact references or duplicated versions", refs, err)
		}
		empty, err := retained.ExportReferencesTx(ctx, tx, nil)
		if err != nil || string(empty) != "{\"schema_id\":\"reference_pack_refs.v1\",\"sets\":[],\"versions\":[]}\n" {
			t.Fatal("empty reference catalog", string(empty), err)
		}
		competing, err := pool.Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		defer competing.Rollback(ctx)
		if _, err := competing.Exec(ctx, `SET LOCAL lock_timeout='100ms'`); err != nil {
			t.Fatal(err)
		}
		_, err = competing.Exec(ctx, `SELECT 1 FROM reference_pack_key_state WHERE pack_key='type_registry.host' FOR UPDATE`)
		var locked *pgconn.PgError
		if !errors.As(err, &locked) || locked.Code != "55P03" {
			t.Fatal("removal did not serialize with exact-set pin", err)
		}
		if err := tx.Rollback(ctx); err != nil {
			t.Fatal(err)
		}
		v, err := c.GetVersion(ctx, version.PackKey, version.PackVersion)
		if err != nil || !v.AdministrativelyDisabled || v.Active {
			t.Fatal("historical pin changed administrative state", v, err)
		}
		after, err := repository.CurrentSet(ctx)
		if err != nil || !reflect.DeepEqual(after, current) {
			t.Fatal("historical export rebased current selection", err)
		}
	})
	retainedObjects := len(storage.objects)
	replay, err := c.Import(ctx, actor, "signed-import", bytes.NewReader(container))
	if err != nil || !replay.Replayed || replay.Job.Status != jobs.StatusSucceeded || replay.Job.JobID != accepted.Job.JobID {
		t.Fatal("terminal replay", replay, err)
	}
	if len(storage.staged) != 0 || len(storage.objects) != retainedObjects {
		t.Fatal("replay retained duplicate staging")
	}
	// The operation is admitted while metadata is fresh, then waits in Jobs.
	queued, err := c.VerifyRetained(ctx, VerificationRequest{Kind: "reverify", ActorUserID: actor, ClientTxnID: "expires-in-queue", PackKeys: []string{"type_registry.host"}, KeysProvided: true, PackVersion: "signed-fixture.1"})
	if err != nil {
		t.Fatal(err)
	}
	*clock = vector.Expiry
	execution, claimed, err = manager.Claim(ctx, uuid.MustParse(queued.Job.JobID))
	if err != nil || !claimed {
		t.Fatal("claim expired", claimed, err)
	}
	if err := c.Execute(ctx, execution); err != nil {
		t.Fatal(err)
	}
	job, err = manager.Get(ctx, uuid.MustParse(queued.Job.JobID))
	if err != nil || job.Status != jobs.StatusFailed {
		t.Fatal("queue expiry did not fail", job, err)
	}
	var health string
	var terminal string
	if err := pool.QueryRow(ctx, `SELECT health FROM reference_pack_candidates WHERE pack_key='type_registry.host' AND pack_version='signed-fixture.1'`).Scan(&health); err != nil || health != "failed" {
		t.Fatal("failed verdict not atomically published", health, err)
	}
	// Restore authenticates the historical success at its original instant even
	// though the fresh reverify just condemned its expired metadata. Rebuilding
	// a derived index must not repair or overwrite the live verdict.
	if err := RestoreHistoricalState(ctx, pool, storage, DefaultLimits()); err != nil {
		t.Fatal("historically valid expired metadata failed restore", err)
	}
	if err := pool.QueryRow(ctx, `SELECT health FROM reference_pack_candidates WHERE pack_key='type_registry.host' AND pack_version='signed-fixture.1'`).Scan(&health); err != nil || health != "failed" {
		t.Fatal("restore changed established health", health, err)
	}
	if err := pool.QueryRow(ctx, `SELECT response_json->>'status' FROM route_idempotency WHERE client_txn_id='expires-in-queue'`).Scan(&terminal); err != nil || terminal != jobs.StatusFailed {
		t.Fatal("failed terminal receipt", terminal, err)
	}
	// Cancellation clears the operation boundary without condemning its old
	// successful envelope. A later operation can be admitted on this pack key.
	*clock = vector.At
	cancelJob, err := c.VerifyRetained(ctx, VerificationRequest{Kind: "reverify", ActorUserID: actor, ClientTxnID: "cancel-before-start", PackKeys: []string{"type_registry.host"}, KeysProvided: true, PackVersion: "signed-fixture.1"})
	if err != nil {
		t.Fatal(err)
	}
	cancelID := uuid.MustParse(cancelJob.Job.JobID)
	if _, err := manager.Cancel(ctx, jobs.CancelParams{JobID: cancelID, ActorUserID: actor, ClientTxnID: "cancel-it", NormalizedRequest: []byte(`{"client_txn_id":"cancel-it"}`)}); err != nil {
		t.Fatal(err)
	}
	execution, claimed, err = manager.Claim(ctx, cancelID)
	if err != nil || !claimed {
		t.Fatal("claim canceled", claimed, err)
	}
	if err := c.Execute(ctx, execution); err != nil {
		t.Fatal(err)
	}
	job, err = manager.Get(ctx, cancelID)
	if err != nil || job.Status != jobs.StatusCanceled {
		t.Fatal("cancellation", job, err)
	}
	refresh, err := c.VerifyRetained(ctx, VerificationRequest{Kind: "refresh", ActorUserID: actor, ClientTxnID: "after-cancel"})
	if err != nil {
		t.Fatal("terminal cancellation retained pending-work guard", err)
	}
	removeRequest, apiErr := DecodeActionRequest(strings.NewReader(`{"client_txn_id":"remove-signed","reason":"Fixture retirement"}`))
	if apiErr != nil {
		t.Fatal(apiErr)
	}
	params.Request = removeRequest
	if _, err := c.Remove(ctx, params); err == nil {
		t.Fatal("removal raced admitted refresh")
	}
	execution, claimed, err = manager.Claim(ctx, uuid.MustParse(refresh.Job.JobID))
	if err != nil || !claimed {
		t.Fatal("claim refresh", claimed, err)
	}
	if err := c.Execute(ctx, execution); err != nil {
		t.Fatal(err)
	}
	version, err = c.GetVersion(ctx, "type_registry.host", "signed-fixture.1")
	if err != nil || version.Condition != "disabled" || !version.AdministrativelyDisabled {
		t.Fatal("refresh cleared administrative disablement", version, err)
	}
	removed, err := c.Remove(ctx, params)
	if err != nil || (removed.Version.MissingReason == nil || *removed.Version.MissingReason != "administrative_removal") || removed.Version.ManifestSHA256 == nil {
		t.Fatal("removal lost history or failed", removed, err)
	}
	if _, err := c.VerifyRetained(ctx, VerificationRequest{Kind: "reverify", ActorUserID: actor, ClientTxnID: "removed-reverify", PackKeys: []string{"type_registry.host"}, KeysProvided: true, PackVersion: "signed-fixture.1"}); err == nil {
		t.Fatal("removed content entered reverify")
	}
	t.Run("removed extraction collection and historical backup retention", func(t *testing.T) {
		testRemovedObjectCollection(t, c, storage, "type_registry.host", "signed-fixture.1")
	})
	reimport, err := c.Import(ctx, actor, "restore-exact", bytes.NewReader(container))
	if err != nil {
		t.Fatal(err)
	}
	if err := CollectUnreferencedObjects(ctx, pool, storage); err != nil {
		t.Fatal("pending reimport collection", err)
	}
	execution, claimed, err = manager.Claim(ctx, uuid.MustParse(reimport.Job.JobID))
	if err != nil || !claimed {
		t.Fatal("claim reimport", claimed, err)
	}
	if err := c.Execute(ctx, execution); err != nil {
		t.Fatal(err)
	}
	version, err = c.GetVersion(ctx, "type_registry.host", "signed-fixture.1")
	if err != nil || version.Removed || version.MissingReason != nil || version.Condition != "disabled" {
		t.Fatal("exact reimport changed identity or disablement", version, err)
	}
	if err := CollectUnreferencedObjects(ctx, pool, storage); err != nil {
		t.Fatal("post-reimport collection", err)
	}
	if err := checkRetainedVersion(ctx, pool, storage, "type_registry.host", "signed-fixture.1"); err != nil {
		t.Fatal("stale collection deleted restored content", err)
	}
	var removalCount int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM reference_pack_events WHERE event_kind='removal' AND pack_key='type_registry.host' AND pack_version='signed-fixture.1'`).Scan(&removalCount); err != nil || removalCount != 1 {
		t.Fatal("reimport discarded removal history", removalCount, err)
	}
	empty, err := c.Import(ctx, actor, "empty-container", bytes.NewReader(nil))
	if err != nil {
		t.Fatal(err)
	}
	emptyID := uuid.MustParse(empty.Job.JobID)
	execution, claimed, err = manager.Claim(ctx, emptyID)
	if err != nil || !claimed {
		t.Fatal("claim empty container", claimed, err)
	}
	if err := c.Execute(ctx, execution); err != nil {
		t.Fatal(err)
	}
	job, err = manager.Get(ctx, emptyID)
	if err != nil || job.Status != jobs.StatusFailed || job.ErrorSummary == nil || job.ErrorSummary.Details["reason_code"] != "unsupported_container_format" {
		t.Fatal("empty container lost its content verdict", job, err)
	}
}

type signedCoordinatorVector struct {
	Bootstrap  string    `json:"bootstrap"`
	Repository string    `json:"repository_id"`
	Container  string    `json:"container_base64"`
	At         time.Time `json:"verification_time"`
	Expiry     time.Time `json:"expected_valid_until"`
}

type canonicalCoordinatorFixture struct {
	coordinator  *Coordinator
	pool         *pgxpool.Pool
	manager      *jobs.Manager
	storage      *coordinatorMemoryStorage
	actor        uuid.UUID
	catalog      *jobs.Catalog
	transactions *jobs.TransactionService
	definitions  []jobs.Definition
	vector       signedCoordinatorVector
	container    []byte
	now          *time.Time
}

func newCanonicalCoordinatorFixture(t *testing.T, name string) canonicalCoordinatorFixture {
	t.Helper()
	ctx := context.Background()
	db := pgtest.Start(t).PrepareIsolatedDatabaseT(t, name)
	pool, err := pgxpool.New(ctx, db.DSN)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	var vector signedCoordinatorVector
	data, err := os.ReadFile("../../../contracts/reference-pack-fixtures/fixtures/signed-container.v1.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &vector); err != nil {
		t.Fatal(err)
	}
	container, err := base64.StdEncoding.DecodeString(vector.Container)
	if err != nil {
		t.Fatal(err)
	}
	root, err := packformat.AdmitBootstrap([]byte(vector.Bootstrap))
	if err != nil {
		t.Fatal(err)
	}
	if err := ReconcileTrustBootstrap(ctx, pool, TrustBootstrap{repositories: map[string]packformat.TrustSnapshot{vector.Repository: root}}, time.Now().UTC()); err != nil {
		t.Fatal(err)
	}
	if err := ReconcileTrustBootstrap(ctx, pool, TrustBootstrap{repositories: map[string]packformat.TrustSnapshot{vector.Repository: root}}, time.Now().UTC()); err != nil {
		t.Fatal(err)
	}
	var rootAudits, rootOperations int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM administrative_audit_projections WHERE action_code='reference_pack_root_import' AND actor_kind='system' AND actor_user_id IS NULL AND source='startup'`).Scan(&rootAudits); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM reference_pack_operations WHERE actor_kind='system' AND kind='reconcile' AND terminal_at IS NOT NULL`).Scan(&rootOperations); err != nil {
		t.Fatal(err)
	}
	if rootAudits != 1 || rootOperations != 1 {
		t.Fatalf("bootstrap replay duplicated audit or operation: %d/%d", rootAudits, rootOperations)
	}
	storage := &coordinatorMemoryStorage{objects: map[string][]byte{}}
	if err := ReconcileBaseRelease(ctx, pool, storage, BaseReleaseOptions{ProfileClaimed: true, ClockTrusted: true, Limits: DefaultLimits()}, vector.At); err != nil {
		t.Fatal(err)
	}
	definitions := []jobs.Definition{}
	for _, kind := range []string{"import", "reverify", "refresh"} {
		identitySchema := jobs.HumanRouteIdentitySchema
		if kind == "import" {
			identitySchema = jobs.AttributedRouteIdentitySchema
		}
		definitions = append(definitions, jobs.Definition{JobKind: referenceCatalogJobKind(kind), ProgressUnitID: "reference_pack." + kind + ".request.v1", HandlerName: LifecycleWorkerKind, Extension: &jobs.ExtensionPolicy{IdentitySchemaID: identitySchema, OwnerProfileID: ProfileID, OperationKind: "reference_pack." + kind, ContractSHA256: strings.Repeat("a", 64), ProofRequired: true, MaxProofBytes: 1048576, ResourceRefs: []jobs.ExtensionResourceRefContract{{Kind: "reference_pack_version", MaxRefs: 1024}}}})
	}
	catalog, err := jobs.NewCatalog(definitions)
	if err != nil {
		t.Fatal(err)
	}
	transactions := collaborationsupport.NewJobTransactionsWithTerminalEffects(catalog, JobTerminalEffects{}, collaborationsupport.TestWorkerRuntimeContracts(definitions))
	manager, err := jobs.NewManager(jobs.ManagerOptions{Postgres: pool, Transactions: transactions, Catalog: catalog, Policy: jobs.ProductionRuntimePolicy(), Now: func() time.Time { return time.Now().UTC() }})
	if err != nil {
		t.Fatal(err)
	}
	store, err := extensionstore.New(pool, nil)
	if err != nil {
		t.Fatal(err)
	}
	finalizer, err := extensionstore.NewOwnerFinalizer(store, transactions, collaborationsupport.NewJobOwnerTransactionAdapters(), func() time.Time { return time.Now().UTC() }, func(err error) { t.Errorf("unexpected indeterminate commit: %v", err) })
	if err != nil {
		t.Fatal(err)
	}
	now := vector.At
	c, err := NewCoordinator(CoordinatorOptions{RegistryUsage: lifecycleFixtureUsage{}, Postgres: pool, Storage: storage, Configuration: Configuration{ClockTrusted: true}, Limits: DefaultLimits(), JobAdmission: transactions, JobExecutionGuard: transactions, JobOperations: manager, JobFinalizer: coordinatorTestFinalizer{owner: finalizer}, Now: func() time.Time { return now }})
	if err != nil {
		t.Fatal(err)
	}
	actor := uuid.New()
	if _, err := pool.Exec(ctx, `INSERT INTO users(id,email,display_name,password_hash,mfa_required,is_active,is_deployment_admin) VALUES($1,$2,'Coordinator fixture','hash',false,true,true)`, actor, actor.String()+"@example.test"); err != nil {
		t.Fatal(err)
	}

	return canonicalCoordinatorFixture{coordinator: c, pool: pool, manager: manager, storage: storage, actor: actor, catalog: catalog, transactions: transactions, definitions: definitions, vector: vector, container: container, now: &now}
}
