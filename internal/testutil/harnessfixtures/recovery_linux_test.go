//go:build linux

package harnessfixtures

import (
	"context"
	"encoding/json"
	"errors"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/harnesslocal"
	"github.com/JochiRaider/cartulary/internal/testutil/harnesssession"
	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices"
	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices/leaseproof"
)

func fixtureReference(t *testing.T, runtime *harnesssession.Runtime, kind string, target any) string {
	t.Helper()
	reference, err := runtime.ResourceReference(kind, target)
	if err != nil {
		t.Fatal(err)
	}
	return reference
}
func recordBrowserRecoveryFixture(t *testing.T, runtime *harnesssession.Runtime, browser, kind, dependency, binding string) error {
	t.Helper()
	environmentFile := filepath.Join(filepath.Dir(filepath.Dir(browser)), "stack-environment.json")
	writeFixtureJSON(t, environmentFile, map[string]string{})
	return runtime.RecordDependentResource("browser_stack", browser, "pending", []string{fixtureReference(t, runtime, kind, dependency)}, binding, environmentFile)
}
func TestFixtureRecovery(t *testing.T) {
	testPackageResourcePhysicalRecovery(t)
	testRecoveryEnvironment(t)
	testOwnedLeaseWithoutStartupEnvironment(t)
	testPendingStartupRecovery(t)
	testRecoveryCommand(t)
	testCleanupDependencies(t)
	for _, mode := range []string{"ordinary", "recovery"} {
		t.Run("diagnostic_dependency_blocks_suite_"+mode, func(t *testing.T) {
			runtime, root, run := runtimeFixture(t)
			suite := filepath.Join(runtime.Root, "test-services", "suite-lease.json")
			writeFixtureJSON(t, suite, map[string]string{"proof": "fixture"})
			writeFixtureJSON(t, filepath.Join(filepath.Dir(suite), "suite-environment.json"), map[string]string{})
			if err := runtime.RecordDependentResource("managed_suite", suite, "acquired", []string{}); err != nil {
				t.Fatal(err)
			}
			defer runtime.RecordResource("managed_suite", suite, "released")
			scope := map[string]any{"boot": "unresolved-boot", "start": "1", "token": strings.Repeat("d", 64), "uid": os.Getuid()}
			if err := runtime.RecordDependentResource("diagnostic_scope", scope, "acquired", []string{fixtureReference(t, runtime, "managed_suite", suite)}); err != nil {
				t.Fatal(err)
			}
			defer runtime.RecordResource("diagnostic_scope", scope, "released")
			stops := 0
			p := ProductionOptions{Root: root, RunRoot: run, Runtime: runtime, Environment: map[string]string{"TEST_SERVICES_BIN": "services"}, Run: func(context.Context, string, []string, map[string]string) (CommandResult, error) {
				stops++
				return CommandResult{}, nil
			}}
			var err error
			if mode == "recovery" {
				err = p.Recover(context.Background())
			} else {
				production := NewProduction(p, Options{})
				production.Suite.suite = &Suite{LeaseFile: suite, Close: func(context.Context) error { stops++; return nil }}
				err = production.Close(context.Background())
			}
			if err == nil || stops != 0 {
				t.Fatal("diagnostic consumer did not block its suite", mode, err, stops)
			}
		})
	}
	t.Run("failed_consumer_blocks_only_its_declared_suite", func(t *testing.T) {
		runtime, root, run := runtimeFixture(t)
		a, b := filepath.Join(runtime.Root, "suite-a", "suite-lease.json"), filepath.Join(runtime.Root, "suite-b", "suite-lease.json")
		browser := filepath.Join(runtime.Root, "browser", "owner.json")
		for _, suite := range []string{a, b} {
			writeFixtureJSON(t, suite, map[string]string{"proof": "synthetic"})
			writeFixtureJSON(t, filepath.Join(filepath.Dir(suite), "suite-environment.json"), map[string]string{})
			if err := runtime.RecordDependentResource("managed_suite", suite, "acquired", []string{}); err != nil {
				t.Fatal(err)
			}
			defer runtime.RecordResource("managed_suite", suite, "released")
		}
		if err := recordBrowserRecoveryFixture(t, runtime, browser, "managed_suite", a, a); err != nil {
			t.Fatal(err)
		}
		defer runtime.RecordResource("browser_stack", browser, "released")
		inspections, stops := 0, 0
		outcomes := []CleanupOutcome{}
		p := ProductionOptions{Root: root, RunRoot: run, Runtime: runtime, Environment: map[string]string{"NODE_BIN": "node", "TEST_SERVICES_BIN": "services"}, ReportCleanup: func(v CleanupOutcome) { outcomes = append(outcomes, v) }}
		p.Run = func(_ context.Context, _ string, args []string, _ map[string]string) (CommandResult, error) {
			if args[0] == "terminate-suite" {
				stops++
				if !strings.HasPrefix(args[2], b+".cleanup-") {
					t.Fatal("dependent suite stopped")
				}
				return CommandResult{}, nil
			}
			inspections++
			return CommandResult{}, errors.New("inspection failed")
		}
		if _, err := runtime.BeginRecoveryAttempt(); err != nil {
			t.Fatal(err)
		}
		defer runtime.FinishRecoveryAttempt()
		for range 2 {
			if err := p.Recover(context.Background()); !errors.Is(err, ErrCleanupBlocked) {
				t.Fatal("missing blocked outcome", err)
			}
		}
		if inspections != 1 || stops != 1 {
			t.Fatal("duplicate or starved release", inspections, stops)
		}
		counts := map[string]int{}
		for _, outcome := range outcomes {
			counts[outcome.Outcome]++
		}
		if counts["failed"] != 2 || counts["blocked"] != 2 || counts["completed"] != 1 {
			t.Fatal("lost cleanup outcome", outcomes)
		}
		if err := runtime.FinishRecoveryAttempt(); err != nil {
			t.Fatal(err)
		}
		if _, err := runtime.BeginRecoveryAttempt(); err != nil {
			t.Fatal(err)
		}
		if err := p.Recover(context.Background()); !errors.Is(err, ErrCleanupBlocked) || inspections != 2 || stops != 1 {
			t.Fatal("fresh attempt did not retry unresolved consumer", err, inspections, stops)
		}
	})
	t.Run("stalled_dependency_inspection_is_bounded_once", func(t *testing.T) {
		runtime, root, run := runtimeFixture(t)
		browser := filepath.Join(runtime.Root, "browser-stack-leases", "a", "acquisition", "owner.json")
		suite := filepath.Join(runtime.Root, "suite", "suite-lease.json")
		if err := runtime.RecordDependentResource("managed_suite", suite, "pending", []string{}); err != nil {
			t.Fatal(err)
		}
		defer runtime.RecordResource("managed_suite", suite, "released")
		if err := recordBrowserRecoveryFixture(t, runtime, browser, "managed_suite", suite, suite); err != nil {
			t.Fatal(err)
		}
		defer runtime.RecordResource("browser_stack", browser, "released")
		if _, err := runtime.BeginRecoveryAttempt(); err != nil {
			t.Fatal(err)
		}
		defer runtime.FinishRecoveryAttempt()
		synctest.Test(t, func(t *testing.T) {
			calls := 0
			p := ProductionOptions{Root: root, RunRoot: run, Target: "review", Runtime: runtime, Environment: map[string]string{"NODE_BIN": "node"}}
			p.Run = func(ctx context.Context, _ string, args []string, _ map[string]string) (CommandResult, error) {
				calls++
				if args[1] != "inspect" {
					t.Error("unexpected cleanup")
				}
				<-ctx.Done()
				return CommandResult{}, ctx.Err()
			}
			before := time.Now()
			if err := p.Recover(context.Background()); err == nil {
				t.Fatal("stalled inspection passed")
			}
			if time.Since(before) != 120*time.Second {
				t.Fatal("inspection did not use resource deadline")
			}
			if err := p.Recover(context.Background()); err == nil || calls != 1 {
				t.Fatal("duplicate resource budget", calls, err)
			}
		})
	})
	for _, variant := range []string{"success", "missing_dependency", "unknown_dependency", "failed_browser", "failed_suite", "borrowed", "borrowed_browser_failure"} {
		t.Run(variant, func(t *testing.T) {
			runtime, root, run := runtimeFixture(t)
			suite := filepath.Join(runtime.Root, "test-services", "suite-lease.json")
			browser := filepath.Join(runtime.Root, "browser-stack-leases", "a", "acquisition", "owner.json")
			writeFixtureJSON(t, suite, map[string]string{"proof": "synthetic"})
			writeFixtureJSON(t, filepath.Join(filepath.Dir(suite), "suite-environment.json"), map[string]string{})
			writeFixtureJSON(t, filepath.Join(runtime.Root, "borrowed-environment.json"), map[string]string{})
			writeFixtureJSON(t, browser, map[string]string{"proof": "synthetic"})
			if variant != "missing_dependency" && !strings.HasPrefix(variant, "borrowed") {
				if err := runtime.RecordDependentResource("managed_suite", suite, "acquired", []string{}); err != nil {
					t.Fatal(err)
				}
			}
			if strings.HasPrefix(variant, "borrowed") {
				file := filepath.Join(runtime.Root, "borrowed-suite.json")
				writeFixtureJSON(t, file, map[string]string{"proof": "synthetic"})
				if err := runtime.RecordDependentResource("borrowed_suite", file, "acquired", []string{}); err != nil {
					t.Fatal(err)
				}
			}
			if variant != "borrowed" {
				kind, dependency := "managed_suite", suite
				if strings.HasPrefix(variant, "borrowed") {
					kind, dependency = "borrowed_suite", filepath.Join(runtime.Root, "borrowed-suite.json")
				}
				binding := suite
				if kind == "borrowed_suite" {
					binding = filepath.Join(runtime.Root, "borrowed-service-lease.json")
				}
				err := recordBrowserRecoveryFixture(t, runtime, browser, kind, dependency, binding)
				if variant == "missing_dependency" {
					if err == nil {
						t.Fatal("unresolved dependency admitted")
					}
					return
				}
				if err != nil {
					t.Fatal(err)
				}
			}
			t.Cleanup(func() {
				resources, err := runtime.Resources(64)
				if err != nil {
					t.Error(err)
					return
				}
				for _, r := range resources {
					var target any
					json.Unmarshal(r.Target, &target)
					runtime.RecordResource(r.Kind, target, "released")
				}
			})
			calls := []string{}
			p := ProductionOptions{Root: root, RunRoot: run, Target: "review", Runtime: runtime, Environment: map[string]string{"NODE_BIN": "node", "TEST_SERVICES_BIN": "services"}}
			p.Run = func(_ context.Context, executable string, args []string, _ map[string]string) (CommandResult, error) {
				if args[0] == "terminate-suite" {
					calls = append(calls, "suite")
					if _, err := harnesslocal.ReadFile(args[2], 1048576, false); err != nil {
						t.Fatal(err)
					}
					if variant == "failed_suite" {
						return CommandResult{Started: true, Exit: 12}, errors.New("physical parent unresolved")
					}
					return CommandResult{Started: true}, nil
				}
				op := args[1]
				calls = append(calls, op)
				if op == "inspect" {
					dependency := suite
					if variant == "borrowed_browser_failure" {
						dependency = filepath.Join(runtime.Root, "borrowed-service-lease.json")
					}
					if variant == "unknown_dependency" {
						dependency = filepath.Join(runtime.Root, "foreign", "suite-lease.json")
					}
					data, _ := json.Marshal(map[string]string{"suite_lease": dependency})
					return CommandResult{Started: true, Stdout: data}, nil
				}
				if op == "recover" && (variant == "failed_browser" || variant == "borrowed_browser_failure") {
					if variant == "borrowed_browser_failure" && len(args) != 3 {
						t.Fatal("borrowed service granted destruction authority")
					}
					return CommandResult{Started: true, Exit: 12}, errors.New("physical stop failed")
				}
				if op == "recover" && (len(args) != 4 || args[3] != suite) {
					t.Fatal("missing exact owned parent proof")
				}
				return CommandResult{Started: true, Stdout: []byte(`{}`)}, nil
			}
			err := p.Recover(context.Background())
			joined := strings.Join(calls, ",")
			resources, readErr := runtime.Resources(64)
			if readErr != nil {
				t.Fatal(readErr)
			}
			if variant == "success" && len(resources) != 0 {
				t.Fatal("successful recovery left resource authority", resources)
			}
			if variant == "failed_browser" && len(resources) != 2 {
				t.Fatal("failed browser recovery lost exact browser and suite authority", resources)
			}
			if variant == "success" {
				if err != nil || joined != "inspect,recover,suite" {
					t.Fatal(calls, err)
				}
			} else if variant == "failed_suite" {
				if err == nil || joined != "inspect,recover,suite" || len(resources) != 1 || resources[0].Kind != "managed_suite" {
					t.Fatal("logical browser release concealed unresolved parent", calls, err, resources)
				}
			} else if variant == "borrowed" {
				if err != nil || joined != "detach" {
					t.Fatal(calls, err)
				}
			} else {
				if err == nil || strings.Contains(joined, "suite") || strings.Contains(joined, "detach") {
					t.Fatal("dependency failure authorized suite stop", calls, err)
				}
			}
		})
	}
}

func testPackageResourcePhysicalRecovery(t *testing.T) {
	for _, mode := range []string{"ordinary", "recovery"} {
		t.Run("package_resources_independent_physical_cleanup_"+mode, func(t *testing.T) {
			runtime, root, run := runtimeFixture(t)
			log, worker := filepath.Join(t.TempDir(), "calls"), filepath.Join(t.TempDir(), "physical-worker")
			script := "#!/bin/sh\ninput=$(/bin/cat)\nprintf '%s\\n' \"$input\" >> \"$PACKAGE_FIXTURE_LOG\"\ncase \"$input\" in *'\"role\":\"application\"'*) exit 12;; esac\n"
			if err := os.WriteFile(worker, []byte(script), 0700); err != nil {
				t.Fatal(err)
			}
			for _, role := range []string{"application", "postgres", "seaweedfs", "build_tag"} {
				proof := map[string]any{"daemon_id": "fixture-engine", "token": "11111111-2222-4333-8444-555555555555", "role": role}
				if err := runtime.RecordDependentResource("package_resource", proof, "pending", []string{}); err != nil {
					t.Fatal(err)
				}
				defer runtime.RecordResource("package_resource", proof, "released")
			}
			p := ProductionOptions{Root: root, RunRoot: run, Runtime: runtime, Environment: map[string]string{"NODE_BIN": worker, "PACKAGE_FIXTURE_LOG": log}}
			var err error
			if mode == "recovery" {
				err = p.Recover(context.Background())
			} else {
				err = p.closeResources(context.Background(), "", nil)
			}
			if err == nil {
				t.Fatal("failed package resource was accepted")
			}
			data, readErr := os.ReadFile(log)
			if readErr != nil {
				t.Fatal(readErr)
			}
			roles := map[string]bool{}
			for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
				var resource harnesssession.Resource
				if err := json.Unmarshal([]byte(line), &resource); err != nil {
					t.Fatal(err)
				}
				var proof struct {
					Role string `json:"role"`
				}
				if err := json.Unmarshal(resource.Target, &proof); err != nil {
					t.Fatal(err)
				}
				if roles[proof.Role] || resource.Kind != "package_resource" || resource.RunID != runtime.RunID || resource.LeaseID != runtime.LeaseID {
					t.Fatal("incorrect physical ownership", resource)
				}
				roles[proof.Role] = true
			}
			if len(roles) != 4 {
				t.Fatal("independent package cleanup skipped", roles)
			}
			remaining, err := runtime.Resources(4096)
			if err != nil || len(remaining) != 1 || !strings.Contains(string(remaining[0].Target), `"role":"application"`) {
				t.Fatal("failed proof lost or successful proof retained", remaining, err)
			}
		})
	}
}

func testRecoveryEnvironment(t *testing.T) {
	t.Run("recovery_environment_round_trip", func(t *testing.T) {
		root := t.TempDir()
		if err := os.Chmod(root, 0700); err != nil {
			t.Fatal(err)
		}
		file := filepath.Join(root, "environment.json")
		input := map[string]string{
			"SSL_CERT_FILE": "root.pem", "CARTULARY_TEST_SUITE_ID": "suite",
			"CARTULARY_PGTEST_ADMIN_DSN":              "private-database-binding",
			"CARTULARY_S3TEST_ENDPOINT":               "private-object-binding",
			"CARTULARY_S3TEST_PROXY_CERTIFICATE_PATH": "excluded.pem",
			"CARTULARY_S3TEST_PROXY_PRIVATE_KEY_PATH": "excluded.key",
			"UNRELATED_SECRET":                        "excluded-secret",
		}
		if err := saveRecoveryEnvironment(file, input); err != nil {
			t.Fatal(err)
		}
		got, err := readRecoveryEnvironment(file)
		want := map[string]string{
			"SSL_CERT_FILE": "root.pem", "CARTULARY_TEST_SUITE_ID": "suite",
			"CARTULARY_PGTEST_ADMIN_DSN": "private-database-binding",
			"CARTULARY_S3TEST_ENDPOINT":  "private-object-binding",
		}
		if err != nil || !maps.Equal(got, want) {
			t.Fatal("recovery environment lost its exact allowlisted bindings", err)
		}
		if err := saveRecoveryEnvironment(file, map[string]string{}); err != nil {
			t.Fatal(err)
		}
		if got, err := readRecoveryEnvironment(file); err != nil || got == nil || len(got) != 0 {
			t.Fatal("empty prelaunch environment rejected", err)
		}
	})
	for name, data := range map[string]string{
		"duplicate":        `{"CARTULARY_TEST_SUITE_ID":"foreign","CARTULARY_TEST_SUITE_ID":"suite"}`,
		"null_document":    `null`,
		"null_binding":     `{"SSL_CERT_FILE":null}`,
		"numeric_binding":  `{"SSL_CERT_FILE":12}`,
		"boolean_binding":  `{"SSL_CERT_FILE":false}`,
		"object_binding":   `{"SSL_CERT_FILE":{}}`,
		"array":            `[]`,
		"string":           `"private-input"`,
		"unknown_binding":  `{"UNRELATED_SECRET":"private-input"}`,
		"trailing_value":   `{} {}`,
		"trailing_garbage": `{} private-input`,
		"incomplete":       `{"SSL_CERT_FILE":`,
	} {
		t.Run("recovery_environment_rejects_"+name, func(t *testing.T) {
			file := filepath.Join(t.TempDir(), "environment.json")
			if err := os.WriteFile(file, []byte(data), 0600); err != nil {
				t.Fatal(err)
			}
			if _, err := readRecoveryEnvironment(file); err == nil {
				t.Fatal("ambiguous or malformed recovery environment accepted")
			}
			if retained, err := os.ReadFile(file); err != nil || string(retained) != data {
				t.Fatal("rejected authority was changed", err)
			}
		})
	}
}

func testOwnedLeaseWithoutStartupEnvironment(t *testing.T) {
	for _, variant := range []string{"owned", "failed_cleanup", "foreign_run", "foreign_root", "borrowed", "foreign_label", "short_identity", "pending", "missing_lease", "malformed_environment", "duplicate_environment", "unsafe_environment"} {
		t.Run("lease_without_startup_environment_"+variant, func(t *testing.T) {
			runtime, root, run := runtimeFixture(t)
			file := filepath.Join(runtime.Root, "suite", "suite-lease.json")
			lease := leaseproof.Lease{
				SchemaID: "cartulary.test_services.lease.v3", LeaseID: "original-lease", SuiteID: "original-suite",
				RunID: runtime.RunID, RunRoot: run, ResultRoot: filepath.Dir(run), Target: "check",
				Mode: "owned", OwnershipMode: "owned", OwnerPID: os.Getpid(), CreatedAt: time.Now().UTC().Format(time.RFC3339Nano),
				CleanupState: "not_started", ProofLabels: map[string]string{}, ProofPrefixes: map[string]string{},
				Resources: []leaseproof.LeaseResource{{Kind: "container", Service: "postgres", ContainerID: strings.Repeat("a", 64), Image: "fixture", Labels: map[string]string{
					leaseproof.LabelManaged: leaseproof.ManagedValue, leaseproof.LabelSuiteID: "original-suite",
					leaseproof.LabelRunID: runtime.RunID, leaseproof.LabelTarget: "check", leaseproof.LabelService: "postgres",
				}}},
			}
			state := "acquired"
			switch variant {
			case "foreign_run":
				lease.RunID = "foreign"
			case "foreign_root":
				lease.RunRoot = filepath.Join(root, "foreign")
			case "borrowed":
				lease.OwnershipMode = "attach"
			case "foreign_label":
				lease.Resources[0].Labels[leaseproof.LabelSuiteID] = "foreign"
			case "short_identity":
				lease.Resources[0].ContainerID = "abcdef"
			case "pending":
				state = "pending"
			}
			if variant != "missing_lease" {
				writeFixtureJSON(t, file, lease)
			}
			envFile := filepath.Join(filepath.Dir(file), "suite-environment.json")
			if variant == "malformed_environment" {
				writeFixtureJSON(t, envFile, []string{"invalid"})
			} else if variant == "duplicate_environment" {
				if err := os.WriteFile(envFile, []byte(`{"CARTULARY_TEST_SUITE_ID":"foreign","CARTULARY_TEST_SUITE_ID":"original-suite"}`), 0600); err != nil {
					t.Fatal(err)
				}
			} else if variant == "unsafe_environment" {
				if err := os.Symlink(file, envFile); err != nil {
					t.Fatal(err)
				}
			}
			if err := runtime.RecordDependentResource("managed_suite", file, state, []string{}, envFile); err != nil {
				t.Fatal(err)
			}
			defer runtime.RecordResource("managed_suite", file, "released")
			stops := 0
			p := ProductionOptions{Root: root, RunRoot: run, Runtime: runtime, Environment: map[string]string{"TEST_SERVICES_BIN": "services"}, Run: func(_ context.Context, _ string, args []string, env map[string]string) (CommandResult, error) {
				stops++
				if args[0] != "terminate-suite" || !strings.HasPrefix(args[2], file+".cleanup-") || env[suiteservices.SuiteIDEnv] != lease.SuiteID {
					t.Fatal("cleanup did not use original lease")
				}
				if _, err := os.Lstat(envFile); !errors.Is(err, os.ErrNotExist) {
					t.Fatal("startup environment was recreated")
				}
				if variant == "failed_cleanup" {
					return CommandResult{}, errors.New("physical cleanup failed")
				}
				return CommandResult{}, nil
			}}
			if _, err := runtime.BeginRecoveryAttempt(); err != nil {
				t.Fatal(err)
			}
			defer runtime.FinishRecoveryAttempt()
			for range 2 {
				err := p.Recover(context.Background())
				if (err == nil) != (variant == "owned") {
					t.Fatalf("recovery outcome: %v", err)
				}
			}
			wantStops := 0
			if variant == "owned" || variant == "failed_cleanup" {
				wantStops = 1
			}
			if stops != wantStops {
				t.Fatalf("cleanup calls=%d want=%d", stops, wantStops)
			}
			resources, err := runtime.Resources(4096)
			if err != nil || (len(resources) == 0) != (variant == "owned") {
				t.Fatal("lost recovery authority", err, len(resources))
			}
		})
	}
}
