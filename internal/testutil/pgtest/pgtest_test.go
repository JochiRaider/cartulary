package pgtest

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/moby/moby/api/types/network"
	testcontainers "github.com/testcontainers/testcontainers-go"

	database_migrations "github.com/JochiRaider/cartulary/internal/modules/database_migrations"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/pgschema"
	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices"
	"github.com/JochiRaider/cartulary/internal/testutil/testcontainersx"
)

func TestStartSharedUsesAttachEnvWithoutStartingOwnedHarness(t *testing.T) {
	resetSharedHarness(t)

	t.Setenv(suiteservices.PGAdminDSNEnv, "postgres://cartulary:cartulary@127.0.0.1:5432/postgres?sslmode=disable")
	t.Setenv(suiteservices.PGDSNTemplateEnv, "postgres://cartulary:cartulary@127.0.0.1:5432/{database}?sslmode=disable")
	t.Setenv(suiteservices.PGTemplateDBEnv, "suite_template")

	oldStart := startOwnedHarness
	oldPing := pingAdminDSNFn
	t.Cleanup(func() {
		startOwnedHarness = oldStart
		pingAdminDSNFn = oldPing
	})

	startCalls := 0
	startOwnedHarness = func(ctx context.Context) (*Harness, error) {
		startCalls++
		return nil, errors.New("owned harness should not start when attach env is present")
	}

	var pingDSN string
	pingAdminDSNFn = func(ctx context.Context, dsn string) error {
		pingDSN = dsn
		return nil
	}

	harness, err := StartShared(context.Background())
	if err != nil {
		t.Fatalf("start attached harness: %v", err)
	}
	if startCalls != 0 {
		t.Fatalf("expected attach mode to skip owned startup, got %d calls", startCalls)
	}
	if !harness.attached {
		t.Fatal("expected attached harness")
	}
	if harness.Container != nil {
		t.Fatal("expected attach mode to avoid creating a testcontainer")
	}
	if harness.templateDB != "suite_template" {
		t.Fatalf("unexpected template database: got %q", harness.templateDB)
	}
	if pingDSN != harness.AdminDSN() {
		t.Fatalf("unexpected admin dsn ping: got %q want %q", pingDSN, harness.AdminDSN())
	}
	if got := harness.dsnFor("tenant_one"); !strings.Contains(got, "/tenant_one?") {
		t.Fatalf("expected attach dsn template replacement, got %q", got)
	}
}

func TestStartSharedFallsBackToOwnedHarnessWhenAttachEnvIsAbsent(t *testing.T) {
	resetSharedHarness(t)
	t.Setenv(suiteservices.PGAdminDSNEnv, "")
	t.Setenv(suiteservices.PGDSNTemplateEnv, "")
	t.Setenv(suiteservices.PGTemplateDBEnv, "")

	oldStart := startOwnedHarness
	t.Cleanup(func() {
		startOwnedHarness = oldStart
	})

	startCalls := 0
	want := &Harness{
		adminDSN:    "postgres://cartulary:cartulary@127.0.0.1:5432/postgres?sslmode=disable",
		dsnTemplate: "postgres://cartulary:cartulary@127.0.0.1:5432/{database}?sslmode=disable",
		suiteHash:   "suitehash",
		processHash: "processid",
	}
	startOwnedHarness = func(ctx context.Context) (*Harness, error) {
		startCalls++
		return want, nil
	}

	harness, err := StartShared(context.Background())
	if err != nil {
		t.Fatalf("start shared harness: %v", err)
	}
	if startCalls != 1 {
		t.Fatalf("expected one owned startup, got %d", startCalls)
	}
	if harness != want {
		t.Fatal("expected StartShared to return the owned harness")
	}
	if !harness.shared {
		t.Fatal("expected owned harness returned by StartShared to be marked shared")
	}
}

func TestCreateDatabaseRejectsOmittedServiceBeforeAcquisition(t *testing.T) {
	t.Setenv(suiteservices.HarnessServiceDependenciesEnv, "object_store")

	oldCreate := createDatabaseFn
	createCalls := 0
	createDatabaseFn = func(context.Context, string, string, string) error {
		createCalls++
		return nil
	}
	t.Cleanup(func() {
		createDatabaseFn = oldCreate
	})

	_, err := (&Harness{}).newDatabase(context.Background(), "guarded", suiteservices.FixtureReusePerTest, fixtureAttribution{})
	var dependencyErr *suiteservices.ServiceDependencyError
	if !errors.As(err, &dependencyErr) || dependencyErr.Service != "postgres" || dependencyErr.Reason != "omitted" {
		t.Fatalf("unexpected dependency error: %#v", err)
	}
	if createCalls != 0 {
		t.Fatalf("postgres acquisition ran before the dependency guard: calls=%d", createCalls)
	}
}

func TestPostgresSetupEnvelopeNormalizesReadinessAndCancellation(t *testing.T) {
	readiness := postgresSetupEnvelope("start", &postgresReadinessError{
		Attempts:        17,
		LastErr:         context.DeadlineExceeded,
		DeadlineExpired: true,
	})
	if readiness.FailureClass != "infra" || readiness.FailureReason != "service_readiness_timeout" ||
		readiness.Service != "postgres" || readiness.ReadinessStage != "start" || readiness.AttemptCount != 17 {
		t.Fatalf("unexpected postgres readiness envelope: %#v", readiness)
	}

	cancelled := postgresSetupEnvelope("start", context.Canceled)
	if cancelled.FailureClass != "interrupted" || cancelled.FailureReason != "cancelled_or_interrupted" {
		t.Fatalf("unexpected postgres cancellation envelope: %#v", cancelled)
	}
}

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

func TestPostgresContainerWaitStrategyOnlyWaitsForPortMapping(t *testing.T) {
	strategy := postgresPortWaitStrategy()
	if got := strategy.String(); !strings.Contains(got, "to be mapped") {
		t.Fatalf("expected mapped-port-only wait strategy, got %q", got)
	}
	if timeout := strategy.Timeout(); timeout == nil || *timeout != postgresPortMappingTimeout {
		t.Fatalf("unexpected port mapping timeout: got %v want %v", timeout, postgresPortMappingTimeout)
	}
}

func TestOwnedPostgresAppliesContainerLabels(t *testing.T) {
	stubOwnedPostgresStartup(t)

	var gotRequest testcontainers.ContainerRequest
	startContainerFn = func(ctx context.Context, req testcontainers.GenericContainerRequest) (testcontainers.Container, error) {
		gotRequest = req.ContainerRequest
		return fakePostgresContainer{
			host: "127.0.0.1",
			port: network.MustParsePort("5432/tcp"),
		}, nil
	}
	waitReadyFn = func(context.Context, *Harness) error { return nil }

	labels := map[string]string{"cartulary.test": "suite"}
	if _, err := StartOwnedWithLabels(context.Background(), labels); err != nil {
		t.Fatalf("start labeled postgres: %v", err)
	}
	if gotRequest.Labels["cartulary.test"] != "suite" {
		t.Fatalf("container labels not applied: %#v", gotRequest.Labels)
	}
	if gotRequest.Image != postgresImage || ContainerImage() != postgresImage {
		t.Fatalf("postgres image mismatch: request=%q exported=%q want=%q", gotRequest.Image, ContainerImage(), postgresImage)
	}
	if gotRequest.Env["PGDATA"] != "/var/lib/postgresql/18/docker" ||
		gotRequest.Env["POSTGRES_INITDB_ARGS"] != "--data-checksums --auth-host=scram-sha-256" {
		t.Fatalf("postgres initialization environment mismatch: %#v", gotRequest.Env)
	}
}

func TestOwnedPostgresRetriesReadinessTimeoutAndTerminatesFailedAttempt(t *testing.T) {
	stubOwnedPostgresStartup(t)

	starts := 0
	terminations := 0
	readinessChecks := 0
	var events []testcontainersx.StartEvent
	ports := []network.Port{
		network.MustParsePort("5433/tcp"),
		network.MustParsePort("5434/tcp"),
	}
	startContainerFn = func(ctx context.Context, req testcontainers.GenericContainerRequest) (testcontainers.Container, error) {
		starts++
		return fakePostgresContainer{
			host: "127.0.0.1",
			port: ports[starts-1],
			terminate: func(context.Context) error {
				terminations++
				return nil
			},
		}, nil
	}
	waitReadyFn = func(ctx context.Context, harness *Harness) error {
		readinessChecks++
		if readinessChecks == 1 {
			return &postgresReadinessError{
				LastErr:         context.DeadlineExceeded,
				DeadlineExpired: true,
			}
		}
		return nil
	}

	harness, err := StartOwnedWithOptions(context.Background(), StartOptions{
		Observer: func(event testcontainersx.StartEvent) {
			events = append(events, event)
		},
	})
	if err != nil {
		t.Fatalf("expected retry success, got %v", err)
	}
	if starts != 2 {
		t.Fatalf("expected two container attempts, got %d", starts)
	}
	if readinessChecks != 2 {
		t.Fatalf("expected two readiness checks, got %d", readinessChecks)
	}
	if terminations != 1 {
		t.Fatalf("expected failed readiness attempt to terminate its container once, got %d", terminations)
	}
	if harness.Port != "5434" {
		t.Fatalf("expected second attempt port, got %q", harness.Port)
	}
	if !observedRetry(events) {
		t.Fatalf("expected observer to record retryable attempt and retry decision, got %#v", events)
	}
}

func TestOwnedPostgresDoesNotRetryAuthenticationReadinessFailure(t *testing.T) {
	stubOwnedPostgresStartup(t)

	starts := 0
	terminations := 0
	startContainerFn = func(ctx context.Context, req testcontainers.GenericContainerRequest) (testcontainers.Container, error) {
		starts++
		return fakePostgresContainer{
			host: "127.0.0.1",
			port: network.MustParsePort("5432/tcp"),
			terminate: func(context.Context) error {
				terminations++
				return nil
			},
		}, nil
	}
	waitReadyFn = func(ctx context.Context, harness *Harness) error {
		return &postgresReadinessError{
			LastErr:         errors.New("password authentication failed for user cartulary"),
			DeadlineExpired: true,
		}
	}

	_, err := StartOwnedWithOptions(context.Background(), StartOptions{})
	if err == nil {
		t.Fatal("expected authentication readiness failure")
	}
	if starts != 1 {
		t.Fatalf("expected no retry for auth failure, got %d starts", starts)
	}
	if terminations != 1 {
		t.Fatalf("expected failed auth attempt to terminate its container once, got %d", terminations)
	}

	var startFailure *testcontainersx.StartFailure
	if !errors.As(err, &startFailure) {
		t.Fatalf("expected StartFailure, got %T", err)
	}
	if startFailure.Retryable {
		t.Fatal("auth readiness failure must not be retryable")
	}
	if startFailure.AttemptsStarted != 1 || startFailure.MaxAttempts != 3 {
		t.Fatalf("unexpected attempts: got %d/%d", startFailure.AttemptsStarted, startFailure.MaxAttempts)
	}
}

func TestPostgresWaitReadyRespectsContextCancellation(t *testing.T) {
	oldPing := pingAdminDSNFn
	t.Cleanup(func() {
		pingAdminDSNFn = oldPing
	})
	pingAdminDSNFn = func(ctx context.Context, dsn string) error {
		return ctx.Err()
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	start := time.Now()
	err := (&Harness{adminDSN: "postgres://cartulary:cartulary@127.0.0.1:5432/postgres?sslmode=disable"}).WaitReady(ctx)
	if err == nil {
		t.Fatal("expected context cancellation readiness error")
	}
	if elapsed := time.Since(start); elapsed > 100*time.Millisecond {
		t.Fatalf("WaitReady did not respect context cancellation promptly: %v", elapsed)
	}
	var readinessErr *postgresReadinessError
	if !errors.As(err, &readinessErr) {
		t.Fatalf("expected postgresReadinessError, got %T", err)
	}
	if readinessErr.DeadlineExpired {
		t.Fatal("context cancellation must not be marked as deadline expiry")
	}
}

func TestPrepareDatabaseTemplateModeClonesWithoutMigrationReplay(t *testing.T) {
	t.Setenv(suiteservices.SuiteIDEnv, "suite-template-mode")
	t.Setenv(suiteservices.TargetEnv, "backend-store")
	t.Setenv("CARTULARY_TEST_RESULTS_DIR", t.TempDir())
	t.Setenv("CARTULARY_TEST_RUN_ID", "template-mode")

	oldCreate := createDatabaseFn
	oldMigrate := migrateDatabaseFn
	t.Cleanup(func() {
		createDatabaseFn = oldCreate
		migrateDatabaseFn = oldMigrate
	})

	var createCalls []struct {
		Name     string
		Template string
	}
	createDatabaseFn = func(ctx context.Context, adminDSN string, name string, templateDB string) error {
		createCalls = append(createCalls, struct {
			Name     string
			Template string
		}{
			Name:     name,
			Template: templateDB,
		})
		return nil
	}

	migrateCalls := 0
	migrateDatabaseFn = func(ctx context.Context, db *sql.DB, source *database_migrations.Source) error {
		migrateCalls++
		return nil
	}

	harness := &Harness{
		adminDSN:    "postgres://cartulary:cartulary@127.0.0.1:5432/postgres?sslmode=disable",
		dsnTemplate: "postgres://cartulary:cartulary@127.0.0.1:5432/{database}?sslmode=disable",
		templateDB:  "suite_template",
		suiteHash:   "suitehash",
		processHash: "procaaaa",
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

func TestPostgresFixturePolicyResolutionUsesTopLevelTestAndPackage(t *testing.T) {
	t.Setenv(postgresFixturePolicyTestsEnv, "TestPostgresFixturePolicyResolutionUsesTopLevelTestAndPackage=template_clone")
	t.Setenv(postgresFixturePolicyPackagesEnv, "internal/modules/auth=transaction")
	t.Setenv(postgresFixturePolicyDefaultEnv, "transaction")

	policy := resolvePostgresFixturePolicy(fixtureAttribution{
		TestName:      "TestPostgresFixturePolicyResolutionUsesTopLevelTestAndPackage/subcase",
		CallerPackage: "internal/modules/auth",
	})
	if policy != postgresFixturePolicyTemplateClone {
		t.Fatalf("expected top-level test policy to win, got %q", policy)
	}
	t.Setenv(postgresFixturePolicyTestsEnv, "")
	policy = resolvePostgresFixturePolicy(fixtureAttribution{
		TestName:      "TestOther/subcase",
		CallerPackage: "./internal/modules/auth",
	})
	if policy != postgresFixturePolicyTransaction {
		t.Fatalf("expected package policy to win, got %q", policy)
	}

	t.Setenv(postgresFixturePolicyPackagesEnv, "")
	t.Setenv(postgresFixturePolicyDefaultEnv, "")
	policy = resolvePostgresFixturePolicy(fixtureAttribution{
		TestName:      "TestOther/subcase",
		CallerPackage: "internal/modules/timeline",
	})
	if policy != "" {
		t.Fatalf("expected no implicit fixture policy, got %q", policy)
	}
}

func TestValidateSelectedPostgresFixturePolicy(t *testing.T) {
	if err := validateSelectedPostgresFixturePolicy(postgresFixturePolicyTemplateClone, postgresFixturePolicyTemplateClone); err != nil {
		t.Fatalf("matching explicit policy failed: %v", err)
	}
	if err := validateSelectedPostgresFixturePolicy("", postgresFixturePolicyTemplateClone); err == nil {
		t.Fatal("expected every fixture helper to require an explicit policy")
	}
}

func TestPrepareIsolatedDatabaseTUsesFreshTemplateClones(t *testing.T) {
	t.Setenv(suiteservices.SuiteIDEnv, "")
	t.Setenv(postgresFixturePolicyDefaultEnv, postgresFixturePolicyTemplateClone)

	oldCreate := createDatabaseFn
	oldDrop := dropOwnedDatabaseFn
	t.Cleanup(func() {
		createDatabaseFn = oldCreate
		dropOwnedDatabaseFn = oldDrop
	})

	var creates []struct {
		Name     string
		Template string
	}
	createDatabaseFn = func(ctx context.Context, adminDSN string, name string, templateDB string) error {
		creates = append(creates, struct {
			Name     string
			Template string
		}{Name: name, Template: templateDB})
		return nil
	}
	var drops []string
	dropOwnedDatabaseFn = func(ctx context.Context, adminDSN string, name string, normalLimit time.Duration, forcedLimit time.Duration) (bool, error) {
		drops = append(drops, name)
		return false, nil
	}
	harness := &Harness{
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

func TestPrepareIsolatedDatabaseTCleanupDropsStandaloneDatabase(t *testing.T) {
	t.Setenv(suiteservices.SuiteIDEnv, "")
	t.Setenv(suiteservices.HarnessServiceDependenciesEnv, "postgres")
	t.Setenv(postgresFixturePolicyDefaultEnv, postgresFixturePolicyTemplateClone)

	oldCreate := createDatabaseFn
	oldDrop := dropOwnedDatabaseFn
	t.Cleanup(func() {
		createDatabaseFn = oldCreate
		dropOwnedDatabaseFn = oldDrop
	})

	var dropped []string
	createDatabaseFn = func(ctx context.Context, adminDSN string, name string, templateDB string) error {
		return nil
	}
	dropOwnedDatabaseFn = func(ctx context.Context, adminDSN string, name string, normalLimit time.Duration, forcedLimit time.Duration) (bool, error) {
		dropped = append(dropped, name)
		return false, nil
	}

	harness := &Harness{
		adminDSN:    "postgres://cartulary:cartulary@127.0.0.1:5432/postgres?sslmode=disable",
		dsnTemplate: "postgres://cartulary:cartulary@127.0.0.1:5432/{database}?sslmode=disable",
		templateDB:  "suite_template",
		suiteHash:   "suitehash",
		processHash: "procaaaa",
	}

	var preparedName string
	t.Run("prepare", func(t *testing.T) {
		preparedName = harness.PrepareIsolatedDatabaseT(t, "standalone-cleanup").Name
	})

	if len(dropped) != 1 || dropped[0] != preparedName {
		t.Fatalf("expected standalone PrepareIsolatedDatabaseT cleanup to drop %q, got %v", preparedName, dropped)
	}
}

func TestDatabaseCleanupUsesBoundedForcedFallbackOnlyForOwnedActiveDatabase(t *testing.T) {
	oldDrop := dropOwnedDatabaseFn
	t.Cleanup(func() {
		dropOwnedDatabaseFn = oldDrop
	})

	harness := &Harness{
		adminDSN: "postgres://cartulary:cartulary@127.0.0.1:5432/postgres?sslmode=disable",
		ownedDatabases: map[string]struct{}{
			"ct_deadbeef_cafe0001_000001_owned": {},
		},
	}
	dropCalls := 0
	dropOwnedDatabaseFn = func(ctx context.Context, _ string, _ string, normalLimit time.Duration, forcedLimit time.Duration) (bool, error) {
		dropCalls++
		if ctx.Err() != nil {
			t.Fatalf("cleanup delegate received canceled parent: %v", ctx.Err())
		}
		if normalLimit != postgresDatabaseNormalDropTimeout || forcedLimit != postgresDatabaseForcedDropTimeout {
			t.Fatalf("cleanup limits = (%s, %s), want (%s, %s)", normalLimit, forcedLimit, postgresDatabaseNormalDropTimeout, postgresDatabaseForcedDropTimeout)
		}
		return true, nil
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

func TestDatabaseCleanupUsesFreshForcedFallbackAfterBoundedNormalTimeout(t *testing.T) {
	oldDrop := dropOwnedDatabaseFn
	t.Cleanup(func() {
		dropOwnedDatabaseFn = oldDrop
	})

	const databaseName = "ct_deadbeef_cafe0001_000001_owned"
	harness := &Harness{
		adminDSN: "postgres://cartulary:cartulary@127.0.0.1:5432/postgres?sslmode=disable",
		ownedDatabases: map[string]struct{}{
			databaseName: {},
		},
	}
	dropCalls := 0
	dropOwnedDatabaseFn = func(ctx context.Context, _ string, _ string, normalLimit time.Duration, forcedLimit time.Duration) (bool, error) {
		dropCalls++
		if ctx.Err() != nil {
			t.Fatalf("coordinated cleanup inherited an expired child context: %v", ctx.Err())
		}
		if normalLimit != postgresDatabaseNormalDropTimeout || forcedLimit != postgresDatabaseForcedDropTimeout {
			t.Fatalf("coordinated cleanup limits = (%s, %s)", normalLimit, forcedLimit)
		}
		return true, nil
	}

	if err := harness.dropDatabase(context.Background(), databaseName, suiteservices.FixtureReusePerTest, fixtureAttribution{}); err != nil {
		t.Fatalf("drop owned database after normal timeout: %v", err)
	}
	if dropCalls != 1 {
		t.Fatalf("coordinated cleanup calls = %d, want 1", dropCalls)
	}
}

func TestDatabaseCleanupDoesNotForceNonConnectionFailure(t *testing.T) {
	authorizePGSuite(t)
	t.Setenv(suiteservices.SuiteIDEnv, "suite-cleanup-failure")
	t.Setenv(suiteservices.TargetEnv, "backend-integration")
	t.Setenv("CARTULARY_TEST_RESULTS_DIR", t.TempDir())
	t.Setenv("CARTULARY_TEST_RUN_ID", "cleanup-failure")

	oldDrop := dropOwnedDatabaseFn
	t.Cleanup(func() {
		dropOwnedDatabaseFn = oldDrop
	})

	const databaseName = "ct_deadbeef_cafe0001_000001_owned"
	harness := &Harness{ownedDatabases: map[string]struct{}{databaseName: {}}}
	if err := suiteservices.RecordEvent(nil, suiteservices.Event{Type: suiteservices.EventPostgresDBCreated, Name: databaseName, Kind: "template-clone"}); err != nil {
		t.Fatalf("record owned database: %v", err)
	}
	dropCalls := 0
	dropOwnedDatabaseFn = func(context.Context, string, string, time.Duration, time.Duration) (bool, error) {
		dropCalls++
		return false, errors.New("permission denied")
	}
	if err := harness.dropDatabase(context.Background(), databaseName, suiteservices.FixtureReusePerTest, fixtureAttribution{}); err == nil {
		t.Fatal("expected ordinary cleanup failure")
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

	dropOwnedDatabaseFn = func(context.Context, string, string, time.Duration, time.Duration) (bool, error) {
		dropCalls++
		return false, context.Canceled
	}
	if err := harness.dropDatabase(context.Background(), databaseName, suiteservices.FixtureReusePerTest, fixtureAttribution{}); !errors.Is(err, context.Canceled) {
		t.Fatalf("interrupted cleanup error = %v, want context cancellation", err)
	}
	if dropCalls != 2 {
		t.Fatalf("interrupted cleanup delegated %d total times", dropCalls)
	}
}

func TestPrepareIsolatedDatabaseTCleanupDropsAttachedSuiteTemplateClone(t *testing.T) {
	authorizePGSuite(t)
	t.Setenv(postgresFixturePolicyDefaultEnv, postgresFixturePolicyTemplateClone)
	t.Setenv(suiteservices.SuiteIDEnv, "suite-retained-template")
	t.Setenv(suiteservices.TargetEnv, "backend-integration")
	t.Setenv("CARTULARY_TEST_RESULTS_DIR", t.TempDir())
	t.Setenv("CARTULARY_TEST_RUN_ID", "retained-template")

	oldCreate := createDatabaseFn
	oldDrop := dropOwnedDatabaseFn
	t.Cleanup(func() {
		createDatabaseFn = oldCreate
		dropOwnedDatabaseFn = oldDrop
	})

	createDatabaseFn = func(ctx context.Context, adminDSN string, name string, templateDB string) error {
		return nil
	}
	dropCalls := 0
	dropOwnedDatabaseFn = func(ctx context.Context, adminDSN string, name string, normalLimit time.Duration, forcedLimit time.Duration) (bool, error) {
		dropCalls++
		return false, nil
	}

	harness := &Harness{
		adminDSN:    "postgres://cartulary:cartulary@127.0.0.1:5432/postgres?sslmode=disable",
		dsnTemplate: "postgres://cartulary:cartulary@127.0.0.1:5432/{database}?sslmode=disable",
		templateDB:  "suite_template",
		suiteHash:   "suitehash",
		processHash: "procaaaa",
		attached:    true,
	}

	var preparedName string
	t.Run("prepare", func(t *testing.T) {
		preparedName = harness.PrepareIsolatedDatabaseT(t, "attached-cleanup").Name
	})

	if dropCalls != 1 {
		t.Fatalf("expected attached suite template cleanup to drop %q, got %d drops", preparedName, dropCalls)
	}

	t.Setenv(suiteservices.PersistentBorrowerEnv, "1")
	t.Run("persistent session borrower", func(t *testing.T) {
		harness.PrepareIsolatedDatabaseT(t, "persistent-borrower-cleanup")
	})
	if dropCalls != 2 {
		t.Fatalf("expected a persistent session borrower to use the same eager cleanup path, got %d drops", dropCalls)
	}
}

func TestMigrationDatabaseTCleanupDropsStandaloneScratchDatabase(t *testing.T) {
	t.Setenv(suiteservices.SuiteIDEnv, "")
	t.Setenv(suiteservices.HarnessServiceDependenciesEnv, "postgres")

	oldCreate := createDatabaseFn
	oldDrop := dropOwnedDatabaseFn
	oldMigrate := migrateDatabaseFn
	oldOpen := openMigrationDatabaseFn
	t.Cleanup(func() {
		createDatabaseFn = oldCreate
		dropOwnedDatabaseFn = oldDrop
		migrateDatabaseFn = oldMigrate
		openMigrationDatabaseFn = oldOpen
	})

	var scratchName string
	var dropped []string
	createDatabaseFn = func(ctx context.Context, adminDSN string, name string, templateDB string) error {
		scratchName = name
		return nil
	}
	dropOwnedDatabaseFn = func(ctx context.Context, adminDSN string, name string, normalLimit time.Duration, forcedLimit time.Duration) (bool, error) {
		dropped = append(dropped, name)
		return false, nil
	}
	migrateDatabaseFn = func(ctx context.Context, db *sql.DB, source *database_migrations.Source) error {
		return nil
	}
	openMigrationDatabaseFn = func(string, postgres.Purpose) (*sql.DB, error) {
		return sql.Open("pgx", "postgres://fixture/unused")
	}

	harness := &Harness{
		adminDSN:    "postgres://cartulary:cartulary@127.0.0.1:5432/postgres?sslmode=disable",
		dsnTemplate: "postgres://cartulary:cartulary@127.0.0.1:5432/{database}?sslmode=disable",
		templateDB:  "suite_template",
		suiteHash:   "suitehash",
		processHash: "procaaaa",
	}

	t.Run("migrate scratch", func(t *testing.T) {
		harness.MigrationDatabaseT(t)
	})

	if len(dropped) != 1 || dropped[0] != scratchName {
		t.Fatalf("expected standalone migration scratch cleanup to drop %q, got %v", scratchName, dropped)
	}
}

func TestMigrationDatabaseTCleanupDropsAttachedSuiteScratchDatabase(t *testing.T) {
	authorizePGSuite(t)
	t.Setenv(suiteservices.SuiteIDEnv, "suite-retained-migration-scratch")
	t.Setenv(suiteservices.TargetEnv, "backend-integration")
	t.Setenv("CARTULARY_TEST_RESULTS_DIR", t.TempDir())
	t.Setenv("CARTULARY_TEST_RUN_ID", "retained-migration-scratch")

	oldCreate := createDatabaseFn
	oldDrop := dropOwnedDatabaseFn
	oldMigrate := applyMigrationsThrough
	oldOpen := openMigrationDatabaseFn
	t.Cleanup(func() {
		createDatabaseFn = oldCreate
		dropOwnedDatabaseFn = oldDrop
		applyMigrationsThrough = oldMigrate
		openMigrationDatabaseFn = oldOpen
	})

	var scratchName string
	dropCalls := 0
	createDatabaseFn = func(ctx context.Context, adminDSN string, name string, templateDB string) error {
		scratchName = name
		return nil
	}
	dropOwnedDatabaseFn = func(ctx context.Context, adminDSN string, name string, normalLimit time.Duration, forcedLimit time.Duration) (bool, error) {
		dropCalls++
		return false, nil
	}
	applyMigrationsThrough = func(ctx context.Context, db *sql.DB, version int64) error {
		return nil
	}
	openMigrationDatabaseFn = func(string, postgres.Purpose) (*sql.DB, error) {
		return sql.Open("pgx", "postgres://fixture/unused")
	}

	harness := &Harness{
		adminDSN:    "postgres://cartulary:cartulary@127.0.0.1:5432/postgres?sslmode=disable",
		dsnTemplate: "postgres://cartulary:cartulary@127.0.0.1:5432/{database}?sslmode=disable",
		templateDB:  "suite_template",
		suiteHash:   "suitehash",
		processHash: "procaaaa",
		attached:    true,
	}

	t.Run("migrate scratch", func(t *testing.T) {
		harness.MigrationDatabaseThroughT(t, 1)
	})

	if dropCalls != 1 {
		t.Fatalf("expected attached suite migration scratch cleanup to drop %q, got %d drops", scratchName, dropCalls)
	}
}

func TestMigrationDatabaseTargetedOperationValidation(t *testing.T) {
	hash, err := pgschema.Hash()
	if err != nil {
		t.Fatalf("hash canonical migration catalog: %v", err)
	}
	const wantHash = "ba0bffb76a7193e9616f4fdeee9e16086da7f3c0fe67f5dc26e47ff5f1b750c5"
	if hash != wantHash {
		t.Fatalf("canonical migration schema hash = %s, want %s", hash, wantHash)
	}

	oldApply := applyMigrationsThrough
	oldRollback := rollbackMigrationsThrough
	t.Cleanup(func() {
		applyMigrationsThrough = oldApply
		rollbackMigrationsThrough = oldRollback
	})

	applyCalls := 0
	rollbackCalls := 0
	applyMigrationsThrough = func(context.Context, *sql.DB, int64) error {
		applyCalls++
		return nil
	}
	rollbackMigrationsThrough = func(context.Context, *sql.DB, int64) error {
		rollbackCalls++
		return nil
	}

	capabilityType := reflect.TypeOf(MigrationDatabase{})
	for index := 0; index < capabilityType.NumField(); index++ {
		if capabilityType.Field(index).PkgPath == "" {
			t.Fatalf("migration capability field %q must be unexported", capabilityType.Field(index).Name)
		}
	}
	methodType := reflect.TypeOf((*MigrationDatabase)(nil))
	if methodType.NumMethod() != 3 {
		t.Fatalf("migration capability method count = %d, want 3", methodType.NumMethod())
	}
	for index, name := range []string{"ApplyThrough", "RollbackThrough", "SQL"} {
		if got := methodType.Method(index).Name; got != name {
			t.Fatalf("migration capability method %d = %q, want %q", index, got, name)
		}
	}

	var zero MigrationDatabase
	if err := zero.ApplyThrough(context.Background(), 0); err == nil || err.Error() != "migration apply-through version must be positive" {
		t.Fatalf("unexpected apply target validation error: %v", err)
	}
	if err := zero.RollbackThrough(context.Background(), -1); err == nil || err.Error() != "migration rollback-through version must be non-negative" {
		t.Fatalf("unexpected rollback target validation error: %v", err)
	}
	if applyCalls != 0 || rollbackCalls != 0 {
		t.Fatalf("invalid targets reached migration access: apply=%d rollback=%d", applyCalls, rollbackCalls)
	}
	if zero.SQL() != nil {
		t.Fatal("zero capability exposed a database handle")
	}
	if err := zero.ApplyThrough(context.Background(), 1); err == nil || err.Error() != "migration database capability is not harness-issued" {
		t.Fatalf("unexpected zero capability error: %v", err)
	}

	issued := &MigrationDatabase{db: &sql.DB{}, identity: issuedMigrationDatabaseIdentity}
	if issued.SQL() == nil {
		t.Fatal("issued capability did not expose its scratch handle")
	}
	if err := issued.ApplyThrough(context.Background(), 1); err != nil {
		t.Fatalf("apply through issued capability: %v", err)
	}
	if err := issued.RollbackThrough(context.Background(), 0); err != nil {
		t.Fatalf("rollback through zero issued capability: %v", err)
	}
	if applyCalls != 1 || rollbackCalls != 1 {
		t.Fatalf("unexpected targeted call counts: apply=%d rollback=%d", applyCalls, rollbackCalls)
	}
}

func TestHarnessStartsPostgresAndRunsCurrentMigrationPath(t *testing.T) {
	harness := Start(t)

	testDB, err := harness.PrepareDatabase(context.Background(), "bootstrap")
	if err != nil {
		t.Fatalf("prepare database: %v", err)
	}
	defer func() {
		if err := harness.DropDatabase(context.Background(), testDB.Name); err != nil {
			t.Fatalf("drop database: %v", err)
		}
	}()

	if testDB.DSN == "" {
		t.Fatal("expected database dsn")
	}
}

func TestPrepareIsolatedDatabaseTReturnsMigratedDatabase(t *testing.T) {
	harness := Start(t)
	testDB := harness.PrepareIsolatedDatabaseT(t, "bootstrap_t")

	db, err := sql.Open("pgx", testDB.DSN)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	var exists bool
	if err := db.QueryRowContext(context.Background(), `
SELECT EXISTS (
    SELECT 1
    FROM information_schema.tables
    WHERE table_schema = 'public' AND table_name = 'users'
)`).Scan(&exists); err != nil {
		t.Fatalf("query users table: %v", err)
	}
	if !exists {
		t.Fatal("expected PrepareIsolatedDatabaseT to return a migrated database")
	}
}

func TestBeginRollbackDBTIsolatesRowsPerTransaction(t *testing.T) {
	t.Setenv(postgresFixturePolicyDefaultEnv, postgresFixturePolicyTransaction)
	t.Setenv(suiteservices.SuiteIDEnv, "rollback-db")
	t.Setenv(suiteservices.TargetEnv, "backend-store")
	t.Setenv("CARTULARY_TEST_RESULTS_DIR", t.TempDir())
	t.Setenv("CARTULARY_TEST_RUN_ID", "rollback-db")

	harness := Start(t)

	t.Run("insert rows inside rollback fixture", func(t *testing.T) {
		db := harness.BeginRollbackDBT(t, "rollback-db")
		if _, err := db.Exec(context.Background(), `INSERT INTO users (email, display_name, password_hash, mfa_required) VALUES ($1, $2, $3, false)`, "rollback@example.test", "Rollback", "hash"); err != nil {
			t.Fatalf("seed rollback transaction: %v", err)
		}
	})

	t.Run("next rollback fixture starts clean", func(t *testing.T) {
		db := harness.BeginRollbackDBT(t, "rollback-db")
		var count int
		if err := db.QueryRow(context.Background(), `SELECT COUNT(*) FROM users WHERE email = $1`, "rollback@example.test").Scan(&count); err != nil {
			t.Fatalf("count rollback rows: %v", err)
		}
		if count != 0 {
			t.Fatalf("expected rollback fixture to discard rows, got %d", count)
		}
	})

	scope, ok, err := suiteservices.Summarize(nil)
	if err != nil {
		t.Fatalf("summarize suite service events: %v", err)
	}
	if !ok {
		t.Fatal("expected suite service summary")
	}
	transactionCount := 0
	for _, activity := range scope.Fixture.ByStrategy {
		if activity.Operation == "database-reset" {
			t.Fatalf("rollback fixture must not emit database-reset activity: %#v", activity)
		}
		if activity.Operation == "transaction" && activity.FixturePolicy == postgresFixturePolicyTransaction {
			transactionCount += activity.Count
		}
	}
	if transactionCount != 2 {
		t.Fatalf("expected two rollback transaction events, got %d in %#v", transactionCount, scope.Fixture.ByStrategy)
	}
}

func resetSharedHarness(t testing.TB) {
	t.Helper()

	sharedHarnessMu.Lock()
	sharedHarness = nil
	sharedHarnessMu.Unlock()

	t.Cleanup(func() {
		sharedHarnessMu.Lock()
		sharedHarness = nil
		sharedHarnessMu.Unlock()
	})
}

func authorizePGSuite(t *testing.T) {
	t.Helper()
	t.Setenv(suiteservices.HarnessServiceDependenciesEnv, "postgres")
	t.Setenv(suiteservices.CallModeEnv, "owned")
	t.Setenv(suiteservices.SuiteRuntimeRootEnv, "/private/pgtest-suite-runtime")
	t.Setenv(suiteservices.SuiteRuntimeLeaseIDEnv, "00000000-0000-4000-8000-000000000001")
	t.Setenv(suiteservices.SuiteRuntimeRunIDEnv, "pgtest-suite")
}

func stubOwnedPostgresStartup(t testing.TB) {
	t.Helper()

	oldStartContainer := startContainerFn
	oldWaitReady := waitReadyFn
	oldPreflight := startPreflightFn
	oldSleep := startSleepFn
	t.Cleanup(func() {
		startContainerFn = oldStartContainer
		waitReadyFn = oldWaitReady
		startPreflightFn = oldPreflight
		startSleepFn = oldSleep
	})

	startPreflightFn = func(context.Context) (string, error) {
		return "unix:///var/run/docker.sock", nil
	}
	startSleepFn = func(context.Context, time.Duration) error {
		return nil
	}
}

func observedRetry(events []testcontainersx.StartEvent) bool {
	sawRetryableAttempt := false
	sawRetryScheduled := false
	for _, event := range events {
		if event.Type == testcontainersx.StartEventAttemptEnd && event.Attempt == 1 && event.Retryable && event.Status == "fail" {
			sawRetryableAttempt = true
		}
		if event.Type == testcontainersx.StartEventRetryScheduled && event.Attempt == 1 {
			sawRetryScheduled = true
		}
	}
	return sawRetryableAttempt && sawRetryScheduled
}

type fakePostgresContainer struct {
	testcontainers.Container
	host      string
	port      network.Port
	terminate func(context.Context) error
}

func (c fakePostgresContainer) Host(context.Context) (string, error) {
	return c.host, nil
}

func (c fakePostgresContainer) MappedPort(context.Context, string) (network.Port, error) {
	return c.port, nil
}

func (c fakePostgresContainer) Terminate(ctx context.Context, opts ...testcontainers.TerminateOption) error {
	if c.terminate == nil {
		return nil
	}
	return c.terminate(ctx)
}
