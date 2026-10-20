//go:build linux

package harnessfixtures

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/harnesslocal"
	"github.com/JochiRaider/cartulary/internal/testutil/harnesssession"
	"github.com/JochiRaider/cartulary/internal/testutil/pgschema"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func writeFixtureJSON(t *testing.T, name string, v any) {
	t.Helper()
	directory, err := harnesslocal.PrivateDirectory(filepath.Dir(name))
	if err != nil {
		t.Fatal(err)
	}
	directory.Close()
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	if err := harnesslocal.PublishNew(name, data); err != nil {
		t.Fatal(err)
	}
}
func serviceScope(run, target, suite string) map[string]any {
	startup := map[string]any{"attempt_count": 1, "retry_count": 0, "slowest_attempt_duration_ms": 1, "final_attempt": 1, "final_retryable": false, "final_retry_blocked_by_context": false}
	return map[string]any{"schema_id": "cartulary.test_services.scope.v2", "run_id": run, "target": target, "suite_id": suite, "artifact_dir": "_shared/test-services/" + suite, "readiness_generation": "sha256:" + strings.Repeat("a", 64),
		"wrapper": map[string]int{"owned_count": 1, "pass_through_count": 0}, "preflight": map[string]any{"docker_ok": true, "reaper_ready": true, "stale_containers_scanned": 0, "stale_containers_removed": 0, "stale_containers_deferred": 0, "ryuk_disabled_for_suite_startup": false}, "failures": map[string]any{}, "cleanup": map[string]any{},
		"postgres":     map[string]any{"started": true, "startup": startup, "attached_harness_count": 0, "created_database_count": 0, "migrated_database_count": 0, "template_clone_count": 0},
		"object_store": map[string]any{"started": true, "secure": true, "startup": startup, "attached_harness_count": 0, "bucket_create_count": 0, "bucket_cleanup_count": 0},
		"browser_e2e":  map[string]int{"retired_fixture_count": 0, "cleaned_fixture_count": 0, "reclaimed_fixture_count": 0}, "fixture": map[string]int{"total_count": 0, "total_duration_ms": 0, "strategy_aggregate_count": 0}, "started_services": map[string]any{}}
}
func TestFixtureSuiteOwner(t *testing.T) {
	t.Run("postgres_only_builder_can_bind_independent_objects_without_reprovisioning", func(t *testing.T) {
		runtime, root, run := runtimeFixture(t)
		p := NewProduction(ProductionOptions{Runtime: runtime, Root: root, RunRoot: run}, Options{})
		makeSuite := func(service string) *Suite {
			file := filepath.Join(runtime.Root, service, "suite-lease.json")
			var provisioning any
			environment := map[string]string{}
			if service == "postgres" {
				provisioning = map[string]any{"schema_id": "cartulary.test_services.postgres_provisioning.v2", "postgres_isolation": "exclusive", "allocation_ref": "measurement-fixture", "generation": "12345", "policy_digest": strings.Repeat("a", 64), "template_digest": strings.Repeat("b", 64)}
				environment["CARTULARY_PGTEST_GENERATION"] = "12345"
			} else {
				environment["CARTULARY_S3TEST_ENDPOINT"] = "https://objects.invalid"
			}
			writeFixtureJSON(t, file, map[string]any{"schema_id": "cartulary.test_services.lease.v3", "postgres_provisioning": provisioning, "lease_id": runtime.LeaseID, "suite_id": service, "run_id": runtime.RunID, "result_root": filepath.Dir(run), "run_root": run, "target": "test", "mode": "owned", "ownership_mode": "owned", "owner_pid": os.Getpid(), "created_at": time.Now().UTC().Format(time.RFC3339Nano), "resources": []any{map[string]any{"kind": "container", "service": service, "container_id": service, "image": "fixture", "labels": map[string]any{}}}, "proof_labels": map[string]any{}, "proof_prefixes": map[string]any{}, "cleanup_state": "not_started"})
			return &Suite{LeaseFile: file, ResourceRef: service, ResourceRefs: []string{service}, Environment: environment}
		}
		pg, objects := makeSuite("postgres"), makeSuite("object_store")
		request := Request{PostgresIsolation: "exclusive", Services: []string{"postgres"}, Capability: "postgres_dedicated", UnitID: "builder", RuntimeProfile: "default", FixtureProfile: "measurement", SnapshotKey: "snapshot", BuilderUnit: "builder"}
		pgStarts, objectStarts := 0, 0
		p.clusters[clusterKey(request)] = &SuiteOwner{Acquire: func(context.Context) (*Suite, error) { pgStarts++; return pg, nil }}
		p.Suite = &SuiteOwner{Acquire: func(context.Context) (*Suite, error) { objectStarts++; return objects, nil }}
		first, err := p.services(t.Context(), request)
		if err != nil || first != pg || objectStarts != 0 {
			t.Fatal("builder acquired unrelated service", err)
		}
		request.UnitID = "browser"
		request.RuntimeProfile = "network_flow_claimed"
		request.Capability = "browser_stack"
		request.Services = []string{"object_store", "postgres"}
		bound, err := p.services(t.Context(), request)
		if err != nil {
			t.Fatal(err)
		}
		again, err := p.services(t.Context(), request)
		if err != nil || again != bound || pgStarts != 1 || objectStarts != 1 {
			t.Fatal("binding reprovisioned services", err)
		}
		if bound.Environment["CARTULARY_PGTEST_GENERATION"] != "12345" || bound.Environment["CARTULARY_S3TEST_ENDPOINT"] == "" || len(bound.ResourceRefs) != 2 {
			t.Fatal("incomplete physical binding")
		}
		if pg.Environment["CARTULARY_S3TEST_ENDPOINT"] != "" || pg.LeaseFile == bound.LeaseFile {
			t.Fatal("physical owner mutated by a consumer")
		}
		proof, err := readObject(bound.LeaseFile, "cartulary.test_services.lease.v3", 1048576)
		if err != nil || proof["ownership_mode"] != "attach" || len(proof["resources"].([]any)) != 2 {
			t.Fatal("consumer acquired physical termination authority", err)
		}
		other := request
		other.BuilderUnit = "other-builder"
		other.Services = []string{"postgres"}
		otherPG := &Suite{Environment: map[string]string{"CARTULARY_PGTEST_GENERATION": "67890"}}
		p.clusters[clusterKey(other)] = &SuiteOwner{Acquire: func(context.Context) (*Suite, error) { return otherPG, nil }}
		independent, err := p.services(t.Context(), other)
		if err != nil || independent != otherPG || independent == bound {
			t.Fatal("distinct builder borrowed another snapshot's cluster", err)
		}
		if same, err := p.services(t.Context(), request); err != nil || same != bound {
			t.Fatal("distinct builder displaced original snapshot ownership", err)
		}
		for _, invalid := range []Request{
			{PostgresIsolation: "exclusive", Services: []string{"postgres"}, FixtureProfile: "measurement", SnapshotKey: "snapshot"},
			{PostgresIsolation: "exclusive", Services: []string{"postgres"}, FixtureProfile: "measurement", BuilderUnit: "builder"},
			{PostgresIsolation: "exclusive", Services: []string{"postgres"}, SnapshotKey: "snapshot", BuilderUnit: "builder"},
			{PostgresIsolation: "shared", Services: []string{"postgres"}, FixtureProfile: "measurement", SnapshotKey: "snapshot", BuilderUnit: "builder"},
		} {
			if _, err := p.services(t.Context(), invalid); err == nil {
				t.Fatal("incomplete or shared snapshot request provisioned services", invalid)
			}
		}
	})

	t.Run("explicit_cluster_scope_and_lifetime", func(t *testing.T) {
		runtime, root, run := runtimeFixture(t)
		p := NewProduction(ProductionOptions{Runtime: runtime, Root: root, RunRoot: run}, Options{})
		shared := Request{PostgresIsolation: "shared", Services: []string{"postgres"}, Capability: "postgres_transaction", UnitID: "ordinary-a"}
		second := shared
		second.UnitID = "ordinary-b"
		exclusive := shared
		exclusive.PostgresIsolation = "exclusive"
		exclusive.UnitID = "isolated"
		if clusterKey(shared) != clusterKey(second) || clusterKey(shared) == clusterKey(exclusive) {
			t.Fatal("database capability replaced cluster scope")
		}
		closed := 0
		p.clusters[clusterKey(shared)] = &SuiteOwner{Acquire: func(context.Context) (*Suite, error) {
			return &Suite{Environment: map[string]string{"generation": "shared"}}, nil
		}}
		p.clusters[clusterKey(exclusive)] = &SuiteOwner{Acquire: func(context.Context) (*Suite, error) {
			return &Suite{Environment: map[string]string{"generation": "exclusive"}, Close: func(context.Context) error { closed++; return errors.New("uncertain physical settlement") }}, nil
		}}
		a, err := p.services(t.Context(), shared)
		if err != nil {
			t.Fatal(err)
		}
		b, err := p.services(t.Context(), second)
		if err != nil || a != b {
			t.Fatal("ordinary users lost physical sharing", err)
		}
		c, err := p.services(t.Context(), exclusive)
		if err != nil || c == a {
			t.Fatal("exclusive request borrowed shared cluster", err)
		}
		if p.releaseCluster(t.Context(), exclusive) == nil || closed != 1 {
			t.Fatal("uncertain cleanup accepted")
		}
		if _, err := p.services(t.Context(), exclusive); err == nil {
			t.Fatal("quarantined cluster reused")
		}
		p.Options.Run = func(context.Context, string, []string, map[string]string) (CommandResult, error) {
			t.Fatal("invalid scope reached physical provisioning")
			return CommandResult{}, errors.New("unexpected provisioning")
		}
		for _, request := range []Request{{Services: []string{"postgres"}}, {PostgresIsolation: "none", Services: []string{"postgres"}}, {PostgresIsolation: "shared"}, {PostgresIsolation: "shared", Services: []string{"postgres"}, Capability: "browser_stack"}, {PostgresIsolation: "shared", Services: []string{"postgres"}, Capability: "postgres_dedicated"}, {PostgresIsolation: "shared", Services: []string{"postgres"}, Capability: "postgres_migration"}} {
			_, err := p.services(t.Context(), request)
			var failure *Failure
			if !errors.As(err, &failure) || failure.Reason != "fixture_error" {
				t.Fatal("contradictory or missing scope was not rejected", request, err)
			}
		}
	})

	t.Run("borrowed_suite_proof_and_detach_never_terminate", func(t *testing.T) {
		runtime, root, run := runtimeFixture(t)
		detached := 0
		source := map[string]any{"schema_id": "cartulary.test_services.lease.v3", "postgres_provisioning": nil, "lease_id": "foreign-lease", "suite_id": "persistent-suite", "run_id": "persistent-run", "result_root": "/private/persistent", "run_root": "/private/persistent/persistent-run", "target": "test-services-start", "mode": "owned", "ownership_mode": "owned", "owner_pid": os.Getpid(), "created_at": time.Now().UTC().Format(time.RFC3339Nano), "resources": []any{}, "proof_labels": map[string]any{}, "proof_prefixes": map[string]any{}, "cleanup_state": "not_started"}
		p := ProductionOptions{Root: root, RunRoot: run, Target: "browser-e2e", Runtime: runtime, Environment: map[string]string{"NODE_BIN": "node"}}
		p.Run = func(_ context.Context, _ string, args []string, _ map[string]string) (CommandResult, error) {
			if len(args) != 2 {
				t.Fatalf("unexpected physical operation %v", args)
			}
			if args[1] == "detach" {
				detached++
				return CommandResult{Started: true, Stdout: []byte(`{}`)}, nil
			}
			if args[1] != "attach" {
				t.Fatal(args)
			}
			writeFixtureJSON(t, filepath.Join(runtime.Root, "borrowed-suite.json"), map[string]any{"private": "binding"})
			data, _ := json.Marshal(map[string]any{"environment": map[string]string{"CARTULARY_TEST_SUITE_ID": "run-suite"}, "container_proof": map[string]any{"lease": source, "ownership": "borrowed"}})
			return CommandResult{Started: true, Stdout: data}, nil
		}
		suite, err := p.AttachSuite(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		lease, err := readObject(suite.LeaseFile, "cartulary.test_services.lease.v3", 1048576)
		if err != nil {
			t.Fatal(err)
		}
		if lease["mode"] != "attach" || lease["ownership_mode"] != "attach" || lease["run_id"] != runtime.RunID || lease["suite_id"] != "run-suite" || source["mode"] != "owned" {
			t.Fatal("borrower proof lost identity or mutated owner", lease)
		}
		if err := suite.Close(context.Background()); err != nil {
			t.Fatal(err)
		}
		if err := suite.Close(context.Background()); err != nil || detached != 1 {
			t.Fatal(err, detached)
		}
		resources, err := runtime.Resources(64)
		if err != nil || len(resources) != 0 {
			t.Fatal(resources, err)
		}
	})

	t.Run("concurrent_single_acquisition_and_close", func(t *testing.T) {
		var acquired, closed, residencyReleased atomic.Int32
		owner := &SuiteOwner{Acquire: func(context.Context) (*Suite, error) {
			acquired.Add(1)
			return &Suite{Close: func(context.Context) error { closed.Add(1); return nil }}, nil
		}, ReleaseResidency: func(context.Context) error {
			if closed.Load() != 1 {
				t.Error("resident capacity released before physical settlement")
			}
			residencyReleased.Add(1)
			return nil
		}}
		var wg sync.WaitGroup
		for range 12 {
			wg.Go(func() {
				if _, err := owner.Ensure(context.Background()); err != nil {
					t.Error(err)
				}
			})
		}
		wg.Wait()
		for range 12 {
			wg.Go(func() {
				if err := owner.Close(context.Background()); err != nil {
					t.Error(err)
				}
			})
		}
		wg.Wait()
		if acquired.Load() != 1 || closed.Load() != 1 || residencyReleased.Load() != 1 {
			t.Fatal(acquired.Load(), closed.Load(), residencyReleased.Load())
		}
		if _, err := owner.Ensure(context.Background()); err == nil {
			t.Fatal("acquire after close")
		}
	})
	for _, failure := range []string{"physical", "residency"} {
		t.Run("uncertain_cluster_settlement_"+failure, func(t *testing.T) {
			uncertain := errors.New("uncertain settlement")
			physical, residency := 0, 0
			owner := &SuiteOwner{Acquire: func(context.Context) (*Suite, error) {
				return &Suite{Close: func(context.Context) error {
					physical++
					if failure == "physical" {
						return uncertain
					}
					return nil
				}}, nil
			}, ReleaseResidency: func(context.Context) error { residency++; return uncertain }}
			if _, err := owner.Ensure(t.Context()); err != nil {
				t.Fatal(err)
			}
			for range 2 {
				if err := owner.Close(t.Context()); !errors.Is(err, uncertain) {
					t.Fatal("uncertainty was hidden or retried", err)
				}
			}
			if physical != 1 || (failure == "physical" && residency != 0) || (failure == "residency" && residency != 1) {
				t.Fatal("physical/resident settlement order or attempt count", physical, residency)
			}
			if _, err := owner.Ensure(t.Context()); err == nil {
				t.Fatal("uncertain generation was reused")
			}
		})
	}
	t.Run("unpublished_started_owner_retains_residency", func(t *testing.T) {
		runtime, root, run := runtimeFixture(t)
		starts := 0
		options := ProductionOptions{Runtime: runtime, Root: root, RunRoot: run, Environment: map[string]string{"TEST_SERVICES_BIN": "/fixed/testservices"}, Run: func(context.Context, string, []string, map[string]string) (CommandResult, error) {
			starts++
			return CommandResult{Started: true, Exit: 1}, errors.New("interrupted before publication")
		}}
		owner := &SuiteOwner{Acquire: options.StartSuite, ReleaseResidency: func(context.Context) error {
			t.Fatal("unpublished physical owner released residency")
			return nil
		}}
		if _, err := owner.Ensure(t.Context()); !errors.Is(err, harnesssession.ErrUnresolved) {
			t.Fatal("missing settlement proof accepted", err)
		}
		for range 2 {
			if err := owner.Close(t.Context()); !errors.Is(err, harnesssession.ErrUnresolved) {
				t.Fatal("unpublished owner settled", err)
			}
		}
		resources, err := runtime.Resources(64)
		if err != nil || len(resources) == 0 || starts != 1 {
			t.Fatal("intent lost or startup retried", len(resources), starts, err)
		}
		for _, resource := range resources {
			if err := runtime.RecordResource(resource.Kind, resource.Target, "released"); err != nil {
				t.Fatal(err)
			}
		}
	})
	for _, variant := range []string{"ready", "recovery_proof", "missing_recovery_file", "foreign_recovery_file", "unsafe_recovery_file", "forged_scope", "duplicate_environment", "startup_failure", "startup_settled", "startup_cleanup_failed", "missing_result", "missing_result_cleanup_failure"} {
		t.Run(variant, func(t *testing.T) {
			runtime, root, run := runtimeFixture(t)
			stops := 0
			uncertain := errors.New("synthetic settlement uncertainty")
			credentialFiles := []string{}
			options := ProductionOptions{PostgresIsolation: "exclusive", PostgresAllocation: "fixture-owner", Root: root, RunRoot: run, Target: "test", Runtime: runtime, Environment: map[string]string{"TEST_SERVICES_BIN": "/fixed/testservices"}}
			options.Run = func(_ context.Context, executable string, args []string, environment map[string]string) (CommandResult, error) {
				if executable != "/fixed/testservices" {
					t.Fatal(executable)
				}
				if args[0] == "terminate-suite" {
					stops++
					if variant == "missing_result_cleanup_failure" {
						return CommandResult{Started: true, Exit: 1}, uncertain
					}
					if !strings.Contains(args[2], ".cleanup-") {
						t.Fatal("original proof passed to destructive helper")
					}
					if _, err := harnesslocal.ReadFile(args[2], 1048576, false); err != nil {
						t.Fatal(err)
					}
					return CommandResult{Started: true}, nil
				}
				suite := environment["CARTULARY_TEST_SUITE_ID"]
				scope := serviceScope(runtime.RunID, "test", suite)
				status := map[string]any{"schema_id": "cartulary.test_services.start_result.v2", "recovery_files": []string{}, "status": "ready", "run_id": runtime.RunID, "target": "test", "suite_id": suite, "service_scope_ref": "_shared/test-services/" + suite + "/service-scope.json", "failure_class": nil, "failure_reason": nil}
				writeFixtureJSON(t, args[4], map[string]any{"schema_id": "cartulary.test_services.lease.v3", "postgres_provisioning": map[string]any{"schema_id": "cartulary.test_services.postgres_provisioning.v2", "postgres_isolation": "exclusive", "allocation_ref": "fixture-owner", "generation": "12345", "policy_digest": pgtest.ProvisioningDigest(), "template_digest": pgtest.TemplateDigest(pgschema.MustHash())}, "lease_id": runtime.LeaseID, "suite_id": suite, "run_id": runtime.RunID, "result_root": filepath.Dir(run), "run_root": run, "target": "test", "mode": "owned", "ownership_mode": "owned", "owner_pid": os.Getpid(), "created_at": time.Now().UTC().Format(time.RFC3339Nano), "resources": []any{map[string]any{"kind": "container", "service": "postgres", "container_id": "fixture-postgres", "image": "fixture", "labels": map[string]any{}}}, "proof_labels": map[string]any{}, "proof_prefixes": map[string]any{}, "cleanup_state": "not_started"})
				if strings.HasPrefix(variant, "missing_result") {
					return CommandResult{Started: true, Exit: 1}, errors.New("worker death")
				}
				if variant == "forged_scope" {
					scope["run_id"] = "foreign"
				}
				if strings.HasPrefix(variant, "startup_") {
					status["status"] = "failed"
					status["failure_class"] = "infra"
					status["failure_reason"] = "service_start_error"
					scope["failure"] = map[string]any{"failure_class": "infra", "failure_reason": "service_start_error", "attempts_started": 1, "max_attempts": 1, "retryable": false, "retry_blocked_by_context": false}
					if variant == "startup_settled" {
						scope["cleanup"] = map[string]any{"status": "startup_failed", "physical_settlement": "completed"}
					} else if variant == "startup_cleanup_failed" {
						scope["cleanup"] = map[string]any{"status": "startup_failed", "physical_settlement": "unresolved"}
					}
				}
				writeFixtureJSON(t, filepath.Join(run, status["service_scope_ref"].(string)), scope)
				if variant == "duplicate_environment" {
					if err := harnesslocal.PublishNew(args[2], []byte(`{"SSL_CERT_FILE":"a","SSL_CERT_FILE":"b"}`)); err != nil {
						t.Fatal(err)
					}
				} else if variant == "recovery_proof" {
					for _, name := range []string{"root.pem", "admin.pem", "admin.key"} {
						file := filepath.Join(runtime.Root, "test-services", "postgres-tls", name)
						writeFixtureJSON(t, file, map[string]string{"fixture": name})
						credentialFiles = append(credentialFiles, file)
					}
					writeFixtureJSON(t, filepath.Join(runtime.Root, "unrelated-secret.json"), map[string]string{"unneeded": "secret"})
					writeFixtureJSON(t, args[2], map[string]string{"CARTULARY_PGTEST_ISOLATION": "exclusive", "CARTULARY_PGTEST_ALLOCATION_REF": "fixture-owner", "CARTULARY_PGTEST_GENERATION": "12345", "CARTULARY_PGTEST_PROVISIONING_DIGEST": pgtest.ProvisioningDigest(), "CARTULARY_PGTEST_TEMPLATE_DIGEST": pgtest.TemplateDigest(pgschema.MustHash()), "CARTULARY_PGTEST_ADMIN_DSN": "postgres://fixture@localhost/postgres?sslrootcert=" + credentialFiles[0] + "&sslcert=" + credentialFiles[1] + "&sslkey=" + credentialFiles[2]})
				} else {
					writeFixtureJSON(t, args[2], map[string]string{"CARTULARY_PGTEST_ISOLATION": "exclusive", "CARTULARY_PGTEST_ALLOCATION_REF": "fixture-owner", "CARTULARY_PGTEST_GENERATION": "12345", "CARTULARY_PGTEST_PROVISIONING_DIGEST": pgtest.ProvisioningDigest(), "CARTULARY_PGTEST_TEMPLATE_DIGEST": pgtest.TemplateDigest(pgschema.MustHash()), "CARTULARY_PGTEST_ADMIN_DSN": "private-password", "IGNORE_THIS": "foreign"})
				}
				if variant == "recovery_proof" {
					status["recovery_files"] = credentialFiles
				}
				if variant == "missing_recovery_file" {
					status["recovery_files"] = []string{filepath.Join(runtime.Root, "missing.pem")}
				}
				if variant == "foreign_recovery_file" {
					file := filepath.Join(t.TempDir(), "foreign.pem")
					if err := os.WriteFile(file, []byte("foreign proof"), 0600); err != nil {
						t.Fatal(err)
					}
					status["recovery_files"] = []string{file}
				}
				if variant == "unsafe_recovery_file" {
					file := filepath.Join(runtime.Root, "unsafe.pem")
					writeFixtureJSON(t, file, map[string]string{"unsafe": "proof"})
					if err := os.Chmod(file, 0644); err != nil {
						t.Fatal(err)
					}
					status["recovery_files"] = []string{file}
				}
				writeFixtureJSON(t, args[6], status)
				if strings.HasPrefix(variant, "startup_") {
					return CommandResult{Started: true, Exit: 1}, errors.New("startup")
				}
				return CommandResult{Started: true}, nil
			}
			suite, err := options.StartSuite(context.Background())
			if variant == "startup_settled" || variant == "startup_cleanup_failed" {
				if err == nil || stops != 0 {
					t.Fatal("startup failure concealed or cleanup repeated", stops, err)
				}
				resources, resourceErr := runtime.Resources(64)
				if resourceErr != nil {
					t.Fatal(resourceErr)
				}
				if variant == "startup_settled" {
					if suite != nil || len(resources) != 0 {
						t.Fatal("settlement receipt ignored")
					}
				} else {
					if suite == nil || len(resources) == 0 {
						t.Fatal("partial ownership lost")
					}
					for range 2 {
						if suite.Close(t.Context()) == nil || stops != 0 {
							t.Fatal("unresolved startup cleanup retried")
						}
					}
					for _, resource := range resources {
						if err := runtime.RecordResource(resource.Kind, resource.Target, "released"); err != nil {
							t.Fatal(err)
						}
					}
				}
				return
			}
			if variant == "missing_result_cleanup_failure" {
				if err == nil || suite == nil || !errors.Is(err, uncertain) {
					t.Fatal("partial acquisition lost unresolved ownership", err)
				}
				owner := &SuiteOwner{suite: suite, attempted: true, ReleaseResidency: func(context.Context) error {
					t.Fatal("unsettled generation released residency")
					return nil
				}}
				for range 2 {
					if err := owner.Close(t.Context()); !errors.Is(err, uncertain) {
						t.Fatal("unsettled generation became reusable", err)
					}
				}
				resources, err := runtime.Resources(64)
				if err != nil || len(resources) == 0 || stops != 1 {
					t.Fatal("recovery proof lost or cleanup retried", len(resources), stops, err)
				}
				// These are synthetic fault resources, not physical services.
				for _, resource := range resources {
					if err := runtime.RecordResource(resource.Kind, resource.Target, "released"); err != nil {
						t.Fatal(err)
					}
				}
				return
			}
			if variant == "ready" || variant == "recovery_proof" {
				if err != nil {
					t.Fatal(err)
				}
				defer suite.Close(context.Background())
				if variant == "recovery_proof" {
					if err := runtime.PreserveRecovery(); err != nil {
						t.Fatal(err)
					}
					for _, file := range credentialFiles {
						if _, err := os.Stat(file); err != nil {
							t.Fatalf("cleanup credential was purged before its resource settled: %s: %v", filepath.Base(file), err)
						}
					}
					if _, err := os.Stat(filepath.Join(runtime.Root, "unrelated-secret.json")); !errors.Is(err, os.ErrNotExist) {
						t.Fatal("unrelated detail survived preservation", err)
					}
				}
				if suite.Environment["IGNORE_THIS"] != "" {
					t.Fatal("unadmitted binding")
				}
				if values, err := runtime.ForbiddenValues(); err != nil || len(values) != 1 {
					t.Fatal("secret not registered")
				}
				if err := suite.Close(context.Background()); err != nil {
					t.Fatal(err)
				}
				_ = suite.Close(context.Background())
			} else {
				if err == nil {
					t.Fatal("invalid publication accepted")
				}
				var failure *Failure
				if !errors.As(err, &failure) {
					t.Fatal(err)
				}
				expected := "artifact"
				if variant == "startup_failure" {
					expected = "infra"
				}
				if failure.Class != expected {
					t.Fatal(err)
				}
			}
			if stops != 1 {
				t.Fatal("physical cleanup count", stops)
			}
			resources, err := runtime.Resources(64)
			if err != nil || len(resources) != 0 {
				t.Fatal(resources, err)
			}
		})
	}
}
func TestFixtureCommandLifecycle(t *testing.T) {
	runtime, root, _ := runtimeFixture(t)
	run := commandRunner(root, runtime, []string{})
	for _, mode := range []string{"exit", "cancel", "overflow"} {
		t.Run(mode, func(t *testing.T) {
			script := filepath.Join(root, "helper")
			body := "#!/bin/sh\nexit 7\n"
			if mode == "cancel" {
				body = "#!/bin/sh\nsleep 30 &\nwait\n"
			}
			if mode == "overflow" {
				body = "#!/bin/sh\nhead -c 1048577 /dev/zero\n"
			}
			if err := os.WriteFile(script, []byte(body), 0700); err != nil {
				t.Fatal(err)
			}
			ctx, done := context.WithTimeout(context.Background(), time.Second)
			defer done()
			result, err := run(ctx, script, nil, map[string]string{"PATH": "/usr/bin:/bin"})
			if err == nil || !result.Started {
				t.Fatal(result, err)
			}
			if mode == "exit" && result.Exit != 7 {
				t.Fatal(result.Exit)
			}
			if mode == "cancel" && !errors.Is(err, context.DeadlineExceeded) {
				t.Fatal(err)
			}
			if mode == "overflow" && len(result.Stdout) != 1048576 {
				t.Fatal("unbounded output")
			}
			resources, err := runtime.Resources(64)
			if err != nil || len(resources) != 0 {
				t.Fatal(resources, err)
			}
		})
	}
}
