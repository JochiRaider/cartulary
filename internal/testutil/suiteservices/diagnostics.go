package suiteservices

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync/atomic"
	"time"
)

const (
	ServicePostgres = "postgres"
	ServiceMinIO    = "minio"
)

const (
	EventWrapperOwnedStart    = "wrapper-owned-start"
	EventWrapperPassThrough   = "wrapper-pass-through"
	EventServiceStarted       = "service-started"
	EventFailureRecorded      = "failure-recorded"
	EventCleanupCompleted     = "cleanup-completed"
	EventPostgresAttach       = "postgres-attach"
	EventPostgresDBCreated    = "postgres-db-created"
	EventPostgresDBDropped    = "postgres-db-dropped"
	EventPostgresDBMigrated   = "postgres-db-migrated"
	EventPostgresDBRetained   = "postgres-db-retained"
	EventPostgresDBReset      = "postgres-db-reset"
	EventPostgresTransaction  = "postgres-transaction"
	EventPostgresTemplateUse  = "postgres-template-clone"
	EventS3Attach             = "s3-attach"
	EventS3BucketCreated      = "s3-bucket-created"
	EventS3BucketCleaned      = "s3-bucket-cleaned"
	EventS3PrefixCleaned      = "s3-prefix-cleaned"
	EventTimingSpan           = "timing-span"
	EventWebE2EFixtureRetired = "web-e2e-fixture-retired"
	EventWebE2EFixtureCleaned = "web-e2e-fixture-cleaned"
)

const (
	PostgresPreparationTemplate       = "template"
	PostgresPreparationTemplateClone  = "template-clone"
	PostgresPreparationFreshMigration = "fresh-migration"
)

const (
	PostgresFixturePolicyTemplateClone    = "template_clone"
	PostgresFixturePolicyPackageReset     = "package_reset"
	PostgresFixturePolicyMigrationScratch = "migration_scratch"
)

const (
	FixtureReusePerTest          = "per-test"
	FixtureReusePackage          = "package-reused"
	FixtureReuseTransaction      = "transaction"
	FixtureReusePrefix           = "prefix-reused"
	FixtureReuseSuiteTemplate    = "suite-template"
	FixtureReuseMigrationScratch = "migration-scratch"
)

type Event struct {
	Type      string         `json:"type"`
	Timestamp string         `json:"timestamp"`
	PID       int            `json:"pid"`
	Service   string         `json:"service,omitempty"`
	Name      string         `json:"name,omitempty"`
	Kind      string         `json:"kind,omitempty"`
	Status    string         `json:"status,omitempty"`
	Details   map[string]any `json:"details,omitempty"`
}

type ServiceScope struct {
	SuiteID      string               `json:"suite_id"`
	RunID        string               `json:"run_id"`
	ArtifactDir  string               `json:"artifact_dir"`
	Wrapper      WrapperSummary       `json:"wrapper"`
	Failure      *FailureSummary      `json:"failure,omitempty"`
	Cleanup      CleanupSummary       `json:"cleanup"`
	Postgres     PostgresSummary      `json:"postgres"`
	MinIO        MinIOSummary         `json:"minio"`
	BrowserE2E   BrowserE2ESummary    `json:"browser_e2e"`
	Fixture      FixtureSummary       `json:"fixture"`
	StartedNames StartedServiceRecord `json:"started_services"`
}

type WrapperSummary struct {
	OwnedCount       int `json:"owned_count"`
	PassThroughCount int `json:"pass_through_count"`
}

type FailureSummary struct {
	Service               string `json:"service,omitempty"`
	Stage                 string `json:"stage,omitempty"`
	Operation             string `json:"operation,omitempty"`
	Message               string `json:"message,omitempty"`
	AttemptsStarted       int    `json:"attempts_started"`
	MaxAttempts           int    `json:"max_attempts"`
	Retryable             bool   `json:"retryable"`
	RetryBlockedByContext bool   `json:"retry_blocked_by_context"`
	DockerEndpoint        string `json:"docker_endpoint,omitempty"`
}

type CleanupSummary struct {
	Status          string `json:"status,omitempty"`
	CompletedAt     string `json:"completed_at,omitempty"`
	ChildExitStatus *int   `json:"child_exit_status,omitempty"`
}

type PostgresSummary struct {
	Started               bool                          `json:"started"`
	StartedAt             string                        `json:"started_at,omitempty"`
	Host                  string                        `json:"host,omitempty"`
	Port                  string                        `json:"port,omitempty"`
	User                  string                        `json:"user,omitempty"`
	TemplateDatabase      string                        `json:"template_database,omitempty"`
	AttachedHarnessCount  int                           `json:"attached_harness_count"`
	CreatedDatabaseCount  int                           `json:"created_database_count"`
	MigratedDatabaseCount int                           `json:"migrated_database_count"`
	TemplateCloneCount    int                           `json:"template_clone_count"`
	CreatedDatabases      []string                      `json:"created_databases,omitempty"`
	DatabasePreparations  []PostgresDatabasePreparation `json:"database_preparations,omitempty"`
}

type PostgresDatabasePreparation struct {
	Name             string `json:"name"`
	Strategy         string `json:"strategy"`
	TemplateDatabase string `json:"template_database,omitempty"`
	Target           string `json:"target,omitempty"`
	PID              int    `json:"pid"`
	Timestamp        string `json:"timestamp"`
}

type MinIOSummary struct {
	Started              bool     `json:"started"`
	StartedAt            string   `json:"started_at,omitempty"`
	Endpoint             string   `json:"endpoint,omitempty"`
	Secure               bool     `json:"secure"`
	AttachedHarnessCount int      `json:"attached_harness_count"`
	BucketCreateCount    int      `json:"bucket_create_count"`
	BucketCleanupCount   int      `json:"bucket_cleanup_count"`
	CreatedBuckets       []string `json:"created_buckets,omitempty"`
	CleanedBuckets       []string `json:"cleaned_buckets,omitempty"`
}

type BrowserE2ESummary struct {
	RetiredFixtureCount int                    `json:"retired_fixture_count"`
	CleanedFixtureCount int                    `json:"cleaned_fixture_count"`
	RetiredFixtures     []BrowserE2EFixtureRef `json:"retired_fixtures,omitempty"`
	CleanedFixtures     []BrowserE2EFixtureRef `json:"cleaned_fixtures,omitempty"`
}

type BrowserE2EFixtureRef struct {
	DatabaseName string `json:"database_name,omitempty"`
	Bucket       string `json:"bucket,omitempty"`
	Target       string `json:"target,omitempty"`
	Timestamp    string `json:"timestamp,omitempty"`
	PID          int    `json:"pid,omitempty"`
}

type FixtureSummary struct {
	TotalCount      int               `json:"total_count"`
	TotalDurationMS int64             `json:"total_duration_ms"`
	ByPackage       []FixtureActivity `json:"by_package,omitempty"`
	ByTest          []FixtureActivity `json:"by_test,omitempty"`
	ByStrategy      []FixtureActivity `json:"by_strategy,omitempty"`
	Slowest         []FixtureActivity `json:"slowest,omitempty"`
}

type FixtureActivity struct {
	Service         string `json:"service,omitempty"`
	Target          string `json:"target,omitempty"`
	Operation       string `json:"operation,omitempty"`
	Strategy        string `json:"strategy,omitempty"`
	FixturePolicy   string `json:"fixture_policy,omitempty"`
	ReuseScope      string `json:"reuse_scope,omitempty"`
	CallerPackage   string `json:"caller_package,omitempty"`
	CallerFile      string `json:"caller_file,omitempty"`
	TestName        string `json:"test_name,omitempty"`
	Count           int    `json:"count"`
	TotalDurationMS int64  `json:"total_duration_ms"`
}

type StartedServiceRecord struct {
	Names []string `json:"names,omitempty"`
}

var eventSequence uint64

func RecordEvent(env map[string]string, event Event) error {
	suiteDir, ok, err := ResolveSuiteArtifactDir(env)
	if err != nil || !ok {
		return err
	}

	if strings.TrimSpace(event.Type) == "" {
		return fmt.Errorf("record suite-service event: type is required")
	}
	if strings.TrimSpace(event.Timestamp) == "" {
		event.Timestamp = time.Now().UTC().Format(time.RFC3339Nano)
	}
	if event.PID == 0 {
		event.PID = os.Getpid()
	}
	event = sanitizeEvent(event)

	eventsDir := filepath.Join(suiteDir, "events")
	if err := os.MkdirAll(eventsDir, 0o755); err != nil {
		return fmt.Errorf("create suite-service events dir: %w", err)
	}

	sequence := atomic.AddUint64(&eventSequence, 1)
	fileName := fmt.Sprintf("%s-%d-%06d-%s.json", sanitizeFileComponent(event.Timestamp), event.PID, sequence, sanitizeFileComponent(event.Type))
	eventPath := filepath.Join(eventsDir, fileName)
	payload, err := json.MarshalIndent(event, "", "  ")
	if err != nil {
		return fmt.Errorf("encode suite-service event: %w", err)
	}
	if err := os.WriteFile(eventPath, append(payload, '\n'), 0o644); err != nil {
		return fmt.Errorf("write suite-service event: %w", err)
	}

	return nil
}

func RefreshSummary(env map[string]string) error {
	scope, ok, err := Summarize(env)
	if err != nil || !ok {
		return err
	}

	suiteDir, _, err := ResolveSuiteArtifactDir(env)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(suiteDir, 0o755); err != nil {
		return fmt.Errorf("create suite-service artifact dir: %w", err)
	}

	payload, err := json.MarshalIndent(scope, "", "  ")
	if err != nil {
		return fmt.Errorf("encode suite-service summary: %w", err)
	}
	if err := os.WriteFile(filepath.Join(suiteDir, "service-scope.json"), append(payload, '\n'), 0o644); err != nil {
		return fmt.Errorf("write suite-service summary: %w", err)
	}
	return nil
}

func Summarize(env map[string]string) (ServiceScope, bool, error) {
	suiteDir, ok, err := ResolveSuiteArtifactDir(env)
	if err != nil || !ok {
		return ServiceScope{}, ok, err
	}

	scope := ServiceScope{
		SuiteID:     SuiteID(env),
		RunID:       ResolveRunID(env),
		ArtifactDir: suiteDir,
	}

	eventFiles, err := filepath.Glob(filepath.Join(suiteDir, "events", "*.json"))
	if err != nil {
		return ServiceScope{}, true, fmt.Errorf("list suite-service events: %w", err)
	}
	sort.Strings(eventFiles)

	createdDatabases := make(map[string]struct{})
	createdBuckets := make(map[string]struct{})
	cleanedBuckets := make(map[string]struct{})
	startedServices := make(map[string]struct{})
	databasePreparations := make(map[string]PostgresDatabasePreparation)
	retiredWebE2EFixtures := make(map[string]BrowserE2EFixtureRef)
	cleanedWebE2EFixtures := make(map[string]BrowserE2EFixtureRef)
	packageFixtures := make(map[string]FixtureActivity)
	testFixtures := make(map[string]FixtureActivity)
	strategyFixtures := make(map[string]FixtureActivity)
	slowestFixtures := make([]FixtureActivity, 0)

	for _, eventPath := range eventFiles {
		raw, err := os.ReadFile(eventPath)
		if err != nil {
			return ServiceScope{}, true, fmt.Errorf("read suite-service event %s: %w", eventPath, err)
		}

		var event Event
		if err := json.Unmarshal(raw, &event); err != nil {
			return ServiceScope{}, true, fmt.Errorf("decode suite-service event %s: %w", eventPath, err)
		}

		switch event.Type {
		case EventWrapperOwnedStart:
			scope.Wrapper.OwnedCount++
		case EventWrapperPassThrough:
			scope.Wrapper.PassThroughCount++
		case EventFailureRecorded:
			if scope.Failure == nil {
				scope.Failure = &FailureSummary{
					Service:               event.Service,
					Stage:                 stringDetail(event.Details, "stage"),
					Operation:             stringDetail(event.Details, "operation"),
					Message:               stringDetail(event.Details, "message"),
					AttemptsStarted:       intValue(event.Details, "attempts_started"),
					MaxAttempts:           intValue(event.Details, "max_attempts"),
					Retryable:             boolDetail(event.Details, "retryable"),
					RetryBlockedByContext: boolDetail(event.Details, "retry_blocked_by_context"),
					DockerEndpoint:        stringDetail(event.Details, "docker_endpoint"),
				}
			}
		case EventServiceStarted:
			startedServices[event.Service] = struct{}{}
			switch event.Service {
			case ServicePostgres:
				scope.Postgres.Started = true
				scope.Postgres.StartedAt = earliestTimestamp(scope.Postgres.StartedAt, event.Timestamp)
				scope.Postgres.Host = stringDetail(event.Details, "host")
				scope.Postgres.Port = stringDetail(event.Details, "port")
				scope.Postgres.User = stringDetail(event.Details, "user")
				if templateDB := stringDetail(event.Details, "template_database"); templateDB != "" {
					scope.Postgres.TemplateDatabase = templateDB
				}
			case ServiceMinIO:
				scope.MinIO.Started = true
				scope.MinIO.StartedAt = earliestTimestamp(scope.MinIO.StartedAt, event.Timestamp)
				scope.MinIO.Endpoint = stringDetail(event.Details, "endpoint")
				scope.MinIO.Secure = boolDetail(event.Details, "secure")
			}
		case EventCleanupCompleted:
			scope.Cleanup.Status = event.Status
			scope.Cleanup.CompletedAt = latestTimestamp(scope.Cleanup.CompletedAt, event.Timestamp)
			if childExit, ok := intDetail(event.Details, "child_exit_status"); ok {
				value := childExit
				scope.Cleanup.ChildExitStatus = &value
			}
		case EventPostgresAttach:
			scope.Postgres.AttachedHarnessCount++
		case EventPostgresDBCreated:
			scope.Postgres.CreatedDatabaseCount++
			if event.Name != "" {
				createdDatabases[event.Name] = struct{}{}
			}
			if event.Kind == "template" && event.Name != "" {
				scope.Postgres.TemplateDatabase = event.Name
			}
			upsertPostgresPreparation(databasePreparations, event, strategyForPostgresCreateEvent(event))
			recordFixtureActivity(&scope.Fixture, packageFixtures, testFixtures, strategyFixtures, &slowestFixtures, event)
		case EventPostgresDBDropped:
			recordFixtureActivity(&scope.Fixture, packageFixtures, testFixtures, strategyFixtures, &slowestFixtures, event)
		case EventPostgresDBRetained:
			recordFixtureActivity(&scope.Fixture, packageFixtures, testFixtures, strategyFixtures, &slowestFixtures, event)
		case EventPostgresDBMigrated:
			scope.Postgres.MigratedDatabaseCount++
			upsertPostgresPreparation(databasePreparations, event, strategyForPostgresMigratedEvent(event))
			recordFixtureActivity(&scope.Fixture, packageFixtures, testFixtures, strategyFixtures, &slowestFixtures, event)
		case EventPostgresDBReset, EventPostgresTransaction:
			recordFixtureActivity(&scope.Fixture, packageFixtures, testFixtures, strategyFixtures, &slowestFixtures, event)
		case EventPostgresTemplateUse:
			scope.Postgres.TemplateCloneCount++
			upsertPostgresPreparation(databasePreparations, event, PostgresPreparationTemplateClone)
		case EventS3Attach:
			scope.MinIO.AttachedHarnessCount++
		case EventS3BucketCreated:
			scope.MinIO.BucketCreateCount++
			if event.Name != "" {
				createdBuckets[event.Name] = struct{}{}
			}
			recordFixtureActivity(&scope.Fixture, packageFixtures, testFixtures, strategyFixtures, &slowestFixtures, event)
		case EventS3BucketCleaned:
			scope.MinIO.BucketCleanupCount++
			if event.Name != "" {
				cleanedBuckets[event.Name] = struct{}{}
			}
			recordFixtureActivity(&scope.Fixture, packageFixtures, testFixtures, strategyFixtures, &slowestFixtures, event)
		case EventS3PrefixCleaned:
			recordFixtureActivity(&scope.Fixture, packageFixtures, testFixtures, strategyFixtures, &slowestFixtures, event)
		case EventWebE2EFixtureRetired:
			scope.BrowserE2E.RetiredFixtureCount++
			upsertWebE2EFixture(retiredWebE2EFixtures, event)
		case EventWebE2EFixtureCleaned:
			scope.BrowserE2E.CleanedFixtureCount++
			upsertWebE2EFixture(cleanedWebE2EFixtures, event)
		}
	}

	scope.Postgres.CreatedDatabases = sortedKeys(createdDatabases)
	scope.Postgres.DatabasePreparations = sortedPostgresPreparations(databasePreparations)
	scope.MinIO.CreatedBuckets = sortedKeys(createdBuckets)
	scope.MinIO.CleanedBuckets = sortedKeys(cleanedBuckets)
	scope.BrowserE2E.RetiredFixtures = sortedWebE2EFixtures(retiredWebE2EFixtures)
	scope.BrowserE2E.CleanedFixtures = sortedWebE2EFixtures(cleanedWebE2EFixtures)
	scope.StartedNames.Names = sortedKeys(startedServices)
	scope.Fixture.ByPackage = sortedFixtureActivities(packageFixtures)
	scope.Fixture.ByTest = sortedFixtureActivities(testFixtures)
	scope.Fixture.ByStrategy = sortedFixtureActivities(strategyFixtures)
	scope.Fixture.Slowest = topFixtureActivities(slowestFixtures, 10)

	return scope, true, nil
}

func recordFixtureActivity(summary *FixtureSummary, byPackage map[string]FixtureActivity, byTest map[string]FixtureActivity, byStrategy map[string]FixtureActivity, slowest *[]FixtureActivity, event Event) {
	activity := fixtureActivityFromEvent(event)
	if activity.Service == "" {
		return
	}

	summary.TotalCount++
	summary.TotalDurationMS += activity.TotalDurationMS
	upsertFixtureActivity(byPackage, fixtureKey(activity.Service, activity.Target, activity.CallerPackage, activity.FixturePolicy, activity.Operation, activity.ReuseScope), activity, func(existing *FixtureActivity) {
		existing.Service = activity.Service
		existing.Target = activity.Target
		existing.CallerPackage = activity.CallerPackage
		existing.FixturePolicy = activity.FixturePolicy
		existing.Operation = activity.Operation
		existing.ReuseScope = activity.ReuseScope
	})
	upsertFixtureActivity(byTest, fixtureKey(activity.Service, activity.Target, activity.TestName, activity.FixturePolicy, activity.Operation, activity.ReuseScope), activity, func(existing *FixtureActivity) {
		existing.Service = activity.Service
		existing.Target = activity.Target
		existing.TestName = activity.TestName
		existing.FixturePolicy = activity.FixturePolicy
		existing.Operation = activity.Operation
		existing.ReuseScope = activity.ReuseScope
	})
	upsertFixtureActivity(byStrategy, fixtureKey(activity.Service, activity.Target, activity.Strategy, activity.FixturePolicy, activity.Operation, activity.ReuseScope), activity, func(existing *FixtureActivity) {
		existing.Service = activity.Service
		existing.Target = activity.Target
		existing.Strategy = activity.Strategy
		existing.FixturePolicy = activity.FixturePolicy
		existing.Operation = activity.Operation
		existing.ReuseScope = activity.ReuseScope
	})
	*slowest = append(*slowest, activity)
}

func fixtureActivityFromEvent(event Event) FixtureActivity {
	service := fixtureServiceForEvent(event.Type)
	if service == "" {
		return FixtureActivity{}
	}
	activity := FixtureActivity{
		Service:         service,
		Target:          stringDetail(event.Details, "target"),
		Operation:       fixtureOperationForEvent(event.Type),
		Strategy:        firstNonEmpty(stringDetail(event.Details, "strategy"), stringDetail(event.Details, "preparation_strategy")),
		FixturePolicy:   stringDetail(event.Details, "fixture_policy"),
		ReuseScope:      firstNonEmpty(stringDetail(event.Details, "reuse_scope"), FixtureReusePerTest),
		CallerPackage:   stringDetail(event.Details, "caller_package"),
		CallerFile:      stringDetail(event.Details, "caller_file"),
		TestName:        stringDetail(event.Details, "test_name"),
		Count:           1,
		TotalDurationMS: int64Value(event.Details, "duration_ms"),
	}
	return activity
}

func fixtureServiceForEvent(eventType string) string {
	switch eventType {
	case EventPostgresDBCreated, EventPostgresDBDropped, EventPostgresDBMigrated, EventPostgresDBRetained, EventPostgresDBReset, EventPostgresTransaction:
		return ServicePostgres
	case EventS3BucketCreated, EventS3BucketCleaned, EventS3PrefixCleaned:
		return ServiceMinIO
	default:
		return ""
	}
}

func fixtureOperationForEvent(eventType string) string {
	switch eventType {
	case EventPostgresDBCreated:
		return "database-create"
	case EventPostgresDBDropped:
		return "database-drop"
	case EventPostgresDBRetained:
		return "database-retain"
	case EventPostgresDBMigrated:
		return "database-migrate"
	case EventPostgresDBReset:
		return "database-reset"
	case EventPostgresTransaction:
		return "transaction"
	case EventS3BucketCreated:
		return "bucket-create"
	case EventS3BucketCleaned:
		return "bucket-clean"
	case EventS3PrefixCleaned:
		return "prefix-clean"
	default:
		return eventType
	}
}

func upsertFixtureActivity(values map[string]FixtureActivity, key string, activity FixtureActivity, initialize func(*FixtureActivity)) {
	if key == "" {
		return
	}
	existing := values[key]
	if existing.Count == 0 {
		initialize(&existing)
	}
	existing.Count += activity.Count
	existing.TotalDurationMS += activity.TotalDurationMS
	values[key] = existing
}

func sortedFixtureActivities(values map[string]FixtureActivity) []FixtureActivity {
	if len(values) == 0 {
		return nil
	}
	activities := make([]FixtureActivity, 0, len(values))
	for _, activity := range values {
		activities = append(activities, activity)
	}
	sort.Slice(activities, func(i, j int) bool {
		if activities[i].TotalDurationMS != activities[j].TotalDurationMS {
			return activities[i].TotalDurationMS > activities[j].TotalDurationMS
		}
		if activities[i].Count != activities[j].Count {
			return activities[i].Count > activities[j].Count
		}
		return fixtureActivitySortKey(activities[i]) < fixtureActivitySortKey(activities[j])
	})
	return activities
}

func topFixtureActivities(values []FixtureActivity, limit int) []FixtureActivity {
	if len(values) == 0 || limit <= 0 {
		return nil
	}
	sort.Slice(values, func(i, j int) bool {
		if values[i].TotalDurationMS != values[j].TotalDurationMS {
			return values[i].TotalDurationMS > values[j].TotalDurationMS
		}
		return fixtureActivitySortKey(values[i]) < fixtureActivitySortKey(values[j])
	})
	if len(values) > limit {
		values = values[:limit]
	}
	return values
}

func fixtureKey(parts ...string) string {
	trimmed := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed = append(trimmed, strings.TrimSpace(part))
	}
	return strings.Join(trimmed, "\x1f")
}

func fixtureActivitySortKey(activity FixtureActivity) string {
	return strings.Join([]string{
		activity.Service,
		activity.Target,
		activity.Operation,
		activity.Strategy,
		activity.FixturePolicy,
		activity.ReuseScope,
		activity.CallerPackage,
		activity.TestName,
	}, "\x1f")
}

func strategyForPostgresCreateEvent(event Event) string {
	if strategy := stringDetail(event.Details, "preparation_strategy"); strategy != "" {
		return normalizePostgresPreparationStrategy(strategy)
	}
	switch event.Kind {
	case PostgresPreparationTemplate:
		return PostgresPreparationTemplate
	case PostgresPreparationTemplateClone:
		return PostgresPreparationTemplateClone
	case "scratch":
		return PostgresPreparationFreshMigration
	default:
		return ""
	}
}

func strategyForPostgresMigratedEvent(event Event) string {
	if strategy := stringDetail(event.Details, "preparation_strategy"); strategy != "" {
		return normalizePostgresPreparationStrategy(strategy)
	}
	if event.Kind == PostgresPreparationTemplate {
		return PostgresPreparationTemplate
	}
	return PostgresPreparationFreshMigration
}

func normalizePostgresPreparationStrategy(strategy string) string {
	switch strategy {
	case PostgresPreparationTemplate, PostgresPreparationTemplateClone, PostgresPreparationFreshMigration:
		return strategy
	default:
		return ""
	}
}

func upsertPostgresPreparation(preparations map[string]PostgresDatabasePreparation, event Event, strategy string) {
	if event.Name == "" {
		return
	}

	existing := preparations[event.Name]
	if existing.Name == "" {
		existing.Name = event.Name
	}
	if existing.Timestamp == "" || event.Timestamp < existing.Timestamp {
		existing.Timestamp = event.Timestamp
	}
	if existing.PID == 0 {
		existing.PID = event.PID
	}
	if strategy != "" {
		existing.Strategy = strategy
	}
	if target := stringDetail(event.Details, "target"); target != "" {
		existing.Target = target
	}
	if templateDB := stringDetail(event.Details, "template_database"); templateDB != "" {
		existing.TemplateDatabase = templateDB
	}
	preparations[event.Name] = existing
}

func sortedPostgresPreparations(preparations map[string]PostgresDatabasePreparation) []PostgresDatabasePreparation {
	if len(preparations) == 0 {
		return nil
	}
	values := make([]PostgresDatabasePreparation, 0, len(preparations))
	for _, preparation := range preparations {
		values = append(values, preparation)
	}
	sort.Slice(values, func(i, j int) bool {
		if values[i].Name != values[j].Name {
			return values[i].Name < values[j].Name
		}
		return values[i].Timestamp < values[j].Timestamp
	})
	return values
}

func upsertWebE2EFixture(fixtures map[string]BrowserE2EFixtureRef, event Event) {
	fixture := BrowserE2EFixtureRef{
		DatabaseName: stringDetail(event.Details, "database_name"),
		Bucket:       stringDetail(event.Details, "bucket"),
		Target:       stringDetail(event.Details, "target"),
		Timestamp:    event.Timestamp,
		PID:          event.PID,
	}
	key := fixtureKey(fixture.DatabaseName, fixture.Bucket, fixture.Target)
	if key == "" {
		return
	}
	existing := fixtures[key]
	if existing.Timestamp == "" || fixture.Timestamp > existing.Timestamp {
		fixtures[key] = fixture
	}
}

func sortedWebE2EFixtures(fixtures map[string]BrowserE2EFixtureRef) []BrowserE2EFixtureRef {
	if len(fixtures) == 0 {
		return nil
	}
	values := make([]BrowserE2EFixtureRef, 0, len(fixtures))
	for _, fixture := range fixtures {
		values = append(values, fixture)
	}
	sort.Slice(values, func(i, j int) bool {
		if values[i].Target != values[j].Target {
			return values[i].Target < values[j].Target
		}
		if values[i].DatabaseName != values[j].DatabaseName {
			return values[i].DatabaseName < values[j].DatabaseName
		}
		if values[i].Bucket != values[j].Bucket {
			return values[i].Bucket < values[j].Bucket
		}
		return values[i].Timestamp < values[j].Timestamp
	})
	return values
}

func sanitizeFileComponent(value string) string {
	lower := strings.ToLower(value)
	lower = strings.ReplaceAll(lower, ":", "-")
	lower = strings.ReplaceAll(lower, ".", "-")
	lower = strings.ReplaceAll(lower, "_", "-")
	lower = strings.ReplaceAll(lower, "/", "-")
	lower = strings.TrimSpace(lower)
	if lower == "" {
		return "event"
	}

	var builder strings.Builder
	for _, r := range lower {
		switch {
		case r >= 'a' && r <= 'z':
			builder.WriteRune(r)
		case r >= '0' && r <= '9':
			builder.WriteRune(r)
		case r == '-':
			builder.WriteRune(r)
		default:
			builder.WriteByte('-')
		}
	}

	result := strings.Trim(builder.String(), "-")
	if result == "" {
		return "event"
	}
	return result
}

func earliestTimestamp(current string, candidate string) string {
	if current == "" || candidate < current {
		return candidate
	}
	return current
}

func latestTimestamp(current string, candidate string) string {
	if current == "" || candidate > current {
		return candidate
	}
	return current
}

func sortedKeys(values map[string]struct{}) []string {
	if len(values) == 0 {
		return nil
	}
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func stringDetail(details map[string]any, key string) string {
	if details == nil {
		return ""
	}
	value, ok := details[key]
	if !ok {
		return ""
	}
	text, ok := value.(string)
	if !ok {
		return ""
	}
	return text
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func boolDetail(details map[string]any, key string) bool {
	if details == nil {
		return false
	}
	value, ok := details[key]
	if !ok {
		return false
	}
	boolValue, ok := value.(bool)
	return ok && boolValue
}

func intDetail(details map[string]any, key string) (int, bool) {
	if details == nil {
		return 0, false
	}
	value, ok := details[key]
	if !ok {
		return 0, false
	}
	switch typed := value.(type) {
	case float64:
		return int(typed), true
	case int:
		return typed, true
	default:
		return 0, false
	}
}

func intValue(details map[string]any, key string) int {
	value, _ := intDetail(details, key)
	return value
}

func int64Value(details map[string]any, key string) int64 {
	if details == nil {
		return 0
	}
	value, ok := details[key]
	if !ok {
		return 0
	}
	switch typed := value.(type) {
	case float64:
		return int64(typed)
	case int:
		return int64(typed)
	case int64:
		return typed
	default:
		return 0
	}
}
