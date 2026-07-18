package suiteservices

import "testing"

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
				"target":               "test-fast-service-backed",
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
				"target":               "test-fast-service-backed",
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
		Target:    "test-fast-service-backed",
		PID:       101,
		Timestamp: "2026-04-25T12:00:00Z",
	})
}

func TestRecordLifecycleEventTracksConcurrentChildrenAndIllegalTransitions(t *testing.T) {
	env := map[string]string{
		SuiteIDEnv:        "0123456789abcdef01234567",
		TargetEnv:         "check-service-backed",
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
		TargetEnv:         "check-service-backed",
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
		TargetEnv:         "check-service-backed",
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
			Type:      EventPostgresDBReset,
			Timestamp: "2026-04-25T12:00:01Z",
			PID:       101,
			Name:      "ct_auth",
			Details: map[string]any{
				"duration_ms":    float64(7),
				"fixture_policy": PostgresFixturePolicyPackageReset,
				"reuse_scope":    FixtureReusePackage,
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
	if scope.BrowserE2E.RetiredFixtureCount != 2 || len(scope.BrowserE2E.RetiredFixtures) != 1 {
		t.Fatalf("expected deduplicated retired browser fixture summary, got %#v", scope.BrowserE2E)
	}
	if scope.BrowserE2E.CleanedFixtureCount != 1 || len(scope.BrowserE2E.CleanedFixtures) != 1 {
		t.Fatalf("expected cleaned browser fixture summary, got %#v", scope.BrowserE2E)
	}
	if scope.BrowserE2E.ReclaimedFixtureCount != 1 || len(scope.BrowserE2E.ReclaimedFixtures) != 1 {
		t.Fatalf("expected reclaimed browser fixture summary, got %#v", scope.BrowserE2E)
	}
	if scope.BrowserE2E.ReclaimedFixtures[0].ReclaimStrategy != "owned_stack_termination" {
		t.Fatalf("expected reclaim strategy in browser fixture summary, got %#v", scope.BrowserE2E.ReclaimedFixtures[0])
	}
	if scope.BrowserE2E.RetiredFixtures[0].Timestamp != "2026-04-25T12:00:01Z" {
		t.Fatalf("expected latest retired fixture event to win, got %#v", scope.BrowserE2E.RetiredFixtures[0])
	}
}

func assertPreparation(t testing.TB, got PostgresDatabasePreparation, want PostgresDatabasePreparation) {
	t.Helper()

	if got != want {
		t.Fatalf("unexpected database preparation:\ngot  %#v\nwant %#v", got, want)
	}
}
