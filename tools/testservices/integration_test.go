package main

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices"
)

func TestMakeBackendStoreUsesSingleOwnedSuitePair(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	requireDocker(t)

	scope, _ := runMakeAndLoadScopeWithEvents(t, "backend-store-direct", nil, "backend-store")
	if scope.Wrapper.OwnedCount != 1 {
		t.Fatalf("expected exactly one owned wrapper, got %#v", scope.Wrapper)
	}
	if scope.Wrapper.PassThroughCount != 0 {
		t.Fatalf("expected direct backend-store to avoid pass-through wrappers, got %#v", scope.Wrapper)
	}
	assertSuiteServicesStarted(t, scope)
	if scope.Cleanup.Status != "succeeded" {
		t.Fatalf("expected successful cleanup, got %#v", scope.Cleanup)
	}
	if scope.Postgres.MigratedDatabaseCount != 1 {
		t.Fatalf("expected one migrated template database, got %#v", scope.Postgres)
	}
	if scope.Postgres.TemplateCloneCount == 0 {
		t.Fatalf("expected backend-store to clone or reuse the migrated template database, got %#v", scope.Postgres)
	}
	assertPostgresPreparationStrategy(t, scope, suiteservices.PostgresPreparationTemplate)
	assertPostgresPreparationStrategy(t, scope, suiteservices.PostgresPreparationTemplateClone)
	if scope.Postgres.AttachedHarnessCount == 0 {
		t.Fatalf("expected attached postgres harness usage, got %#v", scope.Postgres)
	}
	if scope.Fixture.TotalCount == 0 {
		t.Fatalf("expected backend-store fixture diagnostics, got %#v", scope.Fixture)
	}
	assertNoHotPathPostgresDrops(t, scope)
	assertPostgresDatabaseResets(t, scope)
	assertPostgresFixturePolicy(t, scope, suiteservices.PostgresFixturePolicyPackageReset)
}

func TestMakeTestFastSharesSingleSuiteAcrossServiceBackedLanes(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	requireDocker(t)

	scope, events := runMakeAndLoadScopeWithEvents(t, "test-fast-suite", map[string]string{
		"SERVICE_BACKED_JOBS": "2",
	}, "test-fast")

	if scope.Wrapper.OwnedCount != 1 {
		t.Fatalf("expected exactly one owned wrapper, got %#v", scope.Wrapper)
	}
	if scope.Wrapper.PassThroughCount < 4 {
		t.Fatalf("expected nested target wrappers to pass through on the shared suite env, got %#v", scope.Wrapper)
	}
	assertSuiteServicesStarted(t, scope)
	if scope.Postgres.MigratedDatabaseCount < 1 {
		t.Fatalf("expected at least one migrated template database for the whole suite, got %#v", scope.Postgres)
	}
	if scope.Postgres.TemplateCloneCount == 0 {
		t.Fatalf("expected cloned package or isolated databases, got %#v", scope.Postgres)
	}
	assertPostgresPreparationStrategy(t, scope, suiteservices.PostgresPreparationTemplate)
	assertPostgresPreparationStrategy(t, scope, suiteservices.PostgresPreparationTemplateClone)
	if scope.MinIO.BucketCreateCount == 0 {
		t.Fatalf("expected minio bucket create activity, got %#v", scope.MinIO)
	}
	if scope.Fixture.TotalCount == 0 || len(scope.Fixture.ByPackage) == 0 {
		t.Fatalf("expected fixture diagnostics grouped by package, got %#v", scope.Fixture)
	}
	assertNoHotPathPostgresDrops(t, scope)
	assertPostgresFixturePolicy(t, scope, suiteservices.PostgresFixturePolicyPackageReset)

	postgresPIDs := uniqueEventPIDs(events, suiteservices.EventPostgresDBCreated)
	minioPIDs := uniqueEventPIDs(events, suiteservices.EventS3BucketCreated)
	if len(postgresPIDs) < 2 {
		t.Fatalf("expected postgres database creation from multiple go test package processes, got %v", postgresPIDs)
	}
	if len(minioPIDs) < 2 {
		t.Fatalf("expected bucket creation from multiple go test package processes, got %v", minioPIDs)
	}
	if hasDuplicateNames(events, suiteservices.EventPostgresDBCreated) {
		t.Fatal("expected distinct database names across package binaries")
	}
	if hasDuplicateNames(events, suiteservices.EventS3BucketCreated) {
		t.Fatal("expected distinct bucket names across package binaries")
	}
}

func assertNoHotPathPostgresDrops(t testing.TB, scope suiteservices.ServiceScope) {
	t.Helper()

	for _, activity := range scope.Fixture.ByStrategy {
		if activity.Service != suiteservices.ServicePostgres || activity.Operation != "database-drop" {
			continue
		}
		t.Fatalf("unexpected hot-path postgres database drop activity: %#v", activity)
	}
}

func assertNoPostgresDatabaseResets(t testing.TB, scope suiteservices.ServiceScope) {
	t.Helper()

	for _, activity := range scope.Fixture.ByStrategy {
		if activity.Service == suiteservices.ServicePostgres && activity.Operation == "database-reset" {
			t.Fatalf("unexpected postgres database reset activity: %#v", activity)
		}
	}
}

func assertPostgresDatabaseResets(t testing.TB, scope suiteservices.ServiceScope) {
	t.Helper()

	for _, activity := range scope.Fixture.ByStrategy {
		if activity.Service == suiteservices.ServicePostgres && activity.Operation == "database-reset" {
			return
		}
	}
	t.Fatalf("expected postgres database reset activity, got %#v", scope.Fixture.ByStrategy)
}

func assertPostgresPackageResetsLimitedToHarnessPolicy(t testing.TB, scope suiteservices.ServiceScope) {
	t.Helper()

	for _, activity := range scope.Fixture.ByPackage {
		if activity.Service != suiteservices.ServicePostgres || activity.Operation != "database-reset" {
			continue
		}
		if activity.FixturePolicy == suiteservices.PostgresFixturePolicyPackageReset {
			continue
		}
		t.Fatalf("unexpected non-harness postgres package reset activity: %#v", activity)
	}
}

func assertPostgresFixturePolicy(t testing.TB, scope suiteservices.ServiceScope, policy string) {
	t.Helper()

	for _, activity := range scope.Fixture.ByStrategy {
		if activity.Service == suiteservices.ServicePostgres && activity.FixturePolicy == policy {
			return
		}
	}
	t.Fatalf("expected postgres fixture policy %q in %#v", policy, scope.Fixture.ByStrategy)
}

func TestMakeMigrationDriftDoesNotEmitSuiteServiceArtifacts(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	requireDocker(t)

	resultsDir := t.TempDir()
	runID := "migration-drift-direct"
	runMakeTarget(t, resultsDir, runID, nil, "migration-drift")

	sharedRoot := filepath.Join(resultsDir, runID, "_shared", "test-services")
	matches, err := filepath.Glob(filepath.Join(sharedRoot, "*", "service-scope.json"))
	if err != nil {
		t.Fatalf("list suite service summaries: %v", err)
	}
	if len(matches) != 0 {
		t.Fatalf("migration-drift must not use suite template services, found %v", matches)
	}
}

func runMakeAndLoadSingleScope(t testing.TB, runID string, extraEnv map[string]string, target string) suiteservices.ServiceScope {
	t.Helper()

	scope, _ := runMakeAndLoadScopeWithEvents(t, runID, extraEnv, target)
	return scope
}

func runMakeAndLoadScopeWithEvents(t testing.TB, runID string, extraEnv map[string]string, target string) (suiteservices.ServiceScope, []suiteservices.Event) {
	t.Helper()

	resultsDir := t.TempDir()
	runMakeTarget(t, resultsDir, runID, extraEnv, target)

	summaries, err := filepath.Glob(filepath.Join(resultsDir, runID, "_shared", "test-services", "*", "service-scope.json"))
	if err != nil {
		t.Fatalf("list suite service summaries: %v", err)
	}
	if len(summaries) != 1 {
		t.Fatalf("expected exactly one suite service summary, found %v", summaries)
	}

	raw, err := os.ReadFile(summaries[0])
	if err != nil {
		t.Fatalf("read suite service summary: %v", err)
	}

	var scope suiteservices.ServiceScope
	if err := json.Unmarshal(raw, &scope); err != nil {
		t.Fatalf("decode suite service summary: %v", err)
	}

	events := loadEvents(t, filepath.Dir(summaries[0]))
	return scope, events
}

func runMakeTarget(t testing.TB, resultsDir string, runID string, extraEnv map[string]string, target string) {
	t.Helper()

	repoRoot, err := suiteservices.FindRepoRoot()
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	command := exec.CommandContext(ctx, "make", "--no-print-directory", target)
	command.Dir = repoRoot
	command.Env = append(os.Environ(),
		"CARTULARY_TEST_RESULTS_DIR="+resultsDir,
		"CARTULARY_TEST_RUN_ID="+runID,
		"CARTULARY_OUTPUT_MODE=quiet",
	)
	for key, value := range extraEnv {
		command.Env = append(command.Env, key+"="+value)
	}
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("run make %s: %v\n%s", target, err, output)
	}
}

func requireDocker(t testing.TB) {
	t.Helper()

	command := exec.Command("docker", "info")
	if output, err := command.CombinedOutput(); err != nil {
		t.Skipf("docker unavailable for integration test: %v\n%s", err, output)
	}
}

func loadEvents(t testing.TB, suiteDir string) []suiteservices.Event {
	t.Helper()

	eventFiles, err := filepath.Glob(filepath.Join(suiteDir, "events", "*.json"))
	if err != nil {
		t.Fatalf("list suite service events: %v", err)
	}
	slices.Sort(eventFiles)

	events := make([]suiteservices.Event, 0, len(eventFiles))
	for _, eventPath := range eventFiles {
		raw, err := os.ReadFile(eventPath)
		if err != nil {
			t.Fatalf("read suite service event %s: %v", eventPath, err)
		}
		var event suiteservices.Event
		if err := json.Unmarshal(raw, &event); err != nil {
			t.Fatalf("decode suite service event %s: %v", eventPath, err)
		}
		events = append(events, event)
	}
	return events
}

func assertSuiteServicesStarted(t testing.TB, scope suiteservices.ServiceScope) {
	t.Helper()

	if !scope.Postgres.Started {
		t.Fatalf("expected postgres suite service to start, got %#v", scope.Postgres)
	}
	if !scope.MinIO.Started {
		t.Fatalf("expected minio suite service to start, got %#v", scope.MinIO)
	}
}

func assertPostgresPreparationStrategy(t testing.TB, scope suiteservices.ServiceScope, strategy string) {
	t.Helper()

	for _, preparation := range scope.Postgres.DatabasePreparations {
		if preparation.Strategy == strategy {
			return
		}
	}
	t.Fatalf("expected postgres preparation strategy %q in %#v", strategy, scope.Postgres.DatabasePreparations)
}

func uniqueEventPIDs(events []suiteservices.Event, eventType string) []int {
	seen := make(map[int]struct{})
	for _, event := range events {
		if event.Type != eventType {
			continue
		}
		seen[event.PID] = struct{}{}
	}

	pids := make([]int, 0, len(seen))
	for pid := range seen {
		pids = append(pids, pid)
	}
	slices.Sort(pids)
	return pids
}

func hasDuplicateNames(events []suiteservices.Event, eventType string) bool {
	seen := make(map[string]struct{})
	for _, event := range events {
		if event.Type != eventType || event.Name == "" {
			continue
		}
		if _, exists := seen[event.Name]; exists {
			return true
		}
		seen[event.Name] = struct{}{}
	}
	return false
}
