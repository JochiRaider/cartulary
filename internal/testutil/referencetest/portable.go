package referencetest

import (
	"bytes"
	"context"
	"io"
	"os"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/app/extensionassembly"
	"github.com/JochiRaider/cartulary/internal/app/referenceassembly"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data/testsupport"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/JochiRaider/cartulary/internal/platform/jobs/jobcontract"
	"github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Portable composes real owner admission, leases, persistence and rooted bytes.
// Its source identities come from canonical fixtures, never specification prose.
type Portable struct {
	T                    testing.TB
	Pool                 *pgxpool.Pool
	Root                 *referenceassembly.RootStorage
	Temporary, Published string
	Actor                uuid.UUID
	Pack                 testsupport.PortablePackFixture
	At                   time.Time
	Transactions         *jobs.TransactionService
	Manager              *jobs.Manager
}

func NewPortable(t testing.TB) *Portable {
	t.Helper()
	ctx := context.Background()
	db := pgtest.Start(t).PrepareIsolatedDatabaseT(t, "portable-completion")
	pool, err := pgxpool.New(ctx, db.DSN)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	f := &Portable{T: t, Pool: pool, Temporary: t.TempDir(), Published: t.TempDir(), Actor: uuid.New(), At: time.Now().UTC()}
	f.Root, err = referenceassembly.NewRootStorage(f.Temporary, f.Published)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(f.Root.Close)
	finalizer, err := extensionassembly.NewReferencePackMutationFinalizer(func(err error) { t.Errorf("unexpected fixture fatal: %v", err) })
	if err != nil {
		t.Fatal(err)
	}
	if err := reference_data.ReconcileBaseRelease(ctx, pool, f.Root, reference_data.BaseReleaseOptions{Finalizer: finalizer, Limits: reference_data.DefaultLimits(), ClockTrusted: true}, f.At); err != nil {
		t.Fatal(err)
	}
	read := func(name string) []byte {
		t.Helper()
		data, err := os.ReadFile("../../../contracts/reference-pack-fixtures/fixtures/" + name)
		if err != nil {
			t.Fatal(err)
		}
		return data
	}
	f.Pack = testsupport.PortableFixture(t, f.At, read("portable-input.v1.json"), read("portable-references.v1.json"))
	bootstrap, err := reference_data.AdmitTrustBootstrap(f.Pack.Bootstrap)
	if err != nil {
		t.Fatal(err)
	}
	if err := reference_data.ReconcileTrustBootstrap(ctx, pool, bootstrap, f.At); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO users(id,email,display_name,password_hash,mfa_required,is_active,is_deployment_admin) VALUES($1,$2,'Portable completion fixture','hash',false,true,true)`, f.Actor, f.Actor.String()+"@example.test"); err != nil {
		t.Fatal(err)
	}
	definitions := []jobs.Definition{{JobKind: "test.portable", ProgressUnitID: "reference_pack.import.request.v1", HandlerName: "test_portable", Terminal: jobcontract.TerminalContract{SchemaID: jobcontract.TerminalSuccessSchema, SuccessCodes: []string{"incident_bundle_imported"}}}}
	catalog, err := jobs.NewCatalog(definitions)
	if err != nil {
		t.Fatal(err)
	}
	f.Transactions = collaborationsupport.NewJobTransactionsWithTerminalEffects(catalog, reference_data.JobTerminalEffects{}, collaborationsupport.TestWorkerRuntimeContracts(definitions))
	f.Manager, err = jobs.NewManager(jobs.ManagerOptions{Postgres: pool, Transactions: f.Transactions, Catalog: catalog, Policy: jobs.ProductionRuntimePolicy(), Now: time.Now})
	if err != nil {
		t.Fatal(err)
	}
	return f
}

func (f *Portable) Service(storage reference_data.ArtifactStorage, lifetime *reference_data.VerificationLifetime) reference_data.IncidentReferences {
	f.T.Helper()
	limits := reference_data.DefaultLimits()
	limits.ReferencePacks.MaxVerificationSeconds = 60
	service, err := reference_data.NewIncidentReferences(reference_data.IncidentReferenceOptions{Lifetime: lifetime, Postgres: f.Pool, Storage: storage, Limits: limits, Configuration: reference_data.Configuration{ClockTrusted: true}, JobExecutions: f.Transactions, JobOperations: f.Manager, Now: func() time.Time { return f.At }})
	if err != nil {
		f.T.Fatal(err)
	}
	return service
}

func (f *Portable) Request() reference_data.IncidentReferenceImportRequest {
	f.T.Helper()
	ctx := context.Background()
	tx, err := f.Pool.Begin(ctx)
	if err != nil {
		f.T.Fatal(err)
	}
	defer tx.Rollback(ctx)
	total := 1
	job, err := f.Transactions.CreateQueuedTx(ctx, tx, jobs.EnqueueParams{JobKind: "test.portable", Scope: jobs.Scope{Kind: jobs.ScopeKindDeployment}, SubmittedByUserID: f.Actor, Cancelable: true, Progress: jobs.Progress{Total: &total}}, time.Now())
	if err != nil {
		f.T.Fatal(err)
	}
	if err := tx.Commit(ctx); err != nil {
		f.T.Fatal(err)
	}
	id := uuid.MustParse(job.JobID)
	execution, claimed, err := f.Manager.Claim(ctx, id)
	if err != nil || !claimed {
		f.T.Fatal(claimed, err)
	}
	return f.RequestFor(execution)
}

// RequestFor supplies canonical container input to an already admitted parent.
func (f *Portable) RequestFor(execution jobs.Execution) reference_data.IncidentReferenceImportRequest {
	f.T.Helper()
	content := bytes.Clone(f.Pack.Content)
	return reference_data.IncidentReferenceImportRequest{References: f.Pack.References, ContentManifest: &content, EmbeddedPaths: []string{f.Pack.Path}, IncidentID: uuid.New(), OperationID: execution.JobID(), ActorID: f.Actor, At: f.At, Execution: execution, ExecutionStarted: time.Now(), OpenContainer: func(context.Context, string) (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(f.Pack.Container)), nil
	}}
}

// CohortRequest uses the canonical two-version cohort, with separate containers.
func (f *Portable) CohortRequest() reference_data.IncidentReferenceImportRequest {
	f.T.Helper()
	read := func(name string) []byte {
		data, err := os.ReadFile("../../../contracts/reference-pack-fixtures/fixtures/" + name)
		if err != nil {
			f.T.Fatal(err)
		}
		return data
	}
	bootstrap, refs, content, containers := testsupport.PortableCohortFixture(f.T, f.At, read("portable-input.v1.json"), read("portable-references.v1.json"), false)
	if !bytes.Equal(bootstrap, f.Pack.Bootstrap) {
		f.T.Fatal("cohort trust differs from fixture")
	}
	request := f.Request()
	request.References, request.ContentManifest, request.EmbeddedPaths = refs, &content, nil
	for path := range containers {
		request.EmbeddedPaths = append(request.EmbeddedPaths, path)
	}
	request.OpenContainer = func(_ context.Context, path string) (io.ReadCloser, error) {
		data, ok := containers[path]
		if !ok {
			return nil, os.ErrNotExist
		}
		return io.NopCloser(bytes.NewReader(data)), nil
	}
	return request
}
