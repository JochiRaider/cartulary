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
	EventWrapperOwnedStart   = "wrapper-owned-start"
	EventWrapperPassThrough  = "wrapper-pass-through"
	EventServiceStarted      = "service-started"
	EventCleanupCompleted    = "cleanup-completed"
	EventPostgresAttach      = "postgres-attach"
	EventPostgresDBCreated   = "postgres-db-created"
	EventPostgresDBMigrated  = "postgres-db-migrated"
	EventPostgresTemplateUse = "postgres-template-clone"
	EventS3Attach            = "s3-attach"
	EventS3BucketCreated     = "s3-bucket-created"
	EventS3BucketCleaned     = "s3-bucket-cleaned"
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
	Cleanup      CleanupSummary       `json:"cleanup"`
	Postgres     PostgresSummary      `json:"postgres"`
	MinIO        MinIOSummary         `json:"minio"`
	StartedNames StartedServiceRecord `json:"started_services"`
}

type WrapperSummary struct {
	OwnedCount       int `json:"owned_count"`
	PassThroughCount int `json:"pass_through_count"`
}

type CleanupSummary struct {
	Status          string `json:"status,omitempty"`
	CompletedAt     string `json:"completed_at,omitempty"`
	ChildExitStatus *int   `json:"child_exit_status,omitempty"`
}

type PostgresSummary struct {
	Started               bool     `json:"started"`
	StartedAt             string   `json:"started_at,omitempty"`
	Host                  string   `json:"host,omitempty"`
	Port                  string   `json:"port,omitempty"`
	User                  string   `json:"user,omitempty"`
	TemplateDatabase      string   `json:"template_database,omitempty"`
	AttachedHarnessCount  int      `json:"attached_harness_count"`
	CreatedDatabaseCount  int      `json:"created_database_count"`
	MigratedDatabaseCount int      `json:"migrated_database_count"`
	TemplateCloneCount    int      `json:"template_clone_count"`
	CreatedDatabases      []string `json:"created_databases,omitempty"`
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
		case EventPostgresDBMigrated:
			scope.Postgres.MigratedDatabaseCount++
		case EventPostgresTemplateUse:
			scope.Postgres.TemplateCloneCount++
		case EventS3Attach:
			scope.MinIO.AttachedHarnessCount++
		case EventS3BucketCreated:
			scope.MinIO.BucketCreateCount++
			if event.Name != "" {
				createdBuckets[event.Name] = struct{}{}
			}
		case EventS3BucketCleaned:
			scope.MinIO.BucketCleanupCount++
			if event.Name != "" {
				cleanedBuckets[event.Name] = struct{}{}
			}
		}
	}

	scope.Postgres.CreatedDatabases = sortedKeys(createdDatabases)
	scope.MinIO.CreatedBuckets = sortedKeys(createdBuckets)
	scope.MinIO.CleanedBuckets = sortedKeys(cleanedBuckets)
	scope.StartedNames.Names = sortedKeys(startedServices)

	return scope, true, nil
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
