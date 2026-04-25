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

func assertPreparation(t testing.TB, got PostgresDatabasePreparation, want PostgresDatabasePreparation) {
	t.Helper()

	if got != want {
		t.Fatalf("unexpected database preparation:\ngot  %#v\nwant %#v", got, want)
	}
}
