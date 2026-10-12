//go:build linux

package harnessfixtures

import (
	"context"
	"encoding/json"
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
		sessionRoot, _ := runtime.PrivatePath("browser-stack-leases", "attempt")
		directory, err := harnesslocal.PrivateDirectory(sessionRoot)
		if err != nil {
			t.Fatal(err)
		}
		directory.Close()
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
