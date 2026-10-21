//go:build linux

package harnessruntime

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/harnesslocal"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessschema"
	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices"
)

// The separate installed-service route keeps real Docker work out of the native
// unit contract. Both public output modes traverse Make and its real wrappers.
func TestRuntimeNestedServiceCancellation(t *testing.T) {
	// This fixture owns a processRunner outside Run. Hold the enclosing unit's
	// lifecycle until its own nested work settles, including an outer interrupt.
	parent, settle, err := joinNested(context.Background(), os.Getenv(nestedContextKey))
	if err != nil {
		t.Fatal("enclosing lifecycle unavailable")
	}
	clean := true
	defer func() { settle(clean) }()
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "../../.."))
	results, outer := os.Getenv("CARTULARY_TEST_RESULTS_DIR"), os.Getenv("CARTULARY_TEST_RUN_ID")
	if results == "" || outer == "" {
		t.Fatal("Make-owned retained identity required")
	}
	if !filepath.IsAbs(results) {
		results = filepath.Join(root, results)
	}
	results = filepath.Join(results, outer, "nested-cancellation")
	for _, mode := range []string{"summary", "machine", "summary-startup"} {
		if !t.Run(mode, func(t *testing.T) {
			env := map[string]string{}
			for _, entry := range os.Environ() {
				k, v, ok := strings.Cut(entry, "=")
				if !ok || strings.HasPrefix(k, "CARTULARY_HARNESS_") || strings.HasPrefix(k, "CARTULARY_TEST_SERVICES_") || strings.HasPrefix(k, "CARTULARY_PGTEST_") || strings.HasPrefix(k, "CARTULARY_S3TEST_") {
					continue
				}
				env[k] = v
			}
			for _, k := range []string{"MAKEFLAGS", "MAKEOVERRIDES", "MFLAGS", "JSON", "OWNER", "ROWS", "SCOPE", "SUITE", "SERVICE_BACKED_ONLY", "CARTULARY_MAKE_INPUT_SOURCES"} {
				delete(env, k)
			}
			env["CARTULARY_TEST_RESULTS_DIR"], env["CARTULARY_TEST_RUN_ID"] = results, mode
			env["CARTULARY_OUTPUT_MODE"], env["CARTULARY_HARNESS_CACHE_MODE"] = strings.TrimSuffix(mode, "-startup"), "off"
			run := filepath.Join(results, mode)
			p := processRunner{Root: root, Environment: env}
			u := engineUnit("test:nested-service", 1)
			u.TimeoutMS = 300000
			u.Command = command{Executable: "make", Args: []string{"--no-print-directory", "test-slice", "OWNER=module.collaboration", "ROWS=module.collaboration.store.presence_payloads_are_incident_scoped_and_epheme_185de72cf2,module.collaboration.store.strict_frame_and_failure_contract_a3b4c5d6e7,module.collaboration.store.the_first_application_message_on_the_incident_so_11b91f6bd8", "HARNESS_DIAGNOSTICS=basic"}, Environment: map[string]string{}}
			ctx, cancel := context.WithCancel(parent)
			done := make(chan struct{})
			var outcome unitResult
			clean = false
			go func() { outcome = p.Execute(ctx, executionRequest{Unit: u}); close(done) }()
			defer func() { cancel(); <-done; clean = len(outcome.Cleanup) == 0 }()
			// Observe published state, never private event staging. Dispatch of the
			// service-backed Go unit establishes that its fixture acquisition ended.
			deadline := time.After(3 * time.Minute)
			acquired := false
			for !acquired {
				select {
				case <-done:
					t.Fatalf("nested command ended before cancellation: %+v stderr=%.4096s retained=%s", outcome, outcome.Stderr, run)
				case <-deadline:
					cancel()
					<-done
					t.Fatal("nested service did not dispatch")
				case <-time.After(20 * time.Millisecond):
					data, err := harnesslocal.ReadFile(filepath.Join(run, "diagnostics/live-status.json"), 16*1024*1024, false)
					if err != nil {
						continue
					}
					v, err := harnessschema.Decode(data)
					if err != nil || harnessschema.Validate("cartulary.harness_live_snapshot.v3", v) != nil {
						t.Fatal("invalid live publication")
					}
					view := v.(map[string]any)
					if view["run_id"] != mode {
						t.Fatal("foreign live publication")
					}
					for _, raw := range view["units"].([]any) {
						unit := raw.(map[string]any)
						if mode == "summary-startup" && strings.HasPrefix(unit["unit_id"].(string), "go:") && unit["activity"] == "preparing" {
							// Owned-start is published after startup registration. This
							// observes the public scope, never private journal staging.
							files, _ := filepath.Glob(filepath.Join(run, "_shared/test-services/*/service-scope.json"))
							for _, file := range files {
								data, e := harnesslocal.ReadFile(file, 1048576, false)
								var scope suiteservices.ServiceScope
								if e == nil && json.Unmarshal(data, &scope) == nil && scope.RunID == mode && scope.Wrapper.OwnedCount == 1 {
									acquired = true
								}
							}
						} else if mode != "summary-startup" && strings.HasPrefix(unit["unit_id"].(string), "go:") && unit["activity"] == "executing" {
							acquired = true
						}
					}
				}
			}
			cancel()
			<-done
			r := outcome
			if r.Status != "cancelled" || len(r.Cleanup) != 0 {
				t.Fatalf("nested settlement failed: %+v stderr=%.4096s retained=%s", r, r.Stderr, run)
			}
			read := func(name string, out any) {
				data, err := harnesslocal.ReadFile(filepath.Join(run, name), 16*1024*1024, false)
				if err != nil || json.Unmarshal(data, out) != nil {
					t.Fatalf("missing terminal %s: %v", name, err)
				}
			}
			var summary struct {
				Status string `json:"status"`
			}
			read("run-summary.json", &summary)
			if summary.Status != "cancelled" {
				t.Fatalf("child status %s", summary.Status)
			}
			var cleanup struct {
				Results []struct {
					Outcome string `json:"outcome"`
				} `json:"results"`
			}
			read("cleanup-results.json", &cleanup)
			if len(cleanup.Results) == 0 {
				t.Fatal("no child cleanup receipts")
			}
			for _, result := range cleanup.Results {
				if result.Outcome != "completed" {
					t.Fatal("incomplete child cleanup")
				}
			}
			read("retained-secret-scan.json", &summary)
			if summary.Status != "pass" {
				t.Fatal("child scan did not pass")
			}
			files, err := filepath.Glob(filepath.Join(run, "_shared/test-services/*/service-scope.json"))
			if err != nil || len(files) == 0 {
				t.Fatal("no acquired service evidence")
			}
			for _, file := range files {
				var scope suiteservices.ServiceScope
				name, _ := filepath.Rel(run, file)
				read(name, &scope)
				settled := scope.Cleanup.Status == "succeeded"
				if mode == "summary-startup" {
					settled = settled || scope.Cleanup.Status == "startup_failed" && scope.Cleanup.PhysicalSettlement == "completed"
				}
				if scope.RunID != mode || scope.Wrapper.OwnedCount != 1 || !settled {
					t.Fatal("owned service did not settle")
				}
			}
			t.Logf("nested %s cancellation settled with terminal evidence: %s", mode, run)
		}) {
			return
		}
	}
}
