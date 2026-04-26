package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices"
)

func TestMakeBackendStoreUsesSingleOwnedSuitePair(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	requireDocker(t)

	scope, events := runMakeAndLoadScopeWithEvents(t, "backend-store-direct", nil, "backend-store")
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
	assertNoPostgresDatabaseResets(t, scope)
	assertPostgresFixturePolicy(t, scope, suiteservices.PostgresFixturePolicyTemplateClone)
	assertPostgresFixtureBudgetFromPlan(t, events, "backend-store")
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
	assertPostgresFixtureBudgetFromPlan(t, events, "backend-integration")
	assertPostgresFixtureBudgetFromPlan(t, events, "backend-integration-support")

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

type postgresFixtureBudget struct {
	PackageResetGroups    int
	PackageResetEvents    int
	PackageResetDuration  int64
	MigrationScratchTests int
	TemplateCloneTests    map[string]struct{}
	PackageResetTests     map[string]postgresResetBudget
}

type postgresFixtureStats struct {
	PackageResetCreates     int
	PackageResetEvents      int
	PackageResetDuration    int64
	PackageResetByPackage   map[string]int64
	MigrationScratchCreates int
	TemplateCloneCreates    int
	UnplannedTemplateClones []string
}

type postgresResetBudget struct {
	MaxPackageResets   int
	MaxResetDurationMS int64
}

func assertPostgresFixtureBudgetFromPlan(t testing.TB, events []suiteservices.Event, target string) {
	t.Helper()

	budget := plannedPostgresFixtureBudget(t, target)
	stats := postgresFixtureStatsFromEvents(events, target, budget)
	if stats.PackageResetCreates > budget.PackageResetGroups {
		t.Fatalf("target %s exceeded package-reset postgres fixture budget: got %d creates, budget %d", target, stats.PackageResetCreates, budget.PackageResetGroups)
	}
	if stats.MigrationScratchCreates > budget.MigrationScratchTests*2 {
		t.Fatalf("target %s exceeded migration-scratch postgres fixture budget: got %d creates, budget %d tests", target, stats.MigrationScratchCreates, budget.MigrationScratchTests)
	}
	if len(stats.UnplannedTemplateClones) > 0 {
		t.Fatalf("target %s used unplanned per-test postgres template clones: %s", target, strings.Join(stats.UnplannedTemplateClones, "; "))
	}
	if stats.TemplateCloneCreates > len(budget.TemplateCloneTests)*2 {
		t.Fatalf("target %s exceeded template-clone postgres fixture budget: got %d creates, budget %d tests", target, stats.TemplateCloneCreates, len(budget.TemplateCloneTests))
	}
	if stats.PackageResetEvents > budget.PackageResetEvents {
		t.Fatalf("target %s exceeded package-reset event budget: got %d resets, budget %d", target, stats.PackageResetEvents, budget.PackageResetEvents)
	}
	if stats.PackageResetDuration > budget.PackageResetDuration {
		t.Fatalf("target %s exceeded package-reset duration budget: got %dms, budget %dms", target, stats.PackageResetDuration, budget.PackageResetDuration)
	}
	for pkg, duration := range stats.PackageResetByPackage {
		if duration > 10000 {
			t.Fatalf("target %s exceeded package reset package duration budget: package %s got %dms, budget 10000ms", target, pkg, duration)
		}
	}
}

func plannedPostgresFixtureBudget(t testing.TB, target string) postgresFixtureBudget {
	t.Helper()

	repoRoot, err := suiteservices.FindRepoRoot()
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}
	command := exec.Command("node", "scripts/print-go-shard-plan.mjs", "--json", "--target", target)
	command.Dir = repoRoot
	raw, err := command.Output()
	if err != nil {
		t.Fatalf("load shard plan for %s: %v", target, err)
	}

	var plan struct {
		Shards []struct {
			Items []struct {
				Kind                  string   `json:"kind"`
				Symbol                string   `json:"symbol"`
				ImportPath            string   `json:"import_path"`
				Packages              []string `json:"packages"`
				PostgresFixturePolicy string   `json:"postgres_fixture_policy"`
				PostgresFixtureBudget struct {
					MaxPackageResets   int   `json:"max_package_resets"`
					MaxResetDurationMS int64 `json:"max_reset_duration_ms"`
				} `json:"postgres_fixture_budget"`
			} `json:"items"`
		} `json:"shards"`
	}
	if err := json.Unmarshal(raw, &plan); err != nil {
		t.Fatalf("decode shard plan for %s: %v", target, err)
	}

	budget := postgresFixtureBudget{
		TemplateCloneTests: make(map[string]struct{}),
		PackageResetTests:  make(map[string]postgresResetBudget),
	}
	for _, shard := range plan.Shards {
		packageResetPackages := make(map[string]struct{})
		for _, item := range shard.Items {
			switch item.PostgresFixturePolicy {
			case suiteservices.PostgresFixturePolicyPackageReset:
				resetBudget := postgresResetBudget{
					MaxPackageResets:   item.PostgresFixtureBudget.MaxPackageResets,
					MaxResetDurationMS: item.PostgresFixtureBudget.MaxResetDurationMS,
				}
				if item.Symbol != "" && (resetBudget.MaxPackageResets > 0 || resetBudget.MaxResetDurationMS > 0) {
					budget.PackageResetTests[item.Symbol] = resetBudget
					budget.PackageResetEvents += resetBudget.MaxPackageResets
					budget.PackageResetDuration += resetBudget.MaxResetDurationMS
				}
				if item.Kind == "raw" {
					for _, pkg := range item.Packages {
						packageResetPackages[normalizePlanPackage(pkg)] = struct{}{}
					}
					continue
				}
				packageResetPackages[normalizePlanPackage(item.ImportPath)] = struct{}{}
			case suiteservices.PostgresFixturePolicyMigrationScratch:
				budget.MigrationScratchTests++
			case suiteservices.PostgresFixturePolicyTemplateClone:
				if item.Symbol != "" {
					budget.TemplateCloneTests[item.Symbol] = struct{}{}
				}
			}
		}
		budget.PackageResetGroups += len(packageResetPackages)
	}
	return budget
}

func postgresFixtureStatsFromEvents(events []suiteservices.Event, target string, budget postgresFixtureBudget) postgresFixtureStats {
	stats := postgresFixtureStats{PackageResetByPackage: make(map[string]int64)}
	for _, event := range events {
		if event.Type == suiteservices.EventPostgresDBReset &&
			stringEventDetail(event, "target") == target &&
			stringEventDetail(event, "reuse_scope") == suiteservices.FixtureReusePackage &&
			stringEventDetail(event, "fixture_policy") == suiteservices.PostgresFixturePolicyPackageReset {
			testName := topLevelTestName(stringEventDetail(event, "test_name"))
			if _, ok := budget.PackageResetTests[testName]; ok {
				duration := int64EventDetail(event, "duration_ms")
				stats.PackageResetEvents++
				stats.PackageResetDuration += duration
				stats.PackageResetByPackage[stringEventDetail(event, "caller_package")] += duration
			}
		}
		if event.Type != suiteservices.EventPostgresDBCreated || event.Kind != suiteservices.PostgresPreparationTemplateClone {
			if event.Type == suiteservices.EventPostgresDBCreated && event.Kind == "scratch" && stringEventDetail(event, "target") == target && stringEventDetail(event, "reuse_scope") == suiteservices.FixtureReuseMigrationScratch {
				stats.MigrationScratchCreates++
			}
			continue
		}
		if stringEventDetail(event, "target") != target {
			continue
		}
		reuseScope := stringEventDetail(event, "reuse_scope")
		policy := stringEventDetail(event, "fixture_policy")
		switch {
		case reuseScope == suiteservices.FixtureReusePackage && policy == suiteservices.PostgresFixturePolicyPackageReset:
			stats.PackageResetCreates++
		case reuseScope == suiteservices.FixtureReusePerTest:
			if isPgtestHarnessTemplateClone(event, target) {
				continue
			}
			stats.TemplateCloneCreates++
			testName := topLevelTestName(stringEventDetail(event, "test_name"))
			if _, ok := budget.TemplateCloneTests[testName]; !ok {
				stats.UnplannedTemplateClones = append(stats.UnplannedTemplateClones, fmt.Sprintf("%s %s", testName, stringEventDetail(event, "caller_file")))
			}
		}
	}
	return stats
}

func int64EventDetail(event suiteservices.Event, key string) int64 {
	if event.Details == nil {
		return 0
	}
	value, ok := event.Details[key]
	if !ok || value == nil {
		return 0
	}
	switch typed := value.(type) {
	case float64:
		return int64(typed)
	case int64:
		return typed
	case int:
		return int64(typed)
	default:
		return 0
	}
}

func isPgtestHarnessTemplateClone(event suiteservices.Event, target string) bool {
	if target != "backend-integration" {
		return false
	}
	callerPackage := stringEventDetail(event, "caller_package")
	callerFile := stringEventDetail(event, "caller_file")
	testName := topLevelTestName(stringEventDetail(event, "test_name"))
	if callerPackage == "internal/testutil/pgtest" || callerFile == "internal/testutil/pgtest/pgtest_test.go" {
		return true
	}
	if strings.HasPrefix(testName, "TestPrepareDatabase") || strings.HasPrefix(testName, "TestMigrationDatabase") {
		return true
	}
	return callerPackage == "" && callerFile == "" && testName == ""
}

func normalizePlanPackage(value string) string {
	value = strings.TrimSpace(value)
	value = strings.TrimPrefix(value, "./")
	value = strings.TrimPrefix(value, "github.com/JochiRaider/cartulary/")
	return value
}

func stringEventDetail(event suiteservices.Event, key string) string {
	if event.Details == nil {
		return ""
	}
	value, ok := event.Details[key]
	if !ok || value == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprint(value))
}

func topLevelTestName(testName string) string {
	if before, _, ok := strings.Cut(testName, "/"); ok {
		return before
	}
	return testName
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
