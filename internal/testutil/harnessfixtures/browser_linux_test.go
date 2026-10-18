//go:build linux

package harnessfixtures

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/harnesslocal"
	"github.com/JochiRaider/cartulary/internal/testutil/harnesssession"
)

func TestFixtureBrowserAdapter(t *testing.T) {
	node := os.Getenv("NODE_BIN")
	if node == "" {
		t.Fatal("Make-owned Node binding required")
	}
	cwd, _ := os.Getwd()
	root := filepath.Clean(filepath.Join(cwd, "../../.."))
	t.Run("closed_cleanup_detail_survives_browser_failure_without_reclassification", func(t *testing.T) {
		primary := errors.New("private transport sentinel")
		for index, data := range []string{
			`{"schema_id":"cartulary.harness_cleanup_diagnostic.v1","stage":"process_authority","cause":"unresolved","errno":0}`,
			`{"schema_id":"cartulary.harness_cleanup_diagnostic.v1","stage":"process_authority","cause":"unresolved","errno":0,"pid":123}`,
		} {
			p := ProductionOptions{Environment: map[string]string{"NODE_BIN": node}, Run: func(context.Context, string, []string, map[string]string) (CommandResult, error) {
				return CommandResult{Exit: 12, Stdout: []byte(`{"failure_class":"harness","failure_reason":"cleanup_error"}`), Stderr: []byte(data)}, primary
			}}
			_, err := p.browserOperation(context.Background(), "settle", nil, nil)
			var failure *Failure
			if !errors.As(err, &failure) || failure.Class != "harness" || failure.Reason != "cleanup_error" || !errors.Is(err, primary) {
				t.Fatal("failure classification or native cause changed", err)
			}
			var diagnostic *harnesssession.CleanupDiagnostic
			got := errors.As(err, &diagnostic)
			want := index == 0
			if got != want {
				t.Fatal("invalid diagnostic crossed the boundary", got, want)
			}
			if got && harnesssession.DescribeCleanupFailure(err).Stage != "process_authority" {
				t.Fatal("native cause was discarded")
			}
		}
	})
	t.Run("closed_cli_rejects_unknown_operation", func(t *testing.T) {
		cmd := exec.Command(node, filepath.Join(root, "tools/harness/browser/fixture-lifecycle.mjs"), "unknown")
		cmd.Stderr = os.Stderr
		stdout, err := cmd.Output()
		if err == nil {
			t.Fatal("unknown operation accepted")
		}
		var failure map[string]any
		if json.Unmarshal(stdout, &failure) != nil || len(failure) != 2 || failure["failure_class"] == nil {
			t.Fatalf("missing closed failure: %s", stdout)
		}
	})
	t.Run("private_browser_preparation_joins_native_runtime", func(t *testing.T) {
		run := filepath.Join(t.TempDir(), "native-browser-prepare")
		if err := os.Mkdir(run, 0700); err != nil {
			t.Fatal(err)
		}
		runtime, err := harnesssession.NewRuntime(root, run, "native-browser-prepare", t.TempDir())
		if err != nil {
			t.Fatal(err)
		}
		defer runtime.Close()
		suite := filepath.Join(runtime.Root, "test-services", "suite-lease.json")
		writeFixtureJSON(t, suite, map[string]any{"schema_id": "cartulary.test_services.lease.v1", "run_id": runtime.RunID, "suite_id": "fixture-suite", "lease_id": runtime.LeaseID, "result_root": filepath.Dir(run), "run_root": filepath.Join(filepath.Dir(run), runtime.RunID), "target": "browser-e2e", "mode": "owned", "ownership_mode": "owned", "owner_pid": os.Getpid(), "created_at": time.Now().UTC().Format(time.RFC3339Nano), "resources": []any{}, "proof_labels": map[string]any{}, "proof_prefixes": map[string]any{}, "cleanup_state": "not_started"})
		sessionRoot, _ := runtime.RecoveryPath("browser-stack-leases", "attempt")
		directory, err := harnesslocal.PrivateDirectory(sessionRoot)
		if err != nil {
			t.Fatal(err)
		}
		directory.Close()
		if err := runtime.RecordDependentResource("managed_suite", suite, "acquired", []string{}); err != nil {
			t.Fatal(err)
		}
		defer runtime.RecordResource("managed_suite", suite, "released")
		browserFile := filepath.Join(sessionRoot, "acquisition", "owner.json")
		if err := runtime.RecordDependentResource("browser_stack", browserFile, "pending", []string{fixtureReference(t, runtime, "managed_suite", suite)}); err != nil {
			t.Fatal(err)
		}
		defer runtime.RecordResource("browser_stack", browserFile, "released")
		p := ProductionOptions{Root: root, RunRoot: run, Target: "browser-e2e", Runtime: runtime, Environment: map[string]string{"NODE_BIN": node, "PATH": os.Getenv("PATH"), "CARTULARY_WEB_E2E_PRIVATE_SESSION_ROOT": sessionRoot}}
		native := p.command()
		p.Run = func(ctx context.Context, command string, _ []string, env map[string]string) (CommandResult, error) {
			// Call the physical module directly so a failed synthetic proof exposes
			// its stack in test diagnostics instead of the private CLI's closed error.
			module, _ := json.Marshal("file://" + filepath.Join(root, "tools/harness/browser/fixture-lifecycle.mjs"))
			args := []string{"--input-type=module", "--eval", "import {fixtureLifecycle} from " + string(module) + "; console.log(JSON.stringify(await fixtureLifecycle('prepare',process.argv.slice(1))));", suite}
			result, err := native(ctx, command, args, env)
			if err != nil {
				t.Logf("operation %s failed: %v; stderr=%s", args[1], err, result.Stderr)
			}
			return result, err
		}
		value, err := p.browserOperation(context.Background(), "prepare", []string{suite}, p.environment())
		if err != nil {
			t.Fatal(err)
		}
		if value["file"] != filepath.Join(sessionRoot, "acquisition", "owner.json") || len(value["launch_id"].(string)) != 36 {
			t.Fatal(value)
		}
	})
}
