package suiteservices

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func TestProducerJournalIsBoundedAndRejectsCompletedCorruption(t *testing.T) {
	resultsRoot := t.TempDir()
	env := map[string]string{
		SuiteIDEnv:        "suite-journal",
		TargetEnv:         "check",
		testResultsDirEnv: resultsRoot,
		testRunIDEnv:      "run-journal",
	}
	for index := range 12 {
		if err := RecordEvent(env, Event{
			Type:      EventFailureRecorded,
			Timestamp: fmt.Sprintf("2026-08-17T12:00:%02dZ", index),
			PID:       8080,
			Details: map[string]any{
				"failure_class":  FailureClassInfra,
				"failure_reason": "service_start_error",
			},
		}); err != nil {
			t.Fatal(err)
		}
	}
	suiteDir, _, err := ResolveSuiteArtifactDir(env)
	if err != nil {
		t.Fatal(err)
	}
	journals, err := filepath.Glob(filepath.Join(suiteDir, "journals", "*.ndjson"))
	if err != nil || len(journals) != 1 {
		t.Fatalf("expected one producer journal, got %v (%v)", journals, err)
	}
	scope, ok, err := Summarize(env)
	if err != nil || !ok {
		t.Fatalf("summarize journal: ok=%t err=%v", ok, err)
	}
	if scope.Failures.Counts[FailureClassInfra] != 12 || len(scope.Failures.Exemplars[FailureClassInfra]) != 10 {
		t.Fatalf("failure counts/exemplars are not bounded: %#v", scope.Failures)
	}
	if err := os.Chmod(journals[0], 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := Summarize(env); err == nil {
		t.Fatal("reader must reject a producer journal with non-owner permissions")
	}
	if err := os.Chmod(journals[0], 0o600); err != nil {
		t.Fatal(err)
	}
	journal, err := os.OpenFile(journals[0], os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := journal.WriteString("{"); err != nil {
		t.Fatal(err)
	}
	if err := journal.Close(); err != nil {
		t.Fatal(err)
	}
	if _, _, err := Summarize(env); err != nil {
		t.Fatalf("incomplete trailing crash record must be ignored: %v", err)
	}
	journal, err = os.OpenFile(journals[0], os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := journal.WriteString("\n"); err != nil {
		t.Fatal(err)
	}
	if err := journal.Close(); err != nil {
		t.Fatal(err)
	}
	if _, _, err := Summarize(env); err == nil {
		t.Fatal("malformed completed journal record must be rejected")
	}
}

func TestProducerJournalRejectsSymlinkDestination(t *testing.T) {
	resultsRoot := t.TempDir()
	env := map[string]string{
		SuiteIDEnv:        "suite-journal-symlink",
		TargetEnv:         "check",
		testResultsDirEnv: resultsRoot,
		testRunIDEnv:      "run-journal-symlink",
	}
	suiteDir, _, err := ResolveSuiteArtifactDir(env)
	if err != nil {
		t.Fatal(err)
	}
	journalDir := filepath.Join(suiteDir, "journals")
	if err := os.MkdirAll(journalDir, 0o700); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(t.TempDir(), "outside.ndjson")
	if err := os.WriteFile(target, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, filepath.Join(journalDir, producerIdentity+".ndjson")); err != nil {
		t.Fatal(err)
	}
	if err := RecordEvent(env, Event{Type: EventWrapperOwnedStart}); err == nil {
		t.Fatal("producer journal writer must reject a symlink destination")
	}
}

func TestProducerJournalRejectsSequenceGapsAndDuplicates(t *testing.T) {
	for _, testCase := range []struct {
		name      string
		sequences []uint64
	}{
		{name: "gap", sequences: []uint64{2}},
		{name: "duplicate", sequences: []uint64{1, 1}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			resultsRoot := t.TempDir()
			env := map[string]string{
				SuiteIDEnv:        "suite-journal-" + testCase.name,
				TargetEnv:         "check",
				testResultsDirEnv: resultsRoot,
				testRunIDEnv:      "run-journal-" + testCase.name,
			}
			suiteDir, _, err := ResolveSuiteArtifactDir(env)
			if err != nil {
				t.Fatal(err)
			}
			journalDir := filepath.Join(suiteDir, "journals")
			if err := os.MkdirAll(journalDir, 0o700); err != nil {
				t.Fatal(err)
			}
			producerID := "producer-identity"
			var records strings.Builder
			for index, sequence := range testCase.sequences {
				records.WriteString(fmt.Sprintf(
					`{"schema_id":"cartulary.test_services.journal_event.v1","producer_id":%q,"seq":%d,"type":"wrapper-owned-start","timestamp":"2026-08-17T12:00:%02dZ","pid":8080}`+"\n",
					producerID, sequence, index,
				))
			}
			if err := os.WriteFile(filepath.Join(journalDir, producerID+".ndjson"), []byte(records.String()), 0o600); err != nil {
				t.Fatal(err)
			}
			if _, _, err := Summarize(env); err == nil {
				t.Fatalf("completed journal with %s must be rejected", testCase.name)
			}
		})
	}
}

func TestScopeV2OmitsExactResourcesAndPrivateLedgerIsOwnerOnly(t *testing.T) {
	resultsRoot := t.TempDir()
	env := map[string]string{
		SuiteIDEnv:        "suite-ledger",
		TargetEnv:         "browser-e2e",
		testResultsDirEnv: resultsRoot,
		testRunIDEnv:      "run-ledger",
	}
	for _, event := range []Event{
		{Type: EventServiceStarted, Service: ServicePostgres, PID: 9090, Details: map[string]any{"host": "forbidden.example", "port": "5432", "user": "cartulary"}},
		{Type: EventServiceStarted, Service: ServiceObjectStore, PID: 9090, Details: map[string]any{"endpoint": "forbidden.example:9000", "secure": false}},
		{Type: EventPostgresDBCreated, Name: "ct_private", PID: 9090},
		{Type: EventS3BucketCreated, Name: "ct-private", PID: 9090},
		{Type: EventWebE2EFixtureRetired, PID: 9090, Details: map[string]any{"database_name": "ct_private", "bucket": "ct-private", "target": "browser-e2e"}},
		{
			Type:    EventFailureRecorded,
			Service: ServiceObjectStore,
			PID:     9090,
			Details: map[string]any{
				"failure_class":  FailureClassInfra,
				"failure_reason": "service_readiness_timeout",
				ObjectStoreReadinessExtensionKey: ObjectStoreReadinessDiagnostic{
					SchemaID:            "cartulary.object_store_readiness_diagnostic.v1",
					Phase:               "initial_lane",
					Stage:               "list",
					Outcome:             "deadline_expired",
					AttemptCount:        3,
					AttemptTimeoutCount: 1,
					ElapsedMS:           120000,
					CleanupOutcome:      "not_needed",
					CauseCounts: map[string]int{
						"operation_timeout":     1,
						"transport_unreachable": 2,
					},
					ContainerState: "running",
					HealthState:    "none",
				},
			},
		},
	} {
		if err := RecordEvent(env, event); err != nil {
			t.Fatal(err)
		}
	}
	if err := RefreshSummary(env); err != nil {
		t.Fatal(err)
	}
	scope, ok, err := Summarize(env)
	if err != nil || !ok {
		t.Fatalf("summarize readiness diagnostic: ok=%t err=%v", ok, err)
	}
	diagnostic, ok := scope.Extensions[ObjectStoreReadinessExtensionKey].(ObjectStoreReadinessDiagnostic)
	if !ok || diagnostic.Stage != "list" || diagnostic.AttemptCount != 3 || diagnostic.CauseCounts["transport_unreachable"] != 2 {
		t.Fatalf("unexpected readiness diagnostic: %#v", scope.Extensions)
	}
	suiteDir, _, _ := ResolveSuiteArtifactDir(env)
	summary, err := os.ReadFile(filepath.Join(suiteDir, "service-scope.json"))
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"created_databases", "created_buckets", "retired_fixtures", "by_package", "by_test", "access_key", "secret_key", "docker_endpoint", "endpoint", "forbidden.example", "raw_error", "log_tail"} {
		if strings.Contains(string(summary), forbidden) {
			t.Fatalf("bounded public scope retained %q", forbidden)
		}
	}
	ledgerPath := filepath.Join(suiteDir, "resource-ledger.json")
	info, err := os.Lstat(ledgerPath)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 || !info.Mode().IsRegular() {
		t.Fatalf("resource ledger mode/type = %v", info.Mode())
	}
	ledger, ok, err := ReadResourceLedger(env)
	if err != nil || !ok || len(ledger.Databases) != 1 || len(ledger.Buckets) != 1 || len(ledger.BrowserFixtures) != 1 {
		t.Fatalf("private ledger lost exact cleanup authority: ok=%t err=%v ledger=%#v", ok, err, ledger)
	}
	if err := os.Chmod(ledgerPath, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := ReadResourceLedger(env); err == nil {
		t.Fatal("resource ledger reader must reject non-owner permissions")
	}
	if err := os.Chmod(ledgerPath, 0o600); err != nil {
		t.Fatal(err)
	}
	ledger.SuiteID = "wrong-suite"
	tampered, err := json.Marshal(ledger)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(ledgerPath, append(tampered, '\n'), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := ReadResourceLedger(env); err == nil {
		t.Fatal("resource ledger reader must reject wrong suite identity")
	}
	if err := os.Remove(ledgerPath); err != nil {
		t.Fatal(err)
	}
	outside := filepath.Join(t.TempDir(), "outside-ledger.json")
	if err := os.WriteFile(outside, tampered, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, ledgerPath); err != nil {
		t.Fatal(err)
	}
	if _, _, err := ReadResourceLedger(env); err == nil {
		t.Fatal("resource ledger reader must reject symlinks")
	}
	if err := RemoveResourceLedger(env); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(ledgerPath); !os.IsNotExist(err) {
		t.Fatalf("successful cleanup must remove resource ledger: %v", err)
	}

	malformedEnv := map[string]string{
		SuiteIDEnv:        "suite-malformed-readiness",
		TargetEnv:         "check",
		testResultsDirEnv: resultsRoot,
		testRunIDEnv:      "run-malformed-readiness",
	}
	if err := RecordEvent(malformedEnv, Event{
		Type:    EventFailureRecorded,
		Service: ServiceObjectStore,
		Details: map[string]any{
			ObjectStoreReadinessExtensionKey: map[string]any{
				"schema_id":             "cartulary.object_store_readiness_diagnostic.v1",
				"phase":                 "initial_lane",
				"stage":                 "list",
				"outcome":               "deadline_expired",
				"attempt_count":         1,
				"attempt_timeout_count": 0,
				"elapsed_ms":            1,
				"cleanup_outcome":       "not_needed",
				"cause_counts":          map[string]int{"transport_unreachable": 1},
				"endpoint":              "forbidden.example:9000",
			},
		},
	}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := Summarize(malformedEnv); err == nil {
		t.Fatal("malformed readiness diagnostic must be rejected")
	}

	maximumDiagnostic := ObjectStoreReadinessDiagnostic{
		SchemaID:            "cartulary.object_store_readiness_diagnostic.v1",
		Phase:               "package_admission",
		Stage:               "delete_verify",
		Outcome:             "deadline_expired",
		AttemptCount:        10000,
		AttemptTimeoutCount: 10000,
		ElapsedMS:           600000,
		CleanupOutcome:      "completed",
		CauseCounts:         map[string]int{"operation_timeout": 10000},
	}
	raw := map[string]any{ObjectStoreReadinessExtensionKey: maximumDiagnostic}
	decoded, present, err := objectStoreReadinessDiagnosticDetail(raw)
	if err != nil || !present || decoded.AttemptCount != 10000 {
		t.Fatalf("maximum bounded diagnostic rejected: present=%t diagnostic=%#v err=%v", present, decoded, err)
	}
	first, err := json.Marshal(maximumDiagnostic)
	if err != nil {
		t.Fatal(err)
	}
	second, err := json.Marshal(maximumDiagnostic)
	if err != nil {
		t.Fatal(err)
	}
	if string(first) != string(second) {
		t.Fatalf("diagnostic serialization is not deterministic: %s != %s", first, second)
	}

	maximumDiagnostic.ElapsedMS++
	if _, _, err := objectStoreReadinessDiagnosticDetail(map[string]any{
		ObjectStoreReadinessExtensionKey: maximumDiagnostic,
	}); err == nil {
		t.Fatal("out-of-bounds diagnostic must be rejected")
	}
}

func TestRefreshSummaryPublishesCompleteSnapshotsUnderConcurrency(t *testing.T) {
	resultsRoot := t.TempDir()
	env := map[string]string{
		SuiteIDEnv:        "suite-atomic-summary",
		TargetEnv:         "check",
		testResultsDirEnv: resultsRoot,
		testRunIDEnv:      "run-atomic-summary",
	}
	if err := RecordEvent(env, Event{Type: EventServiceStarted, Timestamp: "2026-04-25T12:00:00Z", PID: 101, Name: ServicePostgres}); err != nil {
		t.Fatalf("record service event: %v", err)
	}
	if err := RefreshSummary(env); err != nil {
		t.Fatalf("write initial summary: %v", err)
	}
	summaryPath := filepath.Join(resultsRoot, "run-atomic-summary", "_shared", "test-services", "suite-atomic-summary", "service-scope.json")

	const refreshers = 8
	const reads = 250
	var writers sync.WaitGroup
	writers.Add(refreshers)
	for range refreshers {
		go func() {
			defer writers.Done()
			for range 20 {
				if err := RefreshSummary(env); err != nil {
					t.Errorf("refresh summary: %v", err)
					return
				}
			}
		}()
	}
	for range reads {
		raw, err := os.ReadFile(summaryPath)
		if err != nil {
			t.Fatalf("read summary: %v", err)
		}
		var scope ServiceScope
		if err := json.Unmarshal(raw, &scope); err != nil {
			t.Fatalf("reader observed incomplete summary: %v", err)
		}
		if scope.SchemaID != "cartulary.test_services.scope.v2" {
			t.Fatalf("unexpected scope schema %q", scope.SchemaID)
		}
	}
	writers.Wait()
}

func TestSummarizeReportsPostgresDatabasePreparations(t *testing.T) {
	env := map[string]string{
		SuiteIDEnv:        "suite-preparations",
		TargetEnv:         "backend-store",
		testResultsDirEnv: t.TempDir(),
		testRunIDEnv:      "run-preparations",
	}

	events := []Event{
		{
			Type:      EventPostgresDBCreated,
			Timestamp: "2026-04-25T12:00:00Z",
			PID:       101,
			Name:      "ct_template",
			Kind:      PostgresPreparationTemplate,
			Details: map[string]any{
				"preparation_strategy": PostgresPreparationTemplate,
				"target":               "test-fast",
			},
		},
		{
			Type:      EventPostgresDBMigrated,
			Timestamp: "2026-04-25T12:00:01Z",
			PID:       101,
			Name:      "ct_template",
			Kind:      PostgresPreparationTemplate,
			Details: map[string]any{
				"preparation_strategy": PostgresPreparationTemplate,
				"target":               "test-fast",
			},
		},
		{
			Type:      EventPostgresDBCreated,
			Timestamp: "2026-04-25T12:00:02Z",
			PID:       202,
			Name:      "ct_clone",
			Kind:      PostgresPreparationTemplateClone,
			Details: map[string]any{
				"preparation_strategy": PostgresPreparationTemplateClone,
				"template_database":    "ct_template",
				"target":               "browser-e2e-webserver-backed",
			},
		},
		{
			Type:      EventPostgresTemplateUse,
			Timestamp: "2026-04-25T12:00:03Z",
			PID:       202,
			Name:      "ct_clone",
			Details: map[string]any{
				"template_database": "ct_template",
				"target":            "browser-e2e-webserver-backed",
			},
		},
		{
			Type:      EventPostgresDBCreated,
			Timestamp: "2026-04-25T12:00:04Z",
			PID:       303,
			Name:      "ct_fresh",
			Kind:      "scratch",
			Details: map[string]any{
				"preparation_strategy": PostgresPreparationFreshMigration,
				"target":               "migration-behavior",
			},
		},
	}

	for _, event := range events {
		if err := RecordEvent(env, event); err != nil {
			t.Fatalf("record event %s/%s: %v", event.Type, event.Name, err)
		}
	}

	scope, ok, err := Summarize(env)
	if err != nil {
		t.Fatalf("summarize suite services: %v", err)
	}
	if !ok {
		t.Fatal("expected suite summary")
	}

	preparations := scope.Postgres.DatabasePreparations
	if len(preparations) != 3 {
		t.Fatalf("expected three database preparations, got %#v", preparations)
	}

	assertPreparation(t, preparations[0], PostgresDatabasePreparation{
		Name:             "ct_clone",
		Strategy:         PostgresPreparationTemplateClone,
		TemplateDatabase: "ct_template",
		Target:           "browser-e2e-webserver-backed",
		PID:              202,
		Timestamp:        "2026-04-25T12:00:02Z",
	})
	assertPreparation(t, preparations[1], PostgresDatabasePreparation{
		Name:      "ct_fresh",
		Strategy:  PostgresPreparationFreshMigration,
		Target:    "migration-behavior",
		PID:       303,
		Timestamp: "2026-04-25T12:00:04Z",
	})
	assertPreparation(t, preparations[2], PostgresDatabasePreparation{
		Name:      "ct_template",
		Strategy:  PostgresPreparationTemplate,
		Target:    "test-fast",
		PID:       101,
		Timestamp: "2026-04-25T12:00:00Z",
	})
}

func TestRecordLifecycleEventTracksConcurrentChildrenAndIllegalTransitions(t *testing.T) {
	env := map[string]string{
		SuiteIDEnv:        "0123456789abcdef01234567",
		TargetEnv:         "check",
		testResultsDirEnv: t.TempDir(),
		testRunIDEnv:      "run-lifecycle",
		LifecycleModeEnv:  "owned",
	}

	for _, step := range []struct {
		event    string
		childKey string
	}{
		{LifecycleEventStartServices, ""},
		{LifecycleEventReadinessPassed, ""},
		{LifecycleEventChildStarted, "alpha"},
		{LifecycleEventChildStarted, "beta"},
		{LifecycleEventChildFinished, "alpha"},
	} {
		if err := RecordLifecycleEvent(env, step.event, step.childKey); err != nil {
			t.Fatalf("record lifecycle %s/%s: %v", step.event, step.childKey, err)
		}
	}

	if err := RecordLifecycleEvent(env, LifecycleEventChildFinished, "missing"); err == nil {
		t.Fatal("expected missing child finish to be illegal")
	}
	if err := RecordLifecycleEvent(env, LifecycleEventChildFinished, "beta"); err != nil {
		t.Fatalf("legal child finish after illegal transition failed: %v", err)
	}

	records, err := ReadLifecycleEvents(env)
	if err != nil {
		t.Fatalf("read lifecycle events: %v", err)
	}
	if len(records) != 7 {
		t.Fatalf("expected seven lifecycle records, got %#v", records)
	}
	if records[2].Event != LifecycleEventChildStarted || records[2].ActiveChildCount != 1 {
		t.Fatalf("unexpected first child count: %#v", records[2])
	}
	if records[3].Event != LifecycleEventChildStarted || records[3].ActiveChildCount != 2 {
		t.Fatalf("unexpected second child count: %#v", records[3])
	}
	illegal := records[5]
	if illegal.TransitionStatus != "illegal" || illegal.FromState != illegal.ToState {
		t.Fatalf("illegal transition must be recorded without state mutation: %#v", illegal)
	}
	if illegal.FailureClass == nil || *illegal.FailureClass != FailureClassHelper {
		t.Fatalf("illegal transition failure class: %#v", illegal.FailureClass)
	}
	if illegal.FailureReason == nil || *illegal.FailureReason != "scheduler_accounting_error" {
		t.Fatalf("illegal transition failure reason: %#v", illegal.FailureReason)
	}
	if records[6].Event != LifecycleEventChildFinished || records[6].ActiveChildCount != 0 || records[6].ToState != "ready" {
		t.Fatalf("legal child finish after illegal transition did not close runtime state: %#v", records[6])
	}
}

func TestRecordLifecycleFailureEventWritesFailureFields(t *testing.T) {
	env := map[string]string{
		SuiteIDEnv:        "abcdef0123456789abcdef01",
		TargetEnv:         "check",
		testResultsDirEnv: t.TempDir(),
		testRunIDEnv:      "run-lifecycle-failure",
		LifecycleModeEnv:  "owned",
	}

	if err := RecordLifecycleEvent(env, LifecycleEventStartServices, ""); err != nil {
		t.Fatalf("record lifecycle start: %v", err)
	}
	if err := RecordLifecycleFailureEvent(env, LifecycleEventStartupFailed, "", FailureClassInfra, "preflight_error"); err != nil {
		t.Fatalf("record startup failure: %v", err)
	}

	records, err := ReadLifecycleEvents(env)
	if err != nil {
		t.Fatalf("read lifecycle events: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("expected two lifecycle records, got %#v", records)
	}
	startupFailed := records[1]
	if startupFailed.Event != LifecycleEventStartupFailed || startupFailed.ToState != "failed_start" {
		t.Fatalf("unexpected startup failure event: %#v", startupFailed)
	}
	if startupFailed.FailureClass == nil || *startupFailed.FailureClass != FailureClassInfra {
		t.Fatalf("startup failure class: %#v", startupFailed.FailureClass)
	}
	if startupFailed.FailureReason == nil || *startupFailed.FailureReason != "preflight_error" {
		t.Fatalf("startup failure reason: %#v", startupFailed.FailureReason)
	}
}

func TestRecordLifecycleEventRejectsDuplicateChildAndTerminalMutation(t *testing.T) {
	env := map[string]string{
		SuiteIDEnv:        "abcdef0123456789abcdef01",
		TargetEnv:         "check",
		testResultsDirEnv: t.TempDir(),
		testRunIDEnv:      "run-lifecycle-terminal",
		LifecycleModeEnv:  "owned",
	}

	for _, step := range []struct {
		event    string
		childKey string
	}{
		{LifecycleEventStartServices, ""},
		{LifecycleEventReadinessPassed, ""},
		{LifecycleEventChildStarted, "alpha"},
	} {
		if err := RecordLifecycleEvent(env, step.event, step.childKey); err != nil {
			t.Fatalf("record lifecycle %s/%s: %v", step.event, step.childKey, err)
		}
	}
	if err := RecordLifecycleEvent(env, LifecycleEventChildStarted, "alpha"); err == nil {
		t.Fatal("expected duplicate child start to be illegal")
	}
	for _, step := range []struct {
		event    string
		childKey string
	}{
		{LifecycleEventChildFinished, "alpha"},
		{LifecycleEventInterruptReceived, ""},
		{LifecycleEventCleanupStarted, ""},
		{LifecycleEventCleanupSucceeded, ""},
	} {
		if err := RecordLifecycleEvent(env, step.event, step.childKey); err != nil {
			t.Fatalf("record lifecycle %s/%s: %v", step.event, step.childKey, err)
		}
	}
	if err := RecordLifecycleEvent(env, LifecycleEventChildFinished, "alpha"); err == nil {
		t.Fatal("expected terminal child finish to be illegal")
	}

	records, err := ReadLifecycleEvents(env)
	if err != nil {
		t.Fatalf("read lifecycle events: %v", err)
	}
	if len(records) != 9 {
		t.Fatalf("expected nine lifecycle records, got %#v", records)
	}
	duplicate := records[3]
	if duplicate.Event != LifecycleEventChildStarted || duplicate.TransitionStatus != "illegal" || duplicate.ActiveChildCount != 1 {
		t.Fatalf("duplicate child start must be recorded without state mutation: %#v", duplicate)
	}
	if duplicate.FailureReason == nil || *duplicate.FailureReason != "scheduler_accounting_error" {
		t.Fatalf("duplicate child start failure reason: %#v", duplicate.FailureReason)
	}
	terminal := records[8]
	if terminal.Event != LifecycleEventChildFinished || terminal.TransitionStatus != "illegal" || terminal.FromState != "cleaned" || terminal.ToState != "cleaned" {
		t.Fatalf("terminal lifecycle mutation must be rejected without state mutation: %#v", terminal)
	}
}

func TestSummarizeReportsFixtureActivity(t *testing.T) {
	env := map[string]string{
		SuiteIDEnv:        "suite-fixtures",
		TargetEnv:         "backend-integration",
		testResultsDirEnv: t.TempDir(),
		testRunIDEnv:      "run-fixtures",
	}

	events := []Event{
		{
			Type:      EventPostgresDBCreated,
			Timestamp: "2026-04-25T12:00:00Z",
			PID:       101,
			Name:      "ct_auth",
			Kind:      PostgresPreparationTemplateClone,
			Details: map[string]any{
				"duration_ms":          float64(25),
				"preparation_strategy": PostgresPreparationTemplateClone,
				"fixture_policy":       PostgresFixturePolicyTemplateClone,
				"reuse_scope":          FixtureReusePerTest,
				"caller_package":       "internal/modules/auth",
				"caller_file":          "internal/modules/auth/integration_test.go",
				"test_name":            "TestAuth",
				"target":               "backend-integration",
			},
		},
		{
			Type:      EventPostgresTransaction,
			Timestamp: "2026-04-25T12:00:01Z",
			PID:       101,
			Name:      "ct_auth",
			Details: map[string]any{
				"duration_ms":    float64(7),
				"fixture_policy": PostgresFixturePolicyTransaction,
				"reuse_scope":    FixtureReuseTransaction,
				"caller_package": "internal/modules/auth",
				"caller_file":    "internal/modules/auth/integration_test.go",
				"test_name":      "TestAuthReplay",
				"target":         "backend-integration",
			},
		},
		{
			Type:      EventS3PrefixCleaned,
			Timestamp: "2026-04-25T12:00:02Z",
			PID:       202,
			Name:      "ct-bucket",
			Details: map[string]any{
				"duration_ms":    float64(4),
				"strategy":       "prefix",
				"reuse_scope":    FixtureReusePrefix,
				"caller_package": "internal/modules/auth",
				"test_name":      "TestAuthReplay",
				"target":         "backend-integration",
			},
		},
	}

	for _, event := range events {
		if err := RecordEvent(env, event); err != nil {
			t.Fatalf("record event %s/%s: %v", event.Type, event.Name, err)
		}
	}

	scope, ok, err := Summarize(env)
	if err != nil {
		t.Fatalf("summarize suite services: %v", err)
	}
	if !ok {
		t.Fatal("expected suite summary")
	}

	if scope.Fixture.TotalCount != 3 {
		t.Fatalf("unexpected fixture count: %#v", scope.Fixture)
	}
	if scope.Fixture.TotalDurationMS != 36 {
		t.Fatalf("unexpected fixture duration: %#v", scope.Fixture)
	}
	if len(scope.Fixture.ByPackage) == 0 || scope.Fixture.ByPackage[0].CallerPackage != "internal/modules/auth" {
		t.Fatalf("expected package fixture aggregation, got %#v", scope.Fixture.ByPackage)
	}
	if len(scope.Fixture.ByStrategy) == 0 || scope.Fixture.ByStrategy[0].TotalDurationMS != 25 {
		t.Fatalf("expected strategy fixture aggregation sorted by cost, got %#v", scope.Fixture.ByStrategy)
	}
	if scope.Fixture.ByStrategy[0].FixturePolicy != PostgresFixturePolicyTemplateClone {
		t.Fatalf("expected fixture policy in strategy diagnostics, got %#v", scope.Fixture.ByStrategy[0])
	}
	if scope.Fixture.ByStrategy[0].Target != "backend-integration" {
		t.Fatalf("expected target in strategy diagnostics, got %#v", scope.Fixture.ByStrategy[0])
	}
	if len(scope.Fixture.Slowest) == 0 || scope.Fixture.Slowest[0].TestName != "TestAuth" {
		t.Fatalf("expected slowest fixture event first, got %#v", scope.Fixture.Slowest)
	}
}

func TestSummarizeBoundsStrategyDiagnosticsWithoutLosingTotals(t *testing.T) {
	env := map[string]string{
		SuiteIDEnv:        "suite-bounded-strategies",
		TargetEnv:         "release-check",
		testResultsDirEnv: t.TempDir(),
		testRunIDEnv:      "run-bounded-strategies",
	}

	for index := 0; index < 40; index++ {
		strategy := fmt.Sprintf("strategy-%02d", index)
		if err := RecordEvent(env, Event{
			Type:      EventPostgresDBCreated,
			Timestamp: fmt.Sprintf("2026-04-25T12:00:%02dZ", index),
			PID:       300 + index,
			Name:      fmt.Sprintf("ct_strategy_%02d", index),
			Details: map[string]any{
				"duration_ms":          int64(index),
				"preparation_strategy": strategy,
				"reuse_scope":          FixtureReusePerTest,
				"target":               "release-check",
			},
		}); err != nil {
			t.Fatalf("record strategy %s: %v", strategy, err)
		}
	}

	scope, ok, err := Summarize(env)
	if err != nil || !ok {
		t.Fatalf("summarize bounded strategies: ok=%t err=%v", ok, err)
	}
	if scope.Fixture.TotalCount != 40 || scope.Fixture.TotalDurationMS != 780 {
		t.Fatalf("fixture totals were truncated: %#v", scope.Fixture)
	}
	if scope.Fixture.StrategyAggregateCount != 40 {
		t.Fatalf("strategy aggregate count = %d, want 40", scope.Fixture.StrategyAggregateCount)
	}
	if len(scope.Fixture.ByStrategy) != 32 {
		t.Fatalf("retained strategy diagnostics = %d, want 32", len(scope.Fixture.ByStrategy))
	}
	if scope.Fixture.ByStrategy[0].Strategy != "strategy-39" ||
		scope.Fixture.ByStrategy[31].Strategy != "strategy-08" {
		t.Fatalf("bounded strategy ordering is not deterministic: first=%#v last=%#v", scope.Fixture.ByStrategy[0], scope.Fixture.ByStrategy[31])
	}
}

func TestSummarizeStrategyIdentityIncludesFixtureClassAndCountTieOrdering(t *testing.T) {
	env := map[string]string{
		SuiteIDEnv:        "suite-strategy-identity",
		TargetEnv:         "fixture-tie",
		testResultsDirEnv: t.TempDir(),
		testRunIDEnv:      "run-strategy-identity",
	}

	fixtures := []struct {
		timestamp    string
		durationMS   int64
		fixtureClass string
		testName     string
	}{
		{timestamp: "2026-04-25T12:00:00Z", durationMS: 10, fixtureClass: "reset", testName: "TestResetA"},
		{timestamp: "2026-04-25T12:00:01Z", durationMS: 10, fixtureClass: "reset", testName: "TestResetB"},
		{timestamp: "2026-04-25T12:00:02Z", durationMS: 20, fixtureClass: "transaction", testName: "TestTxn"},
	}
	for index, fixture := range fixtures {
		if err := RecordEvent(env, Event{
			Type:      EventPostgresTransaction,
			Timestamp: fixture.timestamp,
			PID:       400 + index,
			Name:      fmt.Sprintf("ct_fixture_%d", index),
			Details: map[string]any{
				"duration_ms":          fixture.durationMS,
				"preparation_strategy": PostgresPreparationTemplateClone,
				"fixture_policy":       PostgresFixturePolicyTransaction,
				"fixture_class":        fixture.fixtureClass,
				"reuse_scope":          FixtureReuseTransaction,
				"caller_package":       "internal/modules/auth",
				"test_name":            fixture.testName,
				"target":               "fixture-tie",
			},
		}); err != nil {
			t.Fatalf("record fixture %s: %v", fixture.testName, err)
		}
	}

	scope, ok, err := Summarize(env)
	if err != nil || !ok {
		t.Fatalf("summarize strategy identity: ok=%t err=%v", ok, err)
	}
	if scope.Fixture.StrategyAggregateCount != 2 || len(scope.Fixture.ByStrategy) != 2 {
		t.Fatalf("strategy aggregates = count %d values %#v, want two distinct fixture classes", scope.Fixture.StrategyAggregateCount, scope.Fixture.ByStrategy)
	}
	first := scope.Fixture.ByStrategy[0]
	second := scope.Fixture.ByStrategy[1]
	if first.Service != ServicePostgres || first.Target != "fixture-tie" || first.Operation != "transaction" ||
		first.Strategy != PostgresPreparationTemplateClone || first.FixturePolicy != PostgresFixturePolicyTransaction ||
		first.FixtureClass != "reset" || first.ReuseScope != FixtureReuseTransaction || first.Count != 2 || first.TotalDurationMS != 20 {
		t.Fatalf("first strategy aggregate = %#v, want exact reset aggregate", first)
	}
	if second.FixtureClass != "transaction" || second.Count != 1 || second.TotalDurationMS != 20 {
		t.Fatalf("second strategy aggregate = %#v, want exact transaction aggregate", second)
	}
}

func TestSummarizeReportsBrowserE2EFixtureLifecycle(t *testing.T) {
	env := map[string]string{
		SuiteIDEnv:        "suite-browser-fixtures",
		TargetEnv:         "browser-e2e-webserver-backed",
		testResultsDirEnv: t.TempDir(),
		testRunIDEnv:      "run-browser-fixtures",
	}

	events := []Event{
		{
			Type:      EventWebE2EFixtureRetired,
			Timestamp: "2026-04-25T12:00:00Z",
			PID:       101,
			Details: map[string]any{
				"database_name": "ct_web",
				"bucket":        "ct-web",
				"target":        "browser-e2e-webserver-backed",
			},
		},
		{
			Type:      EventWebE2EFixtureRetired,
			Timestamp: "2026-04-25T12:00:01Z",
			PID:       102,
			Details: map[string]any{
				"database_name": "ct_web",
				"bucket":        "ct-web",
				"target":        "browser-e2e-webserver-backed",
			},
		},
		{
			Type:      EventWebE2EFixtureCleaned,
			Timestamp: "2026-04-25T12:00:02Z",
			PID:       103,
			Details: map[string]any{
				"database_name": "ct_web",
				"bucket":        "ct-web",
				"target":        "browser-e2e-webserver-backed",
			},
		},
		{
			Type:      EventWebE2EFixtureReclaimed,
			Timestamp: "2026-04-25T12:00:03Z",
			PID:       104,
			Details: map[string]any{
				"database_name":    "ct_web",
				"bucket":           "ct-web",
				"target":           "browser-e2e-webserver-backed",
				"reclaim_strategy": "owned_stack_termination",
			},
		},
	}

	for _, event := range events {
		if err := RecordEvent(env, event); err != nil {
			t.Fatalf("record event %s: %v", event.Type, err)
		}
	}

	scope, ok, err := Summarize(env)
	if err != nil {
		t.Fatalf("summarize suite services: %v", err)
	}
	if !ok {
		t.Fatal("expected suite summary")
	}
	if scope.BrowserE2E.RetiredFixtureCount != 2 {
		t.Fatalf("expected full retired browser fixture count, got %#v", scope.BrowserE2E)
	}
	if scope.BrowserE2E.CleanedFixtureCount != 1 {
		t.Fatalf("expected cleaned browser fixture summary, got %#v", scope.BrowserE2E)
	}
	if scope.BrowserE2E.ReclaimedFixtureCount != 1 {
		t.Fatalf("expected reclaimed browser fixture summary, got %#v", scope.BrowserE2E)
	}
	ledger, found, err := CurrentResourceLedger(env)
	if err != nil || !found {
		t.Fatalf("read private resource ledger: found=%t err=%v", found, err)
	}
	if len(ledger.BrowserFixtures) != 0 {
		t.Fatalf("completed browser fixture remained live in private ledger: %#v", ledger.BrowserFixtures)
	}
}

func assertPreparation(t testing.TB, got PostgresDatabasePreparation, want PostgresDatabasePreparation) {
	t.Helper()

	if got != want {
		t.Fatalf("unexpected database preparation:\ngot  %#v\nwant %#v", got, want)
	}
}
