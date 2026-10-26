package pgtest

import (
	"context"
	"errors"
	"github.com/JochiRaider/cartulary/internal/testutil/pgschema"
	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

type attachmentFixture struct {
	*attachmentRegistry
	env map[string]string
}

func (f *attachmentFixture) attach(ctx context.Context) (*Harness, error) {
	return f.attachmentRegistry.attach(ctx, f.env)
}
func (f *attachmentFixture) binding() (attachmentIdentity, error) {
	return bindFixture(f.env).identity()
}

func fixtureAttachmentRegistry() *attachmentFixture {
	env := map[string]string{
		suiteservices.HarnessServiceDependenciesEnv: "postgres",
		suiteservices.PGAdminDSNEnv:                 "postgres://fixture@localhost/postgres",
		suiteservices.PGDSNTemplateEnv:              "postgres://fixture@localhost/{database}",
		suiteservices.PGTemplateDBEnv:               "suite_template",
		suiteservices.PGSchemaHashEnv:               pgschema.MustHash(),
		suiteservices.PGProvisioningDigestEnv:       ProvisioningDigest(),
		suiteservices.PGTemplateDigestEnv:           TemplateDigest(pgschema.MustHash()),
		suiteservices.PGGenerationEnv:               "12345",
		suiteservices.PGIsolationEnv:                "shared",
		suiteservices.PGAllocationRefEnv:            "pgtest-fixture",
	}
	r := &attachmentFixture{attachmentRegistry: newAttachmentRegistry(), env: env}
	r.operations.inspect = func(context.Context, string) (string, error) { return env[suiteservices.PGGenerationEnv], nil }
	r.operations.ping = func(context.Context, string, func(context.Context)) error { return nil }
	return r
}

func TestStartSharedUsesAttachEnvWithoutStartingOwnedHarness(t *testing.T) {
	r := fixtureAttachmentRegistry()
	calls := 0
	r.operations.ping = func(_ context.Context, dsn string, _ func(context.Context)) error {
		calls++
		if dsn != r.env[suiteservices.PGAdminDSNEnv] {
			t.Fatal("incorrect admitted endpoint")
		}
		return nil
	}
	h, err := r.attach(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if _, physical := any(h).(interface{ Close(context.Context) error }); physical {
		t.Fatal("borrower exposes physical shutdown")
	}
	if h.templateDB != "suite_template" || !strings.HasSuffix(h.dsnFor("tenant_one"), "/tenant_one") {
		t.Fatal("incorrect fixture binding")
	}
	for _, key := range []string{suiteservices.PGAdminDSNEnv, suiteservices.PGDSNTemplateEnv, suiteservices.PGTemplateDBEnv, suiteservices.PGSchemaHashEnv, suiteservices.PGProvisioningDigestEnv, suiteservices.PGTemplateDigestEnv, suiteservices.PGIsolationEnv, suiteservices.PGAllocationRefEnv, suiteservices.SuiteRuntimeLeaseIDEnv, "CARTULARY_FIXTURE_CAPABILITY", "CARTULARY_TEST_RESULTS_DIR", postgresFixturePolicyDefaultEnv} {
		t.Run("reject_changed_"+key, func(t *testing.T) {
			old, present := r.env[key]
			r.env[key] = "superseded"
			defer func() {
				if present {
					r.env[key] = old
				} else {
					delete(r.env, key)
				}
			}()
			if got, err := r.attach(t.Context()); got != nil || err == nil {
				t.Fatal("cached generation bypassed binding identity")
			}
		})
	}
	if got, err := r.attach(t.Context()); got != h || err != nil || calls != 1 {
		t.Fatal("repeated attachment lost generation state", err, calls)
	}
	h.quarantined.Store(true)
	if got, err := r.attach(t.Context()); got != nil || !errors.Is(err, errGenerationQuarantined) || calls != 1 {
		t.Fatal("repeated attachment cleared quarantine", err, calls)
	}
	independent := fixtureAttachmentRegistry()
	if got, err := independent.attach(t.Context()); got == h || err != nil {
		t.Fatal("independent registry inherited state", err)
	}
	r.env[suiteservices.PGGenerationEnv] = "67890"
	if got, err := r.attach(t.Context()); got == nil || got == h || err != nil || calls != 2 {
		t.Fatal("independent generation failed admission", err, calls)
	}
}

func TestStartSharedRequiresExplicitAttachment(t *testing.T) {
	t.Run("verified_failure_is_terminal", testAttachmentTerminalFailure)
	t.Run("binding_ignores_ambient_events", testAttachmentAmbientEvents)
	t.Run("admission_concurrency", testAttachmentConcurrency)
	t.Run("bound_authority_and_journals", testAttachmentBoundAuthority)
	t.Run("terminal_cancellation_and_publication", testAttachmentTerminalBoundaries)
	for _, fault := range []string{"missing", "incomplete", "schema", "generation", "undeclared"} {
		t.Run(fault, func(t *testing.T) {
			r := fixtureAttachmentRegistry()
			calls := 0
			r.operations.inspect = func(context.Context, string) (string, error) { calls++; return "12345", nil }
			r.operations.ping = func(context.Context, string, func(context.Context)) error {
				t.Fatal("invalid binding reached readiness")
				return nil
			}
			switch fault {
			case "missing":
				delete(r.env, suiteservices.PGAdminDSNEnv)
				delete(r.env, suiteservices.PGDSNTemplateEnv)
			case "incomplete":
				delete(r.env, suiteservices.PGAdminDSNEnv)
			case "schema":
				r.env[suiteservices.PGSchemaHashEnv] = "old"
			case "generation":
				r.env[suiteservices.PGGenerationEnv] = "old"
			case "undeclared":
				r.env[suiteservices.HarnessServiceDependenciesEnv] = "object_store"
			}
			if got, err := r.attach(t.Context()); got != nil || err == nil {
				t.Fatal("invalid binding admitted")
			}
			if fault != "generation" && calls != 0 {
				t.Fatal("invalid binding acquired resources")
			}
		})
	}
}

func TestCreateDatabaseRejectsOmittedServiceBeforeAcquisition(t *testing.T) {
	dbOps := defaultDatabaseOperations()
	t.Setenv(suiteservices.HarnessServiceDependenciesEnv, "object_store")

	createCalls := 0
	dbOps.create = func(_ context.Context, _, _, _ string, confirmed func() error) (bool, error) {
		createCalls++
		return true, confirmed()
	}

	_, err := (&Harness{binding: bindFixture(processFixtureInputs()), operations: &dbOps}).newDatabase(context.Background(), "guarded", suiteservices.FixtureReusePerTest, fixtureAttribution{})
	var dependencyErr *suiteservices.ServiceDependencyError
	if !errors.As(err, &dependencyErr) || dependencyErr.Service != "postgres" || dependencyErr.Reason != "omitted" {
		t.Fatalf("unexpected dependency error: %#v", err)
	}
	if createCalls != 0 {
		t.Fatalf("postgres acquisition ran before the dependency guard: calls=%d", createCalls)
	}
}

func testAttachmentTerminalFailure(t *testing.T) {
	r := fixtureAttachmentRegistry()
	failure := errors.New("readiness failed")
	probes := 0
	r.operations.ping = func(context.Context, string, func(context.Context)) error {
		probes++
		if probes == 1 {
			return failure
		}
		return nil
	}
	if _, err := r.attach(t.Context()); !errors.Is(err, failure) {
		t.Fatal(err)
	}
	if h, err := r.attach(t.Context()); h != nil || !errors.Is(err, failure) || !errors.Is(err, errGenerationQuarantined) || probes != 1 {
		t.Error("verified generation was readmitted", err, probes)
	}
}

func testAttachmentAmbientEvents(t *testing.T) {
	r := fixtureAttachmentRegistry()
	t.Setenv(suiteservices.SuiteIDEnv, "invalid-ambient-suite")
	t.Setenv("CARTULARY_TEST_RESULTS_DIR", "/dev/null/not-a-directory")
	if _, err := r.attach(t.Context()); err != nil {
		t.Fatal("local registry inherited ambient event failure", err)
	}
}

func testAttachmentTerminalBoundaries(t *testing.T) {
	for _, failure := range []string{"cancellation", "publication"} {
		t.Run(failure, func(t *testing.T) {
			r := fixtureAttachmentRegistry()
			if failure == "publication" {
				r.env[suiteservices.SuiteIDEnv] = "publication"
				r.env["CARTULARY_TEST_RESULTS_DIR"] = "/dev/null/not-a-directory"
			}
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			inspections, probes := 0, 0
			r.operations.inspect = func(context.Context, string) (string, error) { inspections++; return "12345", nil }
			r.operations.ping = func(context.Context, string, func(context.Context)) error {
				probes++
				if failure == "cancellation" {
					cancel()
					return ctx.Err()
				}
				return nil
			}
			_, initial := r.attach(ctx)
			if !errors.Is(initial, errGenerationQuarantined) {
				t.Fatal("verified failure was not quarantined", initial)
			}
			if failure == "cancellation" && !errors.Is(initial, context.Canceled) {
				t.Fatal("lost cancellation", initial)
			}
			for range 2 {
				if h, err := r.attach(t.Context()); h != nil || err != initial || inspections != 1 || probes != 1 {
					t.Fatal("failed generation retried admission", err, inspections, probes)
				}
			}
		})
	}
}

func testAttachmentConcurrency(t *testing.T) {
	r := fixtureAttachmentRegistry()
	entered, release := make(chan struct{}), make(chan struct{})
	var probes atomic.Int32
	r.operations.inspect = func(_ context.Context, dsn string) (string, error) {
		if dsn == "second" {
			return "67890", nil
		}
		return "12345", nil
	}
	r.operations.ping = func(_ context.Context, dsn string, _ func(context.Context)) error {
		probes.Add(1)
		if dsn != "second" {
			close(entered)
			<-release
		}
		return nil
	}
	result := make(chan error, 1)
	go func() { _, err := r.attach(t.Context()); result <- err }()
	<-entered
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Millisecond)
	defer cancel()
	if _, err := r.attach(ctx); !errors.Is(err, context.DeadlineExceeded) {
		t.Error("waiter did not cancel independently", err)
	}
	other := bindFixture(r.env).inputs
	other[suiteservices.PGAdminDSNEnv] = "second"
	other[suiteservices.PGGenerationEnv] = "67890"
	independent := make(chan error, 1)
	go func() { _, err := r.attachmentRegistry.attach(t.Context(), other); independent <- err }()
	select {
	case err := <-independent:
		if err != nil {
			t.Error(err)
		}
	case <-time.After(time.Second):
		t.Error("unrelated generation blocked behind admission")
	}
	close(release)
	if err := <-result; err != nil {
		t.Fatal("waiter retired initiating admission", err)
	}
	if _, err := r.attach(t.Context()); err != nil || probes.Load() != 2 {
		t.Fatal("successful admission was replayed", err)
	}
}

func testAttachmentBoundAuthority(t *testing.T) {
	first, second := fixtureAttachmentRegistry(), fixtureAttachmentRegistry()
	for i, r := range []*attachmentFixture{first, second} {
		r.env[suiteservices.SuiteIDEnv] = "bound"
		r.env["CARTULARY_TEST_RUN_ID"] = "binding-test"
		r.env["CARTULARY_TEST_RESULTS_DIR"] = t.TempDir()
		r.env["CARTULARY_FIXTURE_CAPABILITY"] = "postgres_migration"
		r.env[suiteservices.PGIsolationEnv] = "exclusive"
		r.env[postgresFixturePolicyDefaultEnv] = postgresFixturePolicyTemplateClone
		if i == 0 {
			r.env["CARTULARY_TEST_RESULTS_DIR"] = "/dev/null/not-a-directory"
		}
	}
	if _, err := first.attach(t.Context()); !errors.Is(err, errGenerationQuarantined) {
		t.Fatal("missing publication quarantine", err)
	}
	h, err := second.attach(t.Context())
	if err != nil {
		t.Fatal("independent journal inherited failure", err)
	}
	second.env["CARTULARY_FIXTURE_CAPABILITY"] = "postgres_transaction"
	second.env[suiteservices.HarnessServiceDependenciesEnv] = "object_store"
	second.env[postgresFixturePolicyDefaultEnv] = postgresFixturePolicyTransaction
	t.Setenv("CARTULARY_FIXTURE_CAPABILITY", "postgres_dedicated")
	t.Setenv("CARTULARY_TEST_RESULTS_DIR", "/dev/null/not-a-directory")
	t.Setenv(suiteservices.HarnessServiceDependenciesEnv, "object_store")
	t.Setenv(postgresFixturePolicyDefaultEnv, postgresFixturePolicyTransaction)
	if err := h.admitMigration(); err != nil {
		t.Fatal("admitted authority followed changed inputs", err)
	}
	h.requireSelectedPostgresFixturePolicyT(t, postgresFixturePolicyTemplateClone)
	ops := defaultDatabaseOperations()
	h.operations = &ops
	ops.create = func(_ context.Context, _, _, _ string, confirmed func() error) (bool, error) {
		return true, confirmed()
	}
	if err := h.createDatabase(t.Context(), "ct_bound", "", suiteservices.FixtureReuseMigrationScratch, fixtureAttribution{}); err != nil {
		t.Fatal("creation inherited ambient dependencies or destination", err)
	}
	ledger, found, err := suiteservices.CurrentResourceLedger(h.binding.environment())
	if err != nil || !found || len(ledger.Databases) != 1 || ledger.Databases[0] != "ct_bound" {
		t.Fatal("bound journal lost creation", ledger, err)
	}
	if _, err := second.attach(t.Context()); err == nil {
		t.Fatal("changed binding bypassed cache admission")
	}
}
