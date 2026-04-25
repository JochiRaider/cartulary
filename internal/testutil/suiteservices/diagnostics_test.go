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
				"reuse_scope":          FixtureReusePerTest,
				"caller_package":       "internal/modules/auth",
				"caller_file":          "internal/modules/auth/phase1_integration_test.go",
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
				"reuse_scope":    FixtureReusePackage,
				"caller_package": "internal/modules/auth",
				"caller_file":    "internal/modules/auth/phase1_integration_test.go",
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
	if len(scope.Fixture.Slowest) == 0 || scope.Fixture.Slowest[0].TestName != "TestAuth" {
		t.Fatalf("expected slowest fixture event first, got %#v", scope.Fixture.Slowest)
	}
}

func assertPreparation(t testing.TB, got PostgresDatabasePreparation, want PostgresDatabasePreparation) {
	t.Helper()

	if got != want {
		t.Fatalf("unexpected database preparation:\ngot  %#v\nwant %#v", got, want)
	}
}
