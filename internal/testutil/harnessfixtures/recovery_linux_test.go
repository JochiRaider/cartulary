//go:build linux

package harnessfixtures

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/testutil/harnesslocal"
)

func TestFixtureRecovery(t *testing.T) {
	for _, variant := range []string{"success", "missing_dependency", "unknown_dependency", "failed_browser", "borrowed", "borrowed_browser_failure"} {
		t.Run(variant, func(t *testing.T) {
			runtime, root, run := runtimeFixture(t)
			suite := filepath.Join(runtime.Root, "test-services", "suite-lease.json")
			browser := filepath.Join(runtime.Root, "browser-stack-leases", "a", "acquisition", "owner.json")
			writeFixtureJSON(t, suite, map[string]string{"proof": "synthetic"})
			writeFixtureJSON(t, browser, map[string]string{"proof": "synthetic"})
			if variant != "missing_dependency" && !strings.HasPrefix(variant, "borrowed") {
				if err := runtime.RecordResource("managed_suite", suite, "acquired"); err != nil {
					t.Fatal(err)
				}
			}
			if variant != "borrowed" {
				if err := runtime.RecordResource("browser_stack", browser, "pending"); err != nil {
					t.Fatal(err)
				}
			}
			if strings.HasPrefix(variant, "borrowed") {
				file := filepath.Join(runtime.Root, "borrowed-suite.json")
				writeFixtureJSON(t, file, map[string]string{"proof": "synthetic"})
				if err := runtime.RecordResource("borrowed_suite", file, "acquired"); err != nil {
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
					return CommandResult{Started: true, Exit: 12}, errors.New("physical stop failed")
				}
				return CommandResult{Started: true, Stdout: []byte(`{}`)}, nil
			}
			err := p.Recover(context.Background())
			joined := strings.Join(calls, ",")
			if variant == "success" {
				if err != nil || joined != "inspect,recover,suite" {
					t.Fatal(calls, err)
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
