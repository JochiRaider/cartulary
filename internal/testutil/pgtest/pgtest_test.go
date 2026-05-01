package pgtest

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/moby/moby/api/types/network"
	testcontainers "github.com/testcontainers/testcontainers-go"

	dbmigrations "github.com/JochiRaider/cartulary/db/migrations"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
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

	var gotLabels map[string]string
	startContainerFn = func(ctx context.Context, req testcontainers.GenericContainerRequest) (testcontainers.Container, error) {
		gotLabels = req.Labels
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
	if gotLabels["cartulary.test"] != "suite" {
		t.Fatalf("container labels not applied: %#v", gotLabels)
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
	migrateDatabaseFn = func(ctx context.Context, db *sql.DB, source postgres.MigrationSource, command string, args ...string) (postgres.MigrationStatus, error) {
		migrateCalls++
		return postgres.MigrationStatus{Command: command, Directory: source.Name}, nil
	}

	harness := &Harness{
		adminDSN:    "postgres://cartulary:cartulary@127.0.0.1:5432/postgres?sslmode=disable",
		dsnTemplate: "postgres://cartulary:cartulary@127.0.0.1:5432/{database}?sslmode=disable",
		templateDB:  "suite_template",
		suiteHash:   "suitehash",
		processHash: "procaaaa",
	}

	testDB, status, err := harness.PrepareDatabase(context.Background(), "bootstrap")
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
	if !status.TemplateClone {
		t.Fatal("expected migration status to mark template clone mode")
	}
	if status.Directory != dbmigrations.RepositoryPath {
		t.Fatalf("unexpected migration directory in status: got %q want %q", status.Directory, dbmigrations.RepositoryPath)
	}
	if status.TemplateDatabase != "suite_template" {
		t.Fatalf("unexpected template database in status: got %q", status.TemplateDatabase)
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
	t.Setenv(postgresFixturePolicyPackagesEnv, "internal/modules/auth=package_reset")
	t.Setenv(postgresFixturePolicyDefaultEnv, "package_reset")
	t.Setenv(postgresResetTablesTestsEnv, "TestPostgresFixturePolicyResolutionUsesTopLevelTestAndPackage=users|sessions")
	t.Setenv(postgresResetTablesPackagesEnv, "internal/modules/auth=audit_events|route_idempotency")

	policy := resolvePostgresFixturePolicy(fixtureAttribution{
		TestName:      "TestPostgresFixturePolicyResolutionUsesTopLevelTestAndPackage/subcase",
		CallerPackage: "internal/modules/auth",
	})
	if policy != postgresFixturePolicyTemplateClone {
		t.Fatalf("expected top-level test policy to win, got %q", policy)
	}
	tables := resolvePostgresResetTables(fixtureAttribution{
		TestName:      "TestPostgresFixturePolicyResolutionUsesTopLevelTestAndPackage/subcase",
		CallerPackage: "internal/modules/auth",
	})
	if got, want := strings.Join(tables, ","), "users,sessions"; got != want {
		t.Fatalf("expected top-level reset tables to win, got %q want %q", got, want)
	}

	t.Setenv(postgresFixturePolicyTestsEnv, "")
	t.Setenv(postgresResetTablesTestsEnv, "")
	policy = resolvePostgresFixturePolicy(fixtureAttribution{
		TestName:      "TestOther/subcase",
		CallerPackage: "./internal/modules/auth",
	})
	if policy != postgresFixturePolicyPackageReset {
		t.Fatalf("expected package policy to win, got %q", policy)
	}
	tables = resolvePostgresResetTables(fixtureAttribution{
		TestName:      "TestOther/subcase",
		CallerPackage: "./internal/modules/auth",
	})
	if got, want := strings.Join(tables, ","), "audit_events,route_idempotency"; got != want {
		t.Fatalf("expected package reset tables to win, got %q want %q", got, want)
	}

	t.Setenv(postgresFixturePolicyPackagesEnv, "")
	t.Setenv(postgresFixturePolicyDefaultEnv, "")
	policy = resolvePostgresFixturePolicy(fixtureAttribution{
		TestName:      "TestOther/subcase",
		CallerPackage: "internal/modules/timeline",
	})
	if policy != postgresFixturePolicyTemplateClone {
		t.Fatalf("expected no-env fallback to use template clone, got %q", policy)
	}
}

func TestPrepareGroupDatabaseTReusesTemplateCloneForParentScopedGroup(t *testing.T) {
	t.Setenv(suiteservices.SuiteIDEnv, "")

	oldCreate := createDatabaseFn
	oldDrop := dropDatabaseFn
	t.Cleanup(func() {
		createDatabaseFn = oldCreate
		dropDatabaseFn = oldDrop
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
	dropDatabaseFn = func(ctx context.Context, adminDSN string, name string) error {
		drops = append(drops, name)
		return nil
	}

	harness := &Harness{
		adminDSN:    "postgres://cartulary:cartulary@127.0.0.1:5432/postgres?sslmode=disable",
		dsnTemplate: "postgres://cartulary:cartulary@127.0.0.1:5432/{database}?sslmode=disable",
		templateDB:  "suite_template",
		suiteHash:   "suitehash",
		processHash: "procaaaa",
	}

	var firstName string
	t.Run("group owner", func(t *testing.T) {
		first := harness.PrepareGroupDatabaseT(t, "group-clone", "bootstrap-state")
		second := harness.PrepareGroupDatabaseT(t, "group-clone", "bootstrap-state")
		firstName = first.Name
		if first.Name == "" || first.Name != second.Name {
			t.Fatalf("expected grouped clone reuse, got first=%q second=%q", first.Name, second.Name)
		}
	})
	if len(creates) != 1 {
		t.Fatalf("expected one grouped template clone, got %#v", creates)
	}
	if creates[0].Template != "suite_template" {
		t.Fatalf("expected grouped clone to use suite template, got %#v", creates[0])
	}
	if len(drops) != 1 || drops[0] != firstName {
		t.Fatalf("expected grouped clone cleanup to drop %q, got %v", firstName, drops)
	}
}

func TestPreparePackageDatabaseTTemplateClonePolicyAvoidsPackageReset(t *testing.T) {
	t.Setenv(suiteservices.SuiteIDEnv, "")
	t.Setenv(postgresFixturePolicyDefaultEnv, postgresFixturePolicyTemplateClone)

	oldCreate := createDatabaseFn
	oldDrop := dropDatabaseFn
	oldListMutableTables := listMutableTablesFn
	t.Cleanup(func() {
		createDatabaseFn = oldCreate
		dropDatabaseFn = oldDrop
		listMutableTablesFn = oldListMutableTables
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
	dropDatabaseFn = func(ctx context.Context, adminDSN string, name string) error {
		drops = append(drops, name)
		return nil
	}
	listMutableTablesFn = func(ctx context.Context, db *sql.DB) ([]string, error) {
		return nil, errors.New("template clone policy must not reset mutable tables")
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
		firstName = harness.PreparePackageDatabaseT(t, "clone-policy").Name
	})
	var secondName string
	t.Run("second clone", func(t *testing.T) {
		secondName = harness.PreparePackageDatabaseT(t, "clone-policy").Name
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

func TestPrepareDatabaseTCleanupDropsStandaloneDatabase(t *testing.T) {
	t.Setenv(suiteservices.SuiteIDEnv, "")

	oldCreate := createDatabaseFn
	oldDrop := dropDatabaseFn
	t.Cleanup(func() {
		createDatabaseFn = oldCreate
		dropDatabaseFn = oldDrop
	})

	var dropped []string
	createDatabaseFn = func(ctx context.Context, adminDSN string, name string, templateDB string) error {
		return nil
	}
	dropDatabaseFn = func(ctx context.Context, adminDSN string, name string) error {
		dropped = append(dropped, name)
		return nil
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
		preparedName = harness.PrepareDatabaseT(t, "standalone-cleanup").Name
	})

	if len(dropped) != 1 || dropped[0] != preparedName {
		t.Fatalf("expected standalone PrepareDatabaseT cleanup to drop %q, got %v", preparedName, dropped)
	}
}

func TestPrepareDatabaseTCleanupRetainsAttachedSuiteTemplateClone(t *testing.T) {
	t.Setenv(suiteservices.ActiveEnv, "1")
	t.Setenv(suiteservices.SuiteIDEnv, "suite-retained-template")
	t.Setenv(suiteservices.TargetEnv, "backend-integration")
	t.Setenv("CARTULARY_TEST_RESULTS_DIR", t.TempDir())
	t.Setenv("CARTULARY_TEST_RUN_ID", "retained-template")

	oldCreate := createDatabaseFn
	oldDrop := dropDatabaseFn
	t.Cleanup(func() {
		createDatabaseFn = oldCreate
		dropDatabaseFn = oldDrop
	})

	createDatabaseFn = func(ctx context.Context, adminDSN string, name string, templateDB string) error {
		return nil
	}
	dropCalls := 0
	dropDatabaseFn = func(ctx context.Context, adminDSN string, name string) error {
		dropCalls++
		return nil
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
		preparedName = harness.PrepareDatabaseT(t, "attached-cleanup").Name
	})

	if dropCalls != 0 {
		t.Fatalf("expected attached suite template cleanup to retain database, got %d drops", dropCalls)
	}

	scope, ok, err := suiteservices.Summarize(nil)
	if err != nil {
		t.Fatalf("summarize suite service events: %v", err)
	}
	if !ok {
		t.Fatal("expected suite service summary")
	}
	foundRetain := false
	for _, activity := range scope.Fixture.ByStrategy {
		if activity.Operation == "database-retain" && activity.ReuseScope == suiteservices.FixtureReusePerTest && activity.Strategy == suiteservices.PostgresPreparationTemplateClone {
			foundRetain = true
			break
		}
	}
	if !foundRetain {
		t.Fatalf("expected retained clone fixture activity for %q, got %#v", preparedName, scope.Fixture)
	}
}

func TestMigrationDatabaseTCleanupDropsStandaloneScratchDatabase(t *testing.T) {
	t.Setenv(suiteservices.SuiteIDEnv, "")

	oldCreate := createDatabaseFn
	oldDrop := dropDatabaseFn
	oldMigrate := migrateDatabaseFn
	t.Cleanup(func() {
		createDatabaseFn = oldCreate
		dropDatabaseFn = oldDrop
		migrateDatabaseFn = oldMigrate
	})

	var scratchName string
	var dropped []string
	createDatabaseFn = func(ctx context.Context, adminDSN string, name string, templateDB string) error {
		scratchName = name
		return nil
	}
	dropDatabaseFn = func(ctx context.Context, adminDSN string, name string) error {
		dropped = append(dropped, name)
		return nil
	}
	migrateDatabaseFn = func(ctx context.Context, db *sql.DB, source postgres.MigrationSource, command string, args ...string) (postgres.MigrationStatus, error) {
		return postgres.MigrationStatus{Command: command, Directory: source.Name}, nil
	}

	harness := &Harness{
		adminDSN:    "postgres://cartulary:cartulary@127.0.0.1:5432/postgres?sslmode=disable",
		dsnTemplate: "postgres://cartulary:cartulary@127.0.0.1:5432/{database}?sslmode=disable",
		templateDB:  "suite_template",
		suiteHash:   "suitehash",
		processHash: "procaaaa",
	}

	t.Run("migrate scratch", func(t *testing.T) {
		harness.MigrationDatabaseT(t, "standalone-scratch", "up")
	})

	if len(dropped) != 1 || dropped[0] != scratchName {
		t.Fatalf("expected standalone migration scratch cleanup to drop %q, got %v", scratchName, dropped)
	}
}

func TestMigrationDatabaseTCleanupRetainsAttachedSuiteScratchDatabase(t *testing.T) {
	t.Setenv(suiteservices.ActiveEnv, "1")
	t.Setenv(suiteservices.SuiteIDEnv, "suite-retained-migration-scratch")
	t.Setenv(suiteservices.TargetEnv, "backend-integration")
	t.Setenv("CARTULARY_TEST_RESULTS_DIR", t.TempDir())
	t.Setenv("CARTULARY_TEST_RUN_ID", "retained-migration-scratch")

	oldCreate := createDatabaseFn
	oldDrop := dropDatabaseFn
	oldMigrate := migrateDatabaseFn
	t.Cleanup(func() {
		createDatabaseFn = oldCreate
		dropDatabaseFn = oldDrop
		migrateDatabaseFn = oldMigrate
	})

	var scratchName string
	dropCalls := 0
	createDatabaseFn = func(ctx context.Context, adminDSN string, name string, templateDB string) error {
		scratchName = name
		return nil
	}
	dropDatabaseFn = func(ctx context.Context, adminDSN string, name string) error {
		dropCalls++
		return nil
	}
	migrateDatabaseFn = func(ctx context.Context, db *sql.DB, source postgres.MigrationSource, command string, args ...string) (postgres.MigrationStatus, error) {
		return postgres.MigrationStatus{Command: command, Directory: source.Name}, nil
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
		harness.MigrationDatabaseT(t, "attached-scratch", "up-to", "00001")
	})

	if dropCalls != 0 {
		t.Fatalf("expected attached suite migration scratch cleanup to retain database, got %d drops", dropCalls)
	}

	scope, ok, err := suiteservices.Summarize(nil)
	if err != nil {
		t.Fatalf("summarize suite service events: %v", err)
	}
	if !ok {
		t.Fatal("expected suite service summary")
	}
	foundRetain := false
	for _, activity := range scope.Fixture.ByStrategy {
		if activity.Service == suiteservices.ServicePostgres && activity.Operation == "database-retain" && activity.ReuseScope == suiteservices.FixtureReuseMigrationScratch {
			foundRetain = true
			break
		}
	}
	if !foundRetain {
		t.Fatalf("expected retained migration scratch fixture activity for %q, got %#v", scratchName, scope.Fixture)
	}
}

func TestHarnessStartsPostgresAndRunsCurrentMigrationPath(t *testing.T) {
	harness := Start(t)

	testDB, status, err := harness.PrepareDatabase(context.Background(), "bootstrap")
	if err != nil {
		t.Fatalf("prepare database: %v", err)
	}
	defer func() {
		if harness.retainPreparedDatabaseOnCleanup() {
			harness.recordRetainedDatabase(testDB.Name, suiteservices.FixtureReusePerTest, fixtureAttribution{})
			return
		}
		if err := harness.DropDatabase(context.Background(), testDB.Name); err != nil {
			t.Fatalf("drop database: %v", err)
		}
	}()

	if testDB.DSN == "" {
		t.Fatal("expected database dsn")
	}
	if status.Empty {
		t.Fatal("expected the current bootstrap migration set to include numbered migrations")
	}
}

func TestPrepareDatabaseTReturnsMigratedDatabase(t *testing.T) {
	harness := Start(t)
	testDB := harness.PrepareDatabaseT(t, "bootstrap_t")

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
		t.Fatal("expected PrepareDatabaseT to return a migrated database")
	}
}

func TestBeginRollbackDBTIsolatesRowsWithoutPackageReset(t *testing.T) {
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

func TestPreparePackageDatabaseTReusesAndResetsMutableTables(t *testing.T) {
	t.Setenv(postgresFixturePolicyDefaultEnv, postgresFixturePolicyPackageReset)
	harness := Start(t)

	oldListMutableTables := listMutableTablesFn
	listCalls := 0
	listMutableTablesFn = func(ctx context.Context, db *sql.DB) ([]string, error) {
		listCalls++
		return oldListMutableTables(ctx, db)
	}
	t.Cleanup(func() {
		listMutableTablesFn = oldListMutableTables
	})

	var firstName string
	t.Run("first use seeds rows", func(t *testing.T) {
		first := harness.PreparePackageDatabaseT(t, "package-reset")
		firstName = first.Name

		db, err := sql.Open("pgx", first.DSN)
		if err != nil {
			t.Fatalf("open package database: %v", err)
		}
		defer db.Close()

		if _, err := db.ExecContext(context.Background(), `INSERT INTO users (email, display_name, password_hash, mfa_required) VALUES ($1, $2, $3, false)`, "package-reset@example.test", "Package Reset", "hash"); err != nil {
			t.Fatalf("seed package database: %v", err)
		}
	})
	t.Run("second use resets rows", func(t *testing.T) {
		second := harness.PreparePackageDatabaseT(t, "package-reset")
		if second.Name != firstName {
			t.Fatalf("expected package database reuse, got %q want %q", second.Name, firstName)
		}
		db, err := sql.Open("pgx", second.DSN)
		if err != nil {
			t.Fatalf("open reused package database: %v", err)
		}
		defer db.Close()

		var userCount int
		if err := db.QueryRowContext(context.Background(), `SELECT COUNT(*) FROM users`).Scan(&userCount); err != nil {
			t.Fatalf("count reset users: %v", err)
		}
		if userCount != 0 {
			t.Fatalf("expected package reset to clear users, got %d", userCount)
		}

		var gooseCount int
		if err := db.QueryRowContext(context.Background(), `SELECT COUNT(*) FROM goose_db_version`).Scan(&gooseCount); err != nil {
			t.Fatalf("count goose versions: %v", err)
		}
		if gooseCount == 0 {
			t.Fatal("expected package reset to preserve migration metadata")
		}

		if _, err := db.ExecContext(context.Background(), `INSERT INTO users (email, display_name, password_hash, mfa_required) VALUES ($1, $2, $3, false)`, "package-reset-again@example.test", "Package Reset Again", "hash"); err != nil {
			t.Fatalf("seed package database for cached reset: %v", err)
		}
	})
	if listCalls != 1 {
		t.Fatalf("expected first package reset to discover mutable tables once, got %d calls", listCalls)
	}

	listMutableTablesFn = func(ctx context.Context, db *sql.DB) ([]string, error) {
		return nil, fmt.Errorf("cached reset should not rediscover mutable tables")
	}
	t.Run("third use reuses cached reset statement", func(t *testing.T) {
		third := harness.PreparePackageDatabaseT(t, "package-reset")
		if third.Name != firstName {
			t.Fatalf("expected package database reuse, got %q want %q", third.Name, firstName)
		}
		db, err := sql.Open("pgx", third.DSN)
		if err != nil {
			t.Fatalf("open cached-reset package database: %v", err)
		}
		defer db.Close()

		var userCount int
		if err := db.QueryRowContext(context.Background(), `SELECT COUNT(*) FROM users`).Scan(&userCount); err != nil {
			t.Fatalf("count users after cached reset: %v", err)
		}
		if userCount != 0 {
			t.Fatalf("expected cached package reset to clear users, got %d", userCount)
		}
	})
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
