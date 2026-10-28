//go:build linux

package boundary

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices"
)

// Exercise installed selection and service ownership with exact existing rows.
// Docker unavailability is a setup failure, never a skipped assurance boundary.
func TestNativeSelectedServiceLifecycle(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../.."))
	results, outer := os.Getenv("CARTULARY_TEST_RESULTS_DIR"), os.Getenv("CARTULARY_TEST_RUN_ID")
	if results == "" || outer == "" {
		t.Fatal("Make-owned retained identity required")
	}
	if !filepath.IsAbs(results) {
		results = filepath.Join(root, results)
	}
	results = filepath.Join(results, outer, "selected-service-boundary")
	for _, selection := range []struct {
		id, target     string
		rows, services []string
	}{
		{"pure", "backend-unit", []string{"harness.command_surface.unit.configtest_helper_contract"}, nil},
		{"postgres", "backend-integration", []string{"harness.browser.integration.postgres_transaction_isolation"}, []string{"postgres"}},
		{"object-store", "backend-integration", []string{"harness.browser.integration.object_store_namespace_lifecycle"}, []string{"object_store"}},
		{"mixed", "backend-integration", []string{"harness.browser.integration.object_store_namespace_lifecycle", "harness.browser.integration.postgres_transaction_isolation", "platform.postgres.integration.fresh_fixture_admission"}, []string{"object_store", "postgres"}},
	} {
		t.Run(selection.id, func(t *testing.T) {
			env := map[string]string{}
			for _, entry := range os.Environ() {
				key, value, ok := strings.Cut(entry, "=")
				if !ok || (strings.HasPrefix(key, "CARTULARY_HARNESS_") && key != "CARTULARY_HARNESS_NESTED_LIFECYCLE") || strings.HasPrefix(key, "CARTULARY_TEST_SERVICES_") || strings.HasPrefix(key, "CARTULARY_PGTEST_") || strings.HasPrefix(key, "CARTULARY_S3TEST_") {
					continue
				}
				env[key] = value
			}
			for _, key := range []string{"MAKEFLAGS", "MAKEOVERRIDES", "MFLAGS", "JSON", "OWNER", "ROWS", "SERVICE_BACKED_ONLY"} {
				delete(env, key)
			}
			env["CARTULARY_TEST_RESULTS_DIR"] = results
			probeID := "native-selection-" + selection.id
			env["CARTULARY_TEST_RUN_ID"] = probeID
			env["HARNESS_DIAGNOSTICS"] = "off"
			if selection.id == "mixed" {
				env["HARNESS_DIAGNOSTICS"] = "basic"
			}
			env["CARTULARY_HARNESS_CACHE_MODE"] = "off"
			// Keep CARTULARY_HOST_ADMISSION_LEASE: the native weighted tree
			// accounts for child work under its real parent reservation.
			input := []string{}
			for key, value := range env {
				input = append(input, key+"="+value)
			}
			binary := os.Getenv("CARTULARY_HARNESS_RUNTIME_BIN")
			if binary == "" {
				binary = filepath.Join(root, ".cache/cartulary/toolbin/cartulary-harness")
			}
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()
			run := filepath.Join(results, probeID)
			runInstalledGraph(t, ctx, root, binary, []string{"--selection", "rows", "--target", "harness-contract", "--rows", strings.Join(selection.rows, ",")}, input, run)
			budget, err := strconv.Atoi(env["CARTULARY_UNIT_CPU_TOKENS"])
			if err != nil || budget < 1 || env["CARTULARY_HOST_ADMISSION_LEASE"] == "" {
				t.Fatal("selected lifecycle test requires its actual parent CPU reservation")
			}
			var retained struct {
				Graph struct {
					Units []struct {
						ID     string         `json:"unit_id"`
						Claims map[string]int `json:"resource_claims"`
					} `json:"units"`
				} `json:"graph"`
			}
			readSelectedBoundaryJSON(t, filepath.Join(run, "execution-context.json"), &retained)
			for _, unit := range retained.Graph.Units {
				if strings.HasPrefix(unit.ID, "go:") && unit.Claims["cpu"] > budget {
					t.Fatal("child Go execution exceeds its enclosing worker budget")
				}
			}
			for _, row := range selection.rows {
				var result struct {
					State string `json:"terminal_state"`
				}
				readSelectedBoundaryJSON(t, filepath.Join(run, "rows", row+".json"), &result)
				if result.State != "passed" {
					t.Fatalf("selected row %s did not execute successfully: %s", row, result.State)
				}
			}
			var cleanup struct {
				Results []struct {
					Outcome string `json:"outcome"`
				} `json:"results"`
			}
			readSelectedBoundaryJSON(t, filepath.Join(run, "cleanup-results.json"), &cleanup)
			if len(cleanup.Results) == 0 {
				t.Fatal("cleanup evidence is missing")
			}
			for _, result := range cleanup.Results {
				if result.Outcome != "completed" {
					t.Fatalf("incomplete cleanup: %s", result.Outcome)
				}
			}
			var scan struct {
				Status string `json:"status"`
			}
			readSelectedBoundaryJSON(t, filepath.Join(run, "retained-secret-scan.json"), &scan)
			if scan.Status != "pass" {
				t.Fatal("retained scan did not pass")
			}
			files, err := filepath.Glob(filepath.Join(run, "_shared", "test-services", "*", "service-scope.json"))
			if err != nil {
				t.Fatal(err)
			}
			if len(selection.services) == 0 {
				if len(files) != 0 {
					t.Fatal("pure selection acquired managed services")
				}
				return
			}
			expected := len(selection.services)
			if selection.id == "mixed" {
				expected++
			} // shared and exclusive PostgreSQL generations
			if len(files) != expected {
				t.Fatalf("physical service owners=%d; expected=%d", len(files), expected)
			}
			var scope suiteservices.ServiceScope
			started := map[string]int{}
			for _, file := range files {
				var part suiteservices.ServiceScope
				readSelectedBoundaryJSON(t, file, &part)
				if part.RunID != probeID || part.Wrapper.OwnedCount != 1 || part.Failure != nil || part.Cleanup.Status != "succeeded" {
					t.Fatal("physical service identity or cleanup did not settle")
				}
				if len(part.StartedNames.Names) != 1 {
					t.Fatal("physical owner mixed independent services")
				}
				started[part.StartedNames.Names[0]]++
				scope.Fixture.ByStrategy = append(scope.Fixture.ByStrategy, part.Fixture.ByStrategy...)
				scope.Postgres.CreatedDatabaseCount += part.Postgres.CreatedDatabaseCount
				scope.Postgres.DatabasePreparations = append(scope.Postgres.DatabasePreparations, part.Postgres.DatabasePreparations...)
				scope.ObjectStore.BucketCreateCount += part.ObjectStore.BucketCreateCount
				scope.ObjectStore.BucketCleanupCount += part.ObjectStore.BucketCleanupCount
			}
			names := []string{}
			for service := range started {
				names = append(names, service)
			}
			slices.Sort(names)
			if !slices.Equal(names, selection.services) {
				t.Fatalf("services=%v; expected=%v", names, selection.services)
			}
			if selection.id == "mixed" && started["postgres"] != 2 {
				t.Fatal("exclusive and shared database consumers shared a cluster")
			}
			if selection.id == "mixed" {
				var checkpoints struct {
					Services int `json:"services"`
					Results  []struct {
						Scope      string  `json:"service_scope_ref"`
						Generation string  `json:"generation"`
						Identity   *string `json:"identity_digest"`
						Status     string  `json:"status"`
					} `json:"service_results"`
				}
				readSelectedBoundaryJSON(t, filepath.Join(run, "diagnostics", "postgres-checkpoints.json"), &checkpoints)
				if checkpoints.Services != 2 || len(checkpoints.Results) != 2 {
					t.Fatal("checkpoint collection omitted an owned PostgreSQL generation")
				}
				generations, identities := map[string]bool{}, map[string]bool{}
				for _, result := range checkpoints.Results {
					if result.Status != "observed" || result.Generation == "" || result.Identity == nil || generations[result.Generation] || identities[*result.Identity] || !slices.Contains(files, filepath.Join(run, result.Scope)) {
						t.Fatalf("checkpoint service proof is incomplete or duplicated: %+v", result)
					}
					generations[result.Generation], identities[*result.Identity] = true, true
				}
			}

			operations := map[string]int{}
			ownedDatabases := 0
			for _, activity := range scope.Fixture.ByStrategy {
				operations[activity.Operation] += activity.Count
				if activity.Operation == "database-create" && activity.FixturePolicy == suiteservices.PostgresFixturePolicyTemplateClone {
					ownedDatabases += activity.Count
				}
			}
			if operations["database-reset"] != 0 {
				t.Fatal("transaction/dedicated selection used destructive reset")
			}
			if selection.id == "postgres" {
				if operations["transaction"] != 2 {
					t.Fatal("rollback isolation assertions did not execute")
				}
				return
			}
			if selection.id == "object-store" {
				if scope.Postgres.CreatedDatabaseCount != 0 || scope.ObjectStore.BucketCreateCount == 0 || scope.ObjectStore.BucketCleanupCount < scope.ObjectStore.BucketCreateCount {
					t.Fatal("object-only selection did not preserve namespace isolation and cleanup")
				}
				return
			}
			// These bounded rows exercise distinct real Go consumers. The retained
			// preparation examples must cover the complete small fixture, not a
			// truncated sample masquerading as identity-isolation evidence.
			preparations := scope.Postgres.DatabasePreparations
			if len(preparations) != scope.Postgres.CreatedDatabaseCount {
				t.Fatal("incomplete database identity evidence")
			}
			identities, processes := map[string]bool{}, map[int]bool{}
			for _, preparation := range preparations {
				if preparation.Name == "" || identities[preparation.Name] {
					t.Fatal("database identity was omitted or shared")
				}
				identities[preparation.Name], processes[preparation.PID] = true, true
			}
			if len(processes) < 2 {
				t.Fatal("independent database consumers missing")
			}
			if ownedDatabases == 0 || operations["database-drop"] != ownedDatabases || operations["transaction"] != 2 {
				t.Fatal("owned database cleanup or transaction isolation evidence missing")
			}
			if scope.ObjectStore.BucketCreateCount == 0 || scope.ObjectStore.BucketCleanupCount < scope.ObjectStore.BucketCreateCount {
				t.Fatal("owned object namespaces were not cleaned")
			}
			t.Logf("selected services, consumer isolation and cleanup passed: %s", run)
		})
	}
}

func readSelectedBoundaryJSON(t *testing.T, name string, output any) {
	t.Helper()
	data, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, output); err != nil {
		t.Fatal(err)
	}
}
