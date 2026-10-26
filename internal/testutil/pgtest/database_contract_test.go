package pgtest

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	database_migrations "github.com/JochiRaider/cartulary/internal/modules/database_migrations"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/postgrescleanup"
	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices"
	"github.com/JochiRaider/cartulary/internal/testutil/testfailure"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestDatabaseNamesAreUniqueAcrossSimulatedProcesses(t *testing.T) {
	first := &Harness{suiteHash: "suitehash", processHash: "procaaaa"}
	second := &Harness{suiteHash: "suitehash", processHash: "procbbbb"}

	seen := make(map[string]struct{})
	validName := regexp.MustCompile(`^[a-z0-9_]+$`)
	for _, harness := range []*Harness{first, second} {
		for i := 0; i < 8; i++ {
			name := harness.nextDatabaseName("Prefix With Spaces-and-symbols!")
			if len(name) > 63 {
				t.Fatalf("database name exceeds postgres identifier limit: %q", name)
			}
			if !validName.MatchString(name) {
				t.Fatalf("database name must be sanitized, got %q", name)
			}
			if _, exists := seen[name]; exists {
				t.Fatalf("database name collision detected: %q", name)
			}
			seen[name] = struct{}{}
		}
	}
}

func TestPrepareDatabaseTemplateModeClonesWithoutMigrationReplay(t *testing.T) {
	dbOps := defaultDatabaseOperations()
	t.Setenv(suiteservices.HarnessServiceDependenciesEnv, "postgres")
	t.Run("template_preparation_contract", testTemplatePreparationContract)
	t.Setenv(suiteservices.SuiteIDEnv, "suite-template-mode")
	t.Setenv(suiteservices.TargetEnv, "backend-store")
	t.Setenv("CARTULARY_TEST_RESULTS_DIR", t.TempDir())
	t.Setenv("CARTULARY_TEST_RUN_ID", "template-mode")

	var createCalls []struct {
		Name     string
		Template string
	}
	dbOps.create = func(ctx context.Context, adminDSN string, name string, templateDB string, confirmed func() error) (bool, error) {
		createCalls = append(createCalls, struct {
			Name     string
			Template string
		}{
			Name:     name,
			Template: templateDB,
		})
		return true, confirmed()
	}

	migrateCalls := 0
	initialize := func(ctx context.Context, db *sql.DB, _ string, source *database_migrations.Source) error {
		migrateCalls++
		return nil
	}

	harness := &Harness{binding: bindFixture(processFixtureInputs()), operations: &dbOps,
		adminDSN:    "postgres://cartulary:cartulary@127.0.0.1:5432/postgres?sslmode=disable",
		dsnTemplate: "postgres://cartulary:cartulary@127.0.0.1:5432/{database}?sslmode=disable",
		templateDB:  "suite_template",
		suiteHash:   "suitehash",
		processHash: "procaaaa",
	}

	dbOps.migrate = func(ctx context.Context, db *sql.DB, source *database_migrations.Source) error {
		return initialize(ctx, db, "", source)
	}

	testDB, err := harness.PrepareDatabase(context.Background(), "bootstrap")
	if err != nil {
		t.Fatalf("prepare database from template: %v", err)
	}
	if len(createCalls) != 1 {
		t.Fatalf("expected one database create call, got %d", len(createCalls))
	}
	if createCalls[0].Template != "suite_template" {
		t.Fatalf("expected template clone create call, got template %q", createCalls[0].Template)
	}
	if migrateCalls != 0 {
		t.Fatalf("expected template mode to skip per-database migrations, got %d calls", migrateCalls)
	}
	if testDB.Name == "" || !strings.Contains(testDB.DSN, testDB.Name) {
		t.Fatalf("expected prepared database dsn to reference the cloned database, got %#v", testDB)
	}

	scope, ok, err := suiteservices.Summarize(nil)
	if err != nil {
		t.Fatalf("summarize suite service events: %v", err)
	}
	if !ok {
		t.Fatal("expected suite service summary")
	}
	if len(scope.Postgres.DatabasePreparations) != 1 {
		t.Fatalf("expected one database preparation, got %#v", scope.Postgres.DatabasePreparations)
	}
	preparation := scope.Postgres.DatabasePreparations[0]
	if preparation.Name != testDB.Name {
		t.Fatalf("unexpected prepared database name: got %q want %q", preparation.Name, testDB.Name)
	}
	if preparation.Strategy != suiteservices.PostgresPreparationTemplateClone {
		t.Fatalf("expected template-clone preparation, got %#v", preparation)
	}
	if preparation.TemplateDatabase != "suite_template" {
		t.Fatalf("unexpected template database: got %#v", preparation)
	}
	if preparation.Target != "backend-store" {
		t.Fatalf("unexpected preparation target: got %#v", preparation)
	}
}

func TestPrepareIsolatedDatabaseTUsesFreshTemplateClones(t *testing.T) {
	dbOps := defaultDatabaseOperations()
	t.Setenv(suiteservices.HarnessServiceDependenciesEnv, "postgres")
	t.Setenv(suiteservices.SuiteIDEnv, "")
	t.Setenv(postgresFixturePolicyDefaultEnv, postgresFixturePolicyTemplateClone)

	var creates []struct {
		Name     string
		Template string
	}
	dbOps.create = func(ctx context.Context, adminDSN string, name string, templateDB string, confirmed func() error) (bool, error) {
		creates = append(creates, struct {
			Name     string
			Template string
		}{Name: name, Template: templateDB})
		return true, confirmed()
	}
	var drops []string
	dbOps.drop = func(ctx context.Context, adminDSN string, name string, _ func(postgrescleanup.DropAttempt)) error {
		drops = append(drops, name)
		return nil
	}
	harness := &Harness{binding: bindFixture(processFixtureInputs()), operations: &dbOps,
		adminDSN:    "postgres://cartulary:cartulary@127.0.0.1:5432/postgres?sslmode=disable",
		dsnTemplate: "postgres://cartulary:cartulary@127.0.0.1:5432/{database}?sslmode=disable",
		templateDB:  "suite_template",
		suiteHash:   "suitehash",
		processHash: "procaaaa",
	}

	var firstName string
	t.Run("first clone", func(t *testing.T) {
		firstName = harness.PrepareIsolatedDatabaseT(t, "clone-policy").Name
	})
	var secondName string
	t.Run("second clone", func(t *testing.T) {
		secondName = harness.PrepareIsolatedDatabaseT(t, "clone-policy").Name
	})

	if firstName == "" || secondName == "" || firstName == secondName {
		t.Fatalf("expected distinct template-clone database names, got first=%q second=%q", firstName, secondName)
	}
	if len(creates) != 2 {
		t.Fatalf("expected two database creates, got %#v", creates)
	}
	for _, create := range creates {
		if create.Template != "suite_template" {
			t.Fatalf("expected template clone create, got %#v", create)
		}
	}
	if len(drops) != 2 {
		t.Fatalf("expected cloned databases to be dropped outside active suite, got %v", drops)
	}
}

func TestDatabaseCleanupUsesSingleDropOnlyForOwnedDatabase(t *testing.T) {
	dbOps := defaultDatabaseOperations()

	harness := &Harness{binding: bindFixture(processFixtureInputs()), operations: &dbOps,
		adminDSN: "postgres://cartulary:cartulary@127.0.0.1:5432/postgres?sslmode=disable",
		ownedDatabases: map[string]struct{}{
			"ct_deadbeef_cafe0001_000001_owned": {},
		},
	}
	dropCalls := 0
	dbOps.drop = func(ctx context.Context, _ string, _ string, _ func(postgrescleanup.DropAttempt)) error {
		dropCalls++
		if ctx.Err() != nil {
			t.Fatalf("cleanup delegate received canceled parent: %v", ctx.Err())
		}
		return nil
	}

	if err := harness.dropDatabase(context.Background(), "ct_deadbeef_cafe0001_000001_owned", suiteservices.FixtureReusePerTest, fixtureAttribution{}); err != nil {
		t.Fatalf("drop owned active database: %v", err)
	}
	if dropCalls != 1 {
		t.Fatalf("expected one coordinated drop request, got %d", dropCalls)
	}

	if err := harness.dropDatabase(context.Background(), "ct_deadbeef_cafe0001_000002_unowned", suiteservices.FixtureReusePerTest, fixtureAttribution{}); err == nil {
		t.Fatal("expected unowned database cleanup to fail before deletion")
	}
	if dropCalls != 1 {
		t.Fatalf("unowned database reached deletion: calls=%d", dropCalls)
	}
}

func TestDatabaseCleanupPassesOwnerContext(t *testing.T) {
	dbOps := defaultDatabaseOperations()

	const databaseName = "ct_deadbeef_cafe0001_000001_owned"
	harness := &Harness{binding: bindFixture(processFixtureInputs()), operations: &dbOps,
		adminDSN: "postgres://cartulary:cartulary@127.0.0.1:5432/postgres?sslmode=disable",
		ownedDatabases: map[string]struct{}{
			databaseName: {},
		},
	}
	parent, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	dropCalls := 0
	dbOps.drop = func(ctx context.Context, _ string, _ string, _ func(postgrescleanup.DropAttempt)) error {
		dropCalls++
		if ctx != parent || ctx.Err() != nil {
			t.Fatalf("coordinated cleanup inherited an expired child context: %v", ctx.Err())
		}
		return nil
	}

	if err := harness.dropDatabase(parent, databaseName, suiteservices.FixtureReusePerTest, fixtureAttribution{}); err != nil {
		t.Fatalf("drop owned database: %v", err)
	}
	if dropCalls != 1 {
		t.Fatalf("coordinated cleanup calls = %d, want 1", dropCalls)
	}
}

func TestDatabaseCleanupRetainsFailedRecoveryAuthority(t *testing.T) {
	dbOps := defaultDatabaseOperations()
	authorizePGSuite(t)
	t.Setenv(suiteservices.SuiteIDEnv, "suite-cleanup-failure")
	t.Setenv(suiteservices.TargetEnv, "backend-integration")
	t.Setenv("CARTULARY_TEST_RESULTS_DIR", t.TempDir())
	t.Setenv("CARTULARY_TEST_RUN_ID", "cleanup-failure")
	var createCalls atomic.Int32
	dbOps.create = func(_ context.Context, _, _, _ string, confirmed func() error) (bool, error) {
		createCalls.Add(1)
		return true, confirmed()
	}

	const databaseName = "ct_deadbeef_cafe0001_000001_owned"
	harness := &Harness{binding: bindFixture(processFixtureInputs()), operations: &dbOps, templateDB: "suite_template", ownedDatabases: map[string]struct{}{databaseName: {}}}
	registry := fixtureAttachmentRegistry()
	binding, err := registry.binding()
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan struct{})
	close(done)
	registry.generations[binding.generation] = &admittedAttachment{identity: binding, binding: bindFixture(registry.env), fixture: harness, done: done}
	if err := suiteservices.RecordEvent(nil, suiteservices.Event{Type: suiteservices.EventPostgresDBCreated, Name: databaseName, Kind: "template-clone"}); err != nil {
		t.Fatalf("record owned database: %v", err)
	}
	dropCalls := 0
	dbOps.drop = func(_ context.Context, _ string, _ string, _ func(postgrescleanup.DropAttempt)) error {
		dropCalls++
		return errors.New("permission denied")
	}
	if err := harness.dropDatabase(context.Background(), databaseName, suiteservices.FixtureReusePerTest, fixtureAttribution{}); err == nil {
		t.Fatal("expected cleanup failure")
	}
	if dropCalls != 1 {
		t.Fatalf("non-connection failure delegated %d times", dropCalls)
	}
	ledger, found, err := suiteservices.CurrentResourceLedger(nil)
	if err != nil || !found {
		t.Fatalf("read cleanup-failure ledger: found=%t err=%v", found, err)
	}
	if len(ledger.Databases) != 1 || ledger.Databases[0] != databaseName {
		t.Fatalf("failed cleanup lost recovery authority: %#v", ledger.Databases)
	}

	dbOps.drop = func(_ context.Context, _ string, _ string, _ func(postgrescleanup.DropAttempt)) error {
		dropCalls++
		return context.Canceled
	}
	if err := harness.dropDatabase(context.Background(), databaseName, suiteservices.FixtureReusePerTest, fixtureAttribution{}); !errors.Is(err, context.Canceled) {
		t.Fatalf("interrupted cleanup error = %v, want context cancellation", err)
	}
	if dropCalls != 2 {
		t.Fatalf("interrupted cleanup delegated %d total times", dropCalls)
	}

	var preparations sync.WaitGroup
	errorsSeen := make(chan error, 16)
	for range 16 {
		preparations.Go(func() {
			_, err := harness.PrepareDatabase(t.Context(), "after_failed_cleanup")
			errorsSeen <- err
		})
	}
	preparations.Wait()
	close(errorsSeen)
	for err := range errorsSeen {
		if !errors.Is(err, errGenerationQuarantined) {
			t.Fatalf("preparation after failed cleanup = %v", err)
		}
	}
	if createCalls.Load() != 0 {
		t.Fatalf("quarantined generation performed %d creates", createCalls.Load())
	}

	// Repeated attachment cannot stand in for physical settlement by the owner.
	if acquired, err := registry.attach(t.Context()); acquired != nil || !errors.Is(err, errGenerationQuarantined) {
		t.Fatalf("reacquired quarantined attachment: %v, %v", acquired, err)
	}
	for _, entry := range []struct {
		name string
		run  func(testing.TB)
	}{

		{"fresh", func(tb testing.TB) { harness.NewDatabaseT(tb, "rejected") }},
		{"clone", func(tb testing.TB) { harness.PrepareIsolatedDatabaseT(tb, "rejected") }},
		{"open_clone", func(tb testing.TB) { harness.OpenIsolatedDatabaseT(tb, "rejected", postgres.PurposeRuntime) }},
		{"migration", func(tb testing.TB) { harness.MigrationDatabaseT(tb) }},
		{"migration_through", func(tb testing.TB) { harness.MigrationDatabaseThroughT(tb, 1) }},
		{"rollback", func(tb testing.TB) { harness.BeginRollbackDBT(tb, "rejected") }},
	} {
		t.Run(entry.name, func(t *testing.T) {
			expectQuarantinedPreparation(t, entry.run)
		})
	}
	if createCalls.Load() != 0 {
		t.Fatalf("testing entry point prepared a quarantined generation: %d", createCalls.Load())
	}

	// Existing owned resources may finish cleanup, but success cannot restore
	// this generation's eligibility or erase its prior failed row outcome.
	dbOps.drop = func(_ context.Context, _ string, _ string, _ func(postgrescleanup.DropAttempt)) error {
		dropCalls++
		return nil
	}
	if err := harness.DropDatabase(t.Context(), databaseName); err != nil {
		t.Fatal(err)
	}
	if _, err := harness.PrepareDatabase(t.Context(), "after_late_settlement"); !errors.Is(err, errGenerationQuarantined) {
		t.Fatalf("late settlement readmitted generation: %v", err)
	}
	independent := &Harness{binding: bindFixture(processFixtureInputs()), operations: &dbOps, templateDB: "independent_template"}
	db, err := independent.PrepareDatabase(t.Context(), "independent_generation")
	if err != nil || db == nil || createCalls.Load() != 1 {
		t.Fatalf("unrelated generation failed: %v, %v, creates=%d", db, err, createCalls.Load())
	}
	if err := independent.DropDatabase(t.Context(), db.Name); err != nil {
		t.Fatal(err)
	}

	// Partial provisioning may also report failed cleanup after a committed
	// CREATE. Preserve that failure and reject subsequent preparation.
	partialFailure := errors.Join(errors.New("clone finalization failed"), context.DeadlineExceeded)
	dbOps.create = func(_ context.Context, _, _, _ string, confirmed func() error) (bool, error) {
		createCalls.Add(1)
		return true, errors.Join(confirmed(), partialFailure)
	}
	partial := &Harness{binding: bindFixture(processFixtureInputs()), operations: &dbOps, templateDB: "partial_template"}
	if _, err := partial.PrepareDatabase(t.Context(), "partial"); !errors.Is(err, partialFailure) {
		t.Fatalf("partial preparation lost primary failure: %v", err)
	}
	if _, err := partial.PrepareDatabase(t.Context(), "after_partial"); !errors.Is(err, errGenerationQuarantined) || createCalls.Load() != 2 {
		t.Fatalf("partial preparation retried: %v, creates=%d", err, createCalls.Load())
	}
}

type preparationFailureProbe struct{ testing.TB }
type capturedPreparationFailure struct{ message string }

func (p preparationFailureProbe) Fatal(args ...any) {
	panic(capturedPreparationFailure{message: fmt.Sprint(args...)})
}

func (p preparationFailureProbe) Fatalf(format string, args ...any) {
	panic(capturedPreparationFailure{message: fmt.Sprintf(format, args...)})
}

func expectQuarantinedPreparation(t *testing.T, run func(testing.TB)) {
	t.Helper()
	var captured *capturedPreparationFailure
	func() {
		defer func() {
			if value := recover(); value != nil {
				failure, ok := value.(capturedPreparationFailure)
				if !ok {
					panic(value)
				}
				captured = &failure
			}
		}()
		run(preparationFailureProbe{TB: t})
	}()
	if captured == nil || !strings.HasPrefix(captured.message, testfailure.Marker) {
		t.Fatalf("preparation did not publish typed fixture failure: %#v", captured)
	}
	var envelope testfailure.Envelope
	if err := json.Unmarshal([]byte(strings.TrimPrefix(captured.message, testfailure.Marker)), &envelope); err != nil {
		t.Fatal(err)
	}
	if err := envelope.Validate(); err != nil || envelope.Phase != "setup" || envelope.FailureClass != "harness" || envelope.FailureReason != "fixture_error" || envelope.ReadinessStage != "capability" || envelope.AttemptCount != 0 {
		t.Fatalf("quarantine setup classification: %#v, %v", envelope, err)
	}
}

func authorizePGSuite(t *testing.T) {
	t.Helper()
	t.Setenv(suiteservices.HarnessServiceDependenciesEnv, "postgres")
	t.Setenv(suiteservices.CallModeEnv, "owned")
	t.Setenv(suiteservices.SuiteRuntimeRootEnv, "/private/pgtest-suite-runtime")
	t.Setenv(suiteservices.SuiteRuntimeLeaseIDEnv, "00000000-0000-4000-8000-000000000001")
	t.Setenv(suiteservices.SuiteRuntimeRunIDEnv, "pgtest-suite")
}
